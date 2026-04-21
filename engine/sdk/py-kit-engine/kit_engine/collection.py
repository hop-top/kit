"""CRUD operations on a typed collection."""

from __future__ import annotations

import requests


class Collection:
    """REST client for a single kit collection type."""

    def __init__(self, base_url: str, type_name: str):
        self._url = f"{base_url}/{type_name}"

    def create(self, data: dict) -> dict:
        resp = requests.post(f"{self._url}/", json=data)
        resp.raise_for_status()
        return resp.json()

    def get(self, id: str) -> dict:
        resp = requests.get(f"{self._url}/{id}")
        resp.raise_for_status()
        return resp.json()

    def list(
        self, *, limit: int = 20, offset: int = 0, sort: str = "", search: str = ""
    ) -> list[dict]:
        params: dict = {"limit": limit, "offset": offset}
        if sort:
            params["sort"] = sort
        if search:
            params["search"] = search
        resp = requests.get(f"{self._url}/", params=params)
        resp.raise_for_status()
        return resp.json()

    def update(self, id: str, data: dict) -> dict:
        resp = requests.put(f"{self._url}/{id}", json=data)
        resp.raise_for_status()
        return resp.json()

    def delete(self, id: str) -> None:
        resp = requests.delete(f"{self._url}/{id}")
        resp.raise_for_status()

    def history(self, id: str) -> list[dict]:
        resp = requests.get(f"{self._url}/{id}/history")
        resp.raise_for_status()
        return resp.json()

    def revert(self, id: str, version_id: str) -> dict:
        resp = requests.post(f"{self._url}/{id}/revert/{version_id}")
        resp.raise_for_status()
        return resp.json()
