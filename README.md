# claude-github

A drawer of experiments, "wires of assorted lengths": small apps, tools and
labs, each in its own folder with its own tasks. The root `index.html` is a
landing page with one wire per experiment, published on GitHub Pages.

## What you get

| Wire | Kind | What it does |
|---|---|---|
| [spillelista](apps/spillelista) | app | Spotify playlist page: log in, pick a list, play |
| [skill-audit](tools/skill-audit) | tool | Checks Claude skills against current guidance and writes the prompt to fix them |
| [acme-env](labs/acme-env) | lab | Split DNS and an internal ACME CA in podman: CoreDNS, etcd and step-ca |
| [talos-cluster](labs/talos-cluster) | lab | Talos Kubernetes in Docker that gets DNS records and certificates from acme-env |

On GitHub Pages, from every push to `main`:

- the landing page, drawn from `wires.json`;
- each app, as it is;
- each tool's page, with binaries for Linux, macOS and Windows.

## Run it

You need [Homebrew](https://brew.sh), on macOS or Linux, for the tools. Go and
Task are enough for everything except the labs, which also need podman and a
Docker Engine.

```sh
brew bundle        # the tools: go, task, podman, talosctl, kubectl, helm, step, dig, jq, shellcheck
task ci            # every check: about 50 seconds the first time, 5 seconds after
task site-serve    # the site as GitHub Pages will have it, at http://127.0.0.1:8000/
```

`task ci` ends with each project's result:

```
wires.json OK (4 wires)
docs OK (14 markdown files, all relative links resolve)
>>> infra/drawer/
>>> apps/spillelista/
>>> tools/skill-audit/
>>> labs/acme-env/
>>> labs/talos-cluster/
>>> All good: every project's ci passed.
```

| Task | What it does |
|---|---|
| `task` | List the tasks, the projects' included |
| `task tools` | Install the tools in the Brewfile |
| `task ci` | Check the drawer's index and the docs' links, then run every project's `ci` |
| `task lint` | Run shellcheck on every bash script |
| `task labs-changed -- [BASE]` | Say whether the changes since BASE, `origin/main` by default, need the labs run end to end |
| `task index-check` | Check that `wires.json` lists every project |
| `task docs-check` | Check that every relative link in the markdown files resolves |
| `task serve` | Serve the repository as it is at http://127.0.0.1:8000/ |
| `task site` | Build the GitHub Pages site into `_site/` |
| `task site-serve` | Build `_site/` and serve it |
| `task new-go -- NAME "BLURB"` | Make a Go tool at `tools/NAME` and register it everywhere |
| `task wire-add -- NAME PATH KIND "BLURB"` | Add a project to the drawer |
| `task clean` | Remove `_site/` and every project's build output |

A project's tasks run from the root with its name in front:
`task acme-env:up`, `task skill-audit:audit`.

## Add a wire

A Go tool, in one step. It copies `templates/go-cli`, adds the module to
`go.work`, the project to `wires.json` and to the root `Taskfile.yml`, and
runs its `ci`:

```sh
task new-go -- hello "Says hello"
task hello:run
```

Anything else: make `apps/NAME/`, `tools/NAME/` or `labs/NAME/` with a
`Taskfile.yml` that has a `ci` task, then

```sh
task wire-add -- NAME labs/NAME lab "What it does"
```

and add it to `includes:` at the end of the root `Taskfile.yml`. `task ci`
fails until both are done. An app gets an `href` and is published; for a tool
page, give the entry `"href": "tools/NAME/"` and the tool an `index.html`.

## How it is built

```
index.html, wires.json   the landing page and the index it draws
assets/                  the landing page's background photo
apps/ tools/ labs/       the wires, each with a Taskfile.yml and a README
infra/drawer/            Go: checks wires.json and the docs, builds the site, serves it
templates/go-cli/        what task new-go copies
bin/                     the root tasks' scripts: ci, lint, clean, new-go, labs-changed
Taskfile.yml             the root tasks, and every project's under includes:
go.work                  the Go workspace: every module
Brewfile                 the tools
renovate.json            lets Renovate propose newer versions
.github/workflows/       ci, labs and pages
.claude/                 the session hook for Claude Code in the cloud
CLAUDE.md                the conventions, for Claude Code
docs/                    the details
```

| Document | Content |
|---|---|
| [docs/architecture.md](docs/architecture.md) | The projects, the drawer, the tasks and the workflows |
| [docs/updates.md](docs/updates.md) | Which versions are pinned, how Renovate updates them, how a change is checked |
| [docs/troubleshooting.md](docs/troubleshooting.md) | Where to look, known behaviour, why it is built this way, what is tested |
| [docs/writing-docs.md](docs/writing-docs.md) | How the docs are written |
| [labs/docs/](labs/docs/architecture.md) | The labs: architecture, access, troubleshooting |
