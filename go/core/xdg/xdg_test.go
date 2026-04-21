package xdg_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/kit/go/core/xdg"
)

func TestConfigDir_XDGOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-cfg")
	dir, err := xdg.ConfigDir("mytool")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/xdg-cfg/mytool", dir)
}

func TestDataDir_XDGOverride(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")
	dir, err := xdg.DataDir("mytool")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/xdg-data/mytool", dir)
}

func TestCacheDir_XDGOverride(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
	dir, err := xdg.CacheDir("mytool")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/xdg-cache/mytool", dir)
}

func TestStateDir_XDGOverride(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/xdg-state")
	dir, err := xdg.StateDir("mytool")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/xdg-state/mytool", dir)
}

func TestConfigDir_FallbackContainsTool(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")
	dir, err := xdg.ConfigDir("mytool")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(dir, filepath.Join("mytool")),
		"expected path to end with tool name, got: %s", dir)
}
