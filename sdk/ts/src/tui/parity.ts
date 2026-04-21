/**
 * @module tui/parity
 * @package @hop-top/kit
 *
 * Loads tui/parity.json — the cross-language parity constants SoT.
 * All TUI modules should import constants from here, not hardcode them.
 */

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve, dirname } from 'node:path';

interface ParityData {
  status: {
    symbols: Record<'info' | 'success' | 'error' | 'warn', string>;
  };
  spinner: {
    frames: string[];
    interval_ms: number;
  };
  anim: {
    runes: string;
    interval_ms: number;
    default_width: number;
  };
  help: {
    /** Fang-vocabulary section names in render order. */
    section_order: string[];
    /** Display metadata keyed by fang section name. */
    sections: Record<string, { title: string }>;
  };
}

// Resolve path: sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/sdk/ts/src/tui/ → 3 levels up → repo root → contracts/parity/parity.json
const _dir  = dirname(fileURLToPath(import.meta.url));
const _path = resolve(_dir, '..', '..', '..', '..', 'contracts', 'parity', 'parity.json');

export const parity: ParityData = JSON.parse(readFileSync(_path, 'utf8'));
