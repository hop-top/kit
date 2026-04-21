"""Sync operations client."""

from __future__ import annotations

import requests


class SyncClient:
    """REST client for kit sync endpoints."""

    def __init__(self, base_url: str):
        self._url = f"{base_url}/sync"

    def status(self) -> dict:
        resp = requests.get(f"{self._url}/status")
        resp.raise_for_status()
        return resp.json()

    def trigger(self) -> dict:
        resp = requests.post(f"{self._url}/trigger")
        resp.raise_for_status()
        return resp.json()
