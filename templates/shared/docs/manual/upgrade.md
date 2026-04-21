# Upgrading {{app_name}}

## Check Current Version

```sh
{{app_name}} --version
```

## Upgrade Methods

### Go

```sh
go install {{module_prefix}}/{{app_name}}@latest
```

### npm

```sh
npm update -g {{app_name}}
```

### pip

```sh
pip install --upgrade {{app_name}}
```

### Homebrew

```sh
brew upgrade {{app_name}}
```

## Breaking Changes

Check the CHANGELOG for breaking changes between
versions before upgrading.

## Rollback

If an upgrade causes issues, install the previous
version explicitly:

```sh
# Go
go install {{module_prefix}}/{{app_name}}@v<PREVIOUS_VERSION>

# npm
npm install -g {{app_name}}@<PREVIOUS_VERSION>

# pip
pip install {{app_name}}==<PREVIOUS_VERSION>
```
