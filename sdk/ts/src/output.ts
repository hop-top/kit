/**
 * @module output
 * @package @hop-top/kit
 *
 * Output rendering — writes structured values to a writable stream in one of
 * three formats: table, JSON, or YAML.
 *
 * ## Quick start
 *
 * ```ts
 * import { render, TABLE_FORMAT, JSON_FORMAT, YAML_FORMAT } from '@hop-top/kit/output';
 *
 * // Write a table to stdout
 * render(process.stdout, TABLE_FORMAT, [{ id: '1', name: 'Alice' }]);
 *
 * // Write indented JSON
 * render(process.stdout, JSON_FORMAT, { ok: true });
 *
 * // Write YAML
 * render(process.stdout, YAML_FORMAT, { version: 1 });
 * ```
 *
 * ## Table column alignment algorithm
 *
 * 1. Derive headers from `Object.keys(rows[0])` (or `Object.keys(obj)` for a
 *    single object).
 * 2. For each column compute `maxWidth = max(header.length, max(cell.length))`.
 * 3. Every cell and header is padded with trailing spaces to `maxWidth`.
 * 4. Columns are separated by two spaces (`"  "`).
 * 5. The resulting lines are all the same width, making them easy to scan.
 *
 * An empty array produces **no output** — not even a header row.  Callers
 * that need a "no results" message should check length before calling `render`.
 */

import * as yaml from 'js-yaml';

/**
 * Supported output formats.
 *
 * - `"table"` — aligned ASCII table; headers derived from object keys.
 * - `"json"`  — `JSON.stringify` with 2-space indentation + trailing newline.
 * - `"yaml"`  — YAML serialisation via js-yaml.
 */
export type Format = 'table' | 'json' | 'yaml';

/** Constant for the JSON output format. */
export const JSON_FORMAT = 'json' as const;

/** Constant for the YAML output format. */
export const YAML_FORMAT = 'yaml' as const;

/** Constant for the table output format. */
export const TABLE_FORMAT = 'table' as const;

/**
 * Renders `v` to `w` in the requested `format`.
 *
 * @param w       - Destination writable stream (e.g. `process.stdout`).
 * @param format  - One of {@link JSON_FORMAT}, {@link YAML_FORMAT}, {@link TABLE_FORMAT}.
 * @param v       - Any JSON/YAML-serialisable value.  For table format, pass
 *                  an array of plain objects or a single plain object.
 *
 * @throws {Error} For unknown format strings.
 */
export function render(w: NodeJS.WritableStream, format: Format, v: unknown): void {
  switch (format) {
    case JSON_FORMAT:
      w.write(JSON.stringify(v, null, 2) + '\n');
      break;
    case YAML_FORMAT:
      w.write(yaml.dump(v));
      break;
    case TABLE_FORMAT:
      renderTable(w, v);
      break;
    default: {
      // exhaustive check — TypeScript narrows to `never` here
      const _exhaustive: never = format;
      throw new Error(
        `unknown output format "${_exhaustive}" (valid: json, yaml, table)`,
      );
    }
  }
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

/**
 * Renders `v` as an aligned ASCII table.
 *
 * Accepts:
 *  - An array of plain objects → header row + one row per element.
 *  - A single plain object    → header row + one data row.
 *  - An empty array           → no output.
 */
function renderTable(w: NodeJS.WritableStream, v: unknown): void {
  const rows = normaliseRows(v);
  if (rows.length === 0) return;

  const headers = Object.keys(rows[0]);
  const cells: string[][] = rows.map(row =>
    headers.map(h => String((row as Record<string, unknown>)[h] ?? '')),
  );

  // Compute max width per column (header vs all cell values).
  const widths = headers.map((h, ci) =>
    Math.max(h.length, ...cells.map(row => row[ci].length)),
  );

  const pad = (s: string, w: number) => s + ' '.repeat(w - s.length);
  const formatRow = (cols: string[]) =>
    cols.map((c, i) => pad(c, widths[i])).join('  ');

  w.write(formatRow(headers) + '\n');
  for (const row of cells) {
    w.write(formatRow(row) + '\n');
  }
}

/**
 * Normalises `v` into an array of plain objects.
 *
 * - Array input → returned as-is (may be empty).
 * - Single object input → wrapped in a 1-element array.
 */
function normaliseRows(v: unknown): Record<string, unknown>[] {
  if (Array.isArray(v)) {
    return v as Record<string, unknown>[];
  }
  return [v as Record<string, unknown>];
}
