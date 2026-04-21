"""Identity management client."""

from __future__ import annotations

import requests


class IdentityClient:
    """REST client for kit identity endpoints."""

    def __init__(self, base_url: str):
        self._url = f"{base_url}/identity"

    def whoami(self) -> dict:
        resp = requests.get(self._url)
        resp.raise_for_status()
        return resp.json()

    def set(self, name: str, email: str = "") -> dict:
        resp = requests.put(self._url, json={"name": name, "email": email})
        resp.raise_for_status()
        return resp.json()
