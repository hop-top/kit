package secret

import "fmt"

// Config describes which secret backend to use.
type Config struct {
	Backend string // "env", "file", "keyring", "openbao", "infisical", "memory"
	Prefix  string // for env adapter
	Dir     string // for file adapter
	Service string // for keyring adapter
	Addr    string // for openbao/infisical
	Token   string // for openbao/infisical
	Mount   string // for openbao
	Project string // for infisical
	Env     string // for infisical
}

// Opener is a function that creates a MutableStore from config.
// Registered via RegisterBackend.
type Opener func(cfg Config) (MutableStore, error)

var backends = map[string]Opener{}

// RegisterBackend registers a factory for the named backend.
func RegisterBackend(name string, fn Opener) {
	backends[name] = fn
}

// Open creates a MutableStore from config using registered backends.
func Open(cfg Config) (MutableStore, error) {
	fn, ok := backends[cfg.Backend]
	if !ok {
		return nil, fmt.Errorf("secret: unknown backend %q", cfg.Backend)
	}
	return fn(cfg)
}
