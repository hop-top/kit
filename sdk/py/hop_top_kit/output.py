"""
hop_top_kit.output — table/json/yaml renderer.

Mirrors the Go output.Render API from hop.top/kit/go/console/output/renderer.go.

Public surface
--------------
Format  — Literal type alias: 'table' | 'json' | 'yaml'
render  — write *v* to *w* in the requested *format*

Table alignment algorithm
-------------------------
1. Collect all rows as lists of str values.
2. Compute ``col_widths[i] = max(len(header[i]), max(len(row[i]) for row))``
   for each column index *i*.
3. Pad every cell to its column width with trailing spaces.
4. Join cells in each row with a 2-space gap ('  ').
5. Write header row, then each data row, each terminated with '\\n'.
6. Empty input list → no output (not even a header).

Usage
-----
>>> import io
>>> from hop_top_kit.output import render
>>> buf = io.StringIO()
>>> render(buf, 'json', {'key': 'val'})
>>> print(buf.getvalue())
{
  "key": "val"
}
"""

from __future__ import annotations

import dataclasses
import json
from typing import IO, Any, Literal

import yaml

Format = Literal["table", "json", "yaml"]

_GAP = "  "  # 2-space column separator


# ---------------------------------------------------------------------------
# Public
# ---------------------------------------------------------------------------


def render(w: IO[str], format: Format, v: Any) -> None:
    """Write *v* to *w* in *format*.

    Parameters
    ----------
    w:
        Any writable text stream (e.g. ``sys.stdout``, ``io.StringIO``).
    format:
        One of ``'json'``, ``'yaml'``, or ``'table'``.
    v:
        Value to render.  For ``'table'`` this must be a list of dicts,
        a list of dataclasses, a single dict, or a single dataclass.
        For ``'json'`` / ``'yaml'`` any JSON/YAML-serialisable value works.

    Raises
    ------
    ValueError
        If *format* is not one of the three recognised values.
    """
    if format == "json":
        w.write(json.dumps(v, indent=2))
        w.write("\n")
    elif format == "yaml":
        w.write(yaml.dump(v))
    elif format == "table":
        _render_table(w, v)
    else:
        raise ValueError(f"unknown output format {format!r} (valid: json, yaml, table)")


# ---------------------------------------------------------------------------
# Table helpers
# ---------------------------------------------------------------------------


def _to_rows(v: Any) -> tuple[list[str], list[list[str]]]:
    """Return ``(headers, data_rows)`` from *v*.

    Handles:
    - list of dicts
    - list of dataclasses
    - single dict
    - single dataclass

    An empty list returns ``([], [])``.
    """
    if isinstance(v, list):
        if not v:
            return [], []
        first = v[0]
        if dataclasses.is_dataclass(first) and not isinstance(first, type):
            headers = [f.name for f in dataclasses.fields(first)]
            rows = [[str(getattr(item, h)) for h in headers] for item in v]
        else:
            # list of dicts (or dict-like)
            headers = list(first.keys())
            rows = [[str(item[h]) for h in headers] for item in v]
        return headers, rows

    if dataclasses.is_dataclass(v) and not isinstance(v, type):
        headers = [f.name for f in dataclasses.fields(v)]
        return headers, [[str(getattr(v, h)) for h in headers]]

    if isinstance(v, dict):
        headers = list(v.keys())
        return headers, [[str(v[h]) for h in headers]]

    raise TypeError(f"render table: unsupported type {type(v)!r}")


def _render_table(w: IO[str], v: Any) -> None:
    """Format and write a table with aligned, 2-space-gapped columns."""
    headers, rows = _to_rows(v)
    if not headers:
        return  # empty list — no output

    # Compute per-column max widths (header participates too).
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if len(cell) > col_widths[i]:
                col_widths[i] = len(cell)

    def _fmt_row(cells: list[str]) -> str:
        padded = [cell.ljust(col_widths[i]) for i, cell in enumerate(cells)]
        return _GAP.join(padded).rstrip()

    w.write(_fmt_row(headers) + "\n")
    for row in rows:
        w.write(_fmt_row(row) + "\n")
