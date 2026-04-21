# Troubleshooting {{app_name}}

## Common Issues

### Command not found

Ensure {{app_name}} is installed and on your PATH:

```sh
which {{app_name}}
```

If missing, reinstall (see [install](install.md)).

### Permission denied

Check file permissions. On Unix systems:

```sh
chmod +x $(which {{app_name}})
```

### Config not loading

Verify config file location and syntax:

```sh
{{app_name}} --verbose <command>
```

This prints which config files were loaded.

### Unexpected output format

Set the output format explicitly:

```sh
{{app_name}} --output json <command>
```

## Debug Mode

Run with verbose output for detailed diagnostics:

```sh
{{app_name}} --verbose <command>
```

## Getting Help

1. Check this troubleshooting guide
2. Search existing issues on GitHub
3. Open a new issue with:
   - {{app_name}} version (`{{app_name}} --version`)
   - OS and architecture
   - Steps to reproduce
   - Expected vs actual behavior
