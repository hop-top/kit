"""Peer discovery client."""

from __future__ import annotations

import requests


class PeerClient:
    """REST client for kit peer endpoints."""

    def __init__(self, base_url: str):
        self._url = f"{base_url}/peers"

    def list(self) -> list[dict]:
        resp = requests.get(self._url)
        resp.raise_for_status()
        return resp.json()

    def connect(self, address: str) -> dict:
        resp = requests.post(self._url, json={"address": address})
        resp.raise_for_status()
        return resp.json()
