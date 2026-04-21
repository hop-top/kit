#!/usr/bin/env bash
# post-create.sh — runs after devcontainer create
set -euo pipefail

echo "==> Installing TS dependencies (pnpm)"
if [ -d ts ]; then
  (cd ts && pnpm install)
fi

echo "==> Creating Python venv"
if [ -d py ]; then
  python3 -m venv sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/.venv
  sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/.venv/bin/pip install --upgrade pip
  if [ -f sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/pyproject.toml ]; then
    if ! sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/.venv/bin/pip install -e "py[dev]"; then
      echo "==> Editable install with dev extras failed; falling back"
      sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/py/.venv/bin/pip install -e py
    fi
  fi
fi

echo "==> Checking for AI tools"
if [ -f .ai-tools ] \
    && [ -f .devcontainer/install-ai-tools.sh ]; then
  bash .devcontainer/install-ai-tools.sh --auto
fi

echo "==> Dev container ready"
