# XDG Cache Resolution — AIM

Cache path: `{xdg_cache}/hop/aim/`

## Resolution order

1. `$XDG_CACHE_HOME` env var (if set + non-empty)
2. Platform default (see below)
3. Fallback: `~/.cache`

## Platform defaults

| Platform | Default path        |
|----------|---------------------|
| macOS    | ~/Library/Caches    |
| Windows  | %LOCALAPPDATA%      |
| Linux    | ~/.cache            |

## Algorithm (pseudocode)

```
func cacheDir():
    if env("XDG_CACHE_HOME") != "":
        return env("XDG_CACHE_HOME") / "hop" / "aim"

    switch runtime.os:
        case "darwin":
            base = home() / "Library" / "Caches"
        case "windows":
            base = env("LOCALAPPDATA")
            if base == "":
                base = home() / ".cache"
        default:
            base = home() / ".cache"

    return base / "hop" / "aim"
```

## Notes

- All ports (Go, TS, Python, PHP) MUST use identical resolution
- Env var takes precedence on every platform
- Windows: `%LOCALAPPDATA%` typically `C:\Users\<u>\AppData\Local`
- Final path always ends with `hop/aim/` (forward slash, even Windows)
- Create directory recursively on first write (0755 / user-only)
