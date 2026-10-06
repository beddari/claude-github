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
| [`labs/acme-env`](labs/acme-env) | podman lab | Split DNS and an internal ACME CA: CoreDNS, etcd and step-ca issuing per-name certs over HTTP-01 |
| [`labs/talos-cluster`](labs/talos-cluster) | Talos lab | Talos Kubernetes in Docker that consumes acme-env through external-dns and cert-manager |

## Layout

```
index.html           landing page (GitHub Pages root), rendered from wires.json
assets/              shared static files for the landing page (background photo)
wires.json           index of every experiment; `task ci` fails if one is missing
apps/<name>/         web apps / sites, published as-is
tools/<name>/        CLIs, one Go module each; an index.html makes it a published page with downloads
labs/<name>/         runnable environments (containers, clusters); labs/lab.env holds shared settings
infra/drawer/        repo plumbing (Go): checks wires.json, builds the site, local static server
templates/<kind>/    scaffolds used by `task new:*`
go.work              Go workspace listing every module
Taskfile.yml         root tasks + includes for every project (`includes:` stays last)
Brewfile             host tools for macOS and Linuxbrew (`task tools`)
.github/workflows/   ci.yml runs `task ci`; pages.yml publishes the site from main
```

## Common tasks

```sh
task tools                 # brew bundle: go, task, podman, talosctl, kubectl, helm, step, dig, jq
task                       # list everything
task ci                    # drawer index check + `ci` in every project (what GitHub Actions runs)
task serve                 # the repo as-is at http://127.0.0.1:8000/
task site:serve            # build exactly what Pages publishes into _site/ and preview it
task skill-audit:audit     # run one project's task from the root
task new:go NAME=my-tool BLURB="what it does"
                           # scaffold tools/my-tool; registers it in go.work, the Taskfile
                           # and wires.json, then runs its ci
```

Requirements: [Homebrew](https://brew.sh) (macOS or Linux), then `brew bundle`
(or `task tools`). Go and Task alone are enough for everything except the labs,
which also need podman and a Docker Engine.

## Publishing

`.github/workflows/pages.yml` runs `task index:check site` on every push to
`main` and deploys `_site/` to GitHub Pages:

- the drawer (`index.html` + `wires.json`) at the root
- each `apps/*` wire with an `href`, as its whole folder
- each `tools/*` wire with an `href`: only its `index.html`, plus
  cross-compiled binaries and `SHA256SUMS` in `dl/` for Go tools

One-time setup: **Settings → Pages → Source: GitHub Actions**.

## Adding something that isn't a Go CLI

1. Create `apps/<name>/`, `tools/<name>/` or `labs/<name>/` with a `Taskfile.yml` that has a `ci` task.
2. Add it to `includes:` at the end of the root Taskfile.
3. Put it in the drawer: `task wire:add NAME=<name> PATH=apps/<name> KIND=app|tool|lab BLURB="what it does"`.
   Apps get an `href` automatically. For a tool with its own page, add
   `"href": "tools/<name>/"` and an `index.html`.

`task ci` runs every project that has a Taskfile, and fails if one is
missing from `wires.json` or links a page that doesn't exist.
