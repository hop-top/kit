// Package config provides a layered YAML configuration loader.
//
// Layers are merged in the order: system → user → project → env, where each
// later layer overwrites fields set by earlier layers. Missing files are silently
// skipped; a file that exists but cannot be parsed returns an error.
//
// dst must be a pointer to a struct. YAML unmarshalling follows the rules of
// gopkg.in/yaml.v3: struct fields are matched by their yaml tag, or by
// lower-cased field name if no tag is present.
//
// Warning: YAML unmarshal replaces slice and array fields entirely rather than
// appending to them. Each layer overwrites the previous value for those types.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Options configures the sources for Load.
// Any Path field may be empty; empty paths are skipped entirely.
// EnvOverride, if non-nil, is called after all files have been merged and
// receives the fully-merged dst value so callers can apply environment-variable
// overrides last.
type Options struct {
	// SystemConfigPath is the path to the system-wide config file (e.g. /etc/tool/config.yaml).
	SystemConfigPath string
	// UserConfigPath is the path to the per-user config file (e.g. ~/.config/tool/config.yaml).
	UserConfigPath string
	// ProjectConfigPath is the path to the project-level config file (e.g. .tool.yaml).
	ProjectConfigPath string
	// EnvOverride, if non-nil, is called last with the merged config so callers
	// can layer environment-variable overrides on top of file values.
	EnvOverride func(cfg any)
}

// Load merges configuration into dst from the configured file paths, then
// applies EnvOverride if set.
//
// dst must be a non-nil pointer. Files that do not exist are silently ignored.
// A file that exists but is not valid YAML causes Load to return an error
// wrapping the parse failure and the offending path.
// The merge order is: SystemConfigPath → UserConfigPath → ProjectConfigPath →
// EnvOverride.
func Load(dst any, opts Options) error {
	for _, path := range []string{
		opts.SystemConfigPath, opts.UserConfigPath, opts.ProjectConfigPath,
	} {
		if path == "" {
			continue
		}
		if err := mergeFile(dst, path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("load config %s: %w", path, err)
		}
	}
	if opts.EnvOverride != nil {
		opts.EnvOverride(dst)
	}
	return nil
}

func mergeFile(dst any, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, dst)
}
