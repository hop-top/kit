# Installing {{app_name}}

## From Source

```sh
git clone {{module_prefix}}/{{app_name}}
cd {{app_name}}
task build
```

## Go

```sh
go install {{module_prefix}}/{{app_name}}@latest
```

## npm

```sh
npm install -g {{app_name}}
```

## pip

```sh
pip install {{app_name}}
```

## Homebrew

```sh
brew install {{app_name}}
```

## Verify Installation

```sh
{{app_name}} --version
```

## System Requirements

- Go 1.22+ (for Go builds)
- Node.js 20+ (for TypeScript builds)
- Python 3.11+ (for Python builds)
