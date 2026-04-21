# CLI Demo Media — Reusable GHA Workflow

> Reusable `workflow_call` workflow that records terminal sessions + screenshots,
> commits media to a dedicated assets branch, and posts an embedded PR comment.

---

## Problem

- CLI projects lack a standard way to generate + surface demo GIFs/screenshots in PRs.
- Hand-rolled per-project scripts diverge; no sharing, no consistency.
- Media committed to feature branches bloats history and causes merge conflicts.

---

## Workflow Inputs

| Input | Default | Notes |
|---|---|---|
| `record-command` | `vhs` | Override to skip VHS install + use custom recorder |
| `tape-files` | `tapes/*.tape` | Shell glob; multiple patterns space-separated |
| `screenshot-command` | _(empty)_ | Leave empty to skip screenshot step entirely |
| `media-dir` | `media` | Relative to repo root; where output files land |
| `assets-branch` | `assets` | Long-lived branch; never merged to main |
| `build-command` | _(empty)_ | Runs before recording (e.g. `make build`) |
| `working-directory` | `.` | All commands run here |
| `pr-comment-title` | `## Demo Media` | First line of the posted comment |

**Secrets:** caller passes `secrets: inherit` or maps `token` explicitly.
Workflow uses `secrets.token || github.token` throughout.

---

## Step-by-Step Flow

```
1. checkout source (fetch-depth: 0 — needed for worktree ops)
2. resolve PR number (event.pull_request.number ∥ run_number)
3. prepare assets worktree
   └─ branch exists? → git worktree add
   └─ branch missing? → orphan branch + seed commit + push
4. build (if build-command set)
5a. install VHS (if record-command == "vhs" → charmbracelet/vhs-action@v2)
5b. record: expand tape-files glob → run record-command per tape
6. screenshots (if screenshot-command set)
7. commit media to assets branch
   └─ copy media-dir/ → worktree/<assets-branch>/<pr-num>/
   └─ skip commit if no changes (idempotent)
8. find existing PR comment (peter-evans/find-comment@v3, marker: <!-- cli-demo-media -->)
9. build comment body (raw.githubusercontent.com URLs, embed GIF/PNG, link mp4)
10. post/update comment (peter-evans/create-or-update-comment@v4, edit-mode: replace)
```

---

## Assets Branch Strategy

Files land at `<assets-branch>/<pr-number>/<filename>` so each PR has its
own isolated subdirectory — reruns overwrite without clobbering other PRs.

See: [assets-branch-strategy-v1.mmd](2026-04-18-cli-demo-media-workflow-design/assets-branch-strategy-v1.mmd)

```
assets (branch, orphan)
├── README.md
├── assets/
│   ├── 42/          ← PR #42
│   │   ├── demo.gif
│   │   └── help.png
│   └── 57/          ← PR #57
│       └── install.gif
```

Raw URL pattern:
```
https://raw.githubusercontent.com/<owner>/<repo>/assets/assets/<pr-num>/<file>
```

**Worktree isolation:** `git worktree add` checks out the assets branch into
`$RUNNER_TEMP/assets-worktree` — completely separate from the source checkout.
No cross-contamination; no stash gymnastics.

**[skip ci]** appended to asset commits to prevent infinite workflow loops.

---

## Caller Example — `spaced`

```yaml
# .github/workflows/demo-media.yml  (in the spaced repo)
name: Demo Media

on:
  pull_request:
    paths:
      - 'tapes/**'
      - 'cli/**'

jobs:
  media:
    uses: hop-top/kit/.github/workflows/cli-demo-media.yml@main
    with:
      build-command: make build
      tape-files: tapes/*.tape
      media-dir: media
      pr-comment-title: "## spaced — Demo"
    secrets: inherit
```

---

## Extension Points

### Custom recorder (asciinema)

```yaml
jobs:
  media:
    uses: hop-top/kit/.github/workflows/cli-demo-media.yml@main
    with:
      record-command: asciinema rec --overwrite
      tape-files: "demos/install.sh demos/sync.sh"
      build-command: |
        pip install asciinema
        make build
      media-dir: media
```

When `record-command != "vhs"`, the VHS install step is skipped entirely.
The caller is responsible for installing any dependencies in `build-command`.

### Playwright screenshots

```yaml
with:
  screenshot-command: npx playwright test --project=screenshots
  build-command: make build && npm ci
```

Output PNGs must land in `media-dir`; the build-comment step picks them up
automatically.

---

## GIF vs mp4

VHS can output `.gif`, `.mp4`, or `.svg`. Only `.gif`/`.png`/`.svg` embed
inline in GitHub markdown (`![alt](url)`). `.mp4` requires a download link.
Prefer `.gif` in tape files:

```
# demo.tape
Output demo.gif   # ← not demo.mp4
```

---

## Comment Update Strategy

Marker `<!-- cli-demo-media -->` is embedded in every comment body.
`peter-evans/find-comment` locates it by author + body substring.
`create-or-update-comment` with `edit-mode: replace` overwrites in-place —
one comment per PR regardless of how many times the workflow runs.

---

## Open Questions / Trade-offs

| Question | Decision | Rationale |
|---|---|---|
| Orphan branch vs. regular branch for assets | orphan | no source-code history; clean `git log` |
| Per-PR subdir vs. flat layout | per-PR subdir | reruns safe; no clobber across PRs |
| GIF vs. mp4 preference | GIF | embeds natively in GH markdown; no video player needed |
| Token scope | `secrets.token \|\| github.token` | caller can escalate if cross-repo push needed |
| VHS install on custom recorder | skipped | saves ~30s; caller handles own deps in build-command |
| mp4 in comment | linked, not embedded | GitHub markdown does not render `<video>` in comments |
| Large GIF size | out-of-scope for now | callers can set VHS `Width`/`Height`/`Framerate` in tapes |
