# claude-github

A drawer of experiments ("wires of assorted lengths"), each in its own folder
with its own `Taskfile.yml`. The root `index.html` is the landing page, and
also the index of every experiment. It renders `wires.json` as a drawer of
wires, one per experiment. The tooling is Go plus [Task](https://taskfile.dev),
with no other runtimes.

| wire | kind | what it does |
|---|---|---|
| [`apps/spillelista`](apps/spillelista) | static web app | Spotify playlist page |
| [`tools/skill-audit`](tools/skill-audit) | Go CLI | Audits Claude skills against current best practice and writes the prompt to fix them |

## Layout

```
index.html           landing page (GitHub Pages root), rendered from wires.json
assets/              shared static files for the landing page (background photo)
wires.json           index of every experiment; `task ci` fails if one is missing
apps/<name>/         web apps / sites, published as-is
tools/<name>/        CLIs, one Go module each; an index.html makes it a published page with downloads
infra/drawer/        repo plumbing (Go): checks wires.json, builds the site, local static server
templates/<kind>/    scaffolds used by `task new:*`
go.work              Go workspace listing every module
Taskfile.yml         root tasks + includes for every project (`includes:` stays last)
.github/workflows/   ci.yml runs `task ci`; pages.yml publishes the site from main
```

## Common tasks

```sh
task                       # list everything
task ci                    # drawer index check + `ci` in every project (what GitHub Actions runs)
task serve                 # the repo as-is at http://127.0.0.1:8000/
task site:serve            # build exactly what Pages publishes into _site/ and preview it
task skill-audit:audit     # run one project's task from the root
task new:go NAME=my-tool BLURB="what it does"
                           # scaffold tools/my-tool; registers it in go.work, the Taskfile
                           # and wires.json, then runs its ci
```

Requirements: Go 1.24+ and Task 3.x (`go install github.com/go-task/task/v3/cmd/task@latest`).

## Publishing

`.github/workflows/pages.yml` runs `task index:check site` on every push to
`main` and deploys `_site/` to GitHub Pages:

- the drawer (`index.html` + `wires.json`) at the root
- each `apps/*` wire with an `href`, as its whole folder
- each `tools/*` wire with an `href`: only its `index.html`, plus
  cross-compiled binaries and `SHA256SUMS` in `dl/` for Go tools

One-time setup: **Settings → Pages → Source: GitHub Actions**.

## Adding something that isn't a Go CLI

1. Create `apps/<name>/` or `tools/<name>/` with a `Taskfile.yml` that has a `ci` task.
2. Add it to `includes:` at the end of the root Taskfile.
3. Put it in the drawer: `task wire:add NAME=<name> PATH=apps/<name> KIND=app BLURB="what it does"`.
   Apps get an `href` automatically. For a tool with its own page, add
   `"href": "tools/<name>/"` and an `index.html`.

`task ci` runs every project that has a Taskfile, and fails if one is
missing from `wires.json` or links a page that doesn't exist.
