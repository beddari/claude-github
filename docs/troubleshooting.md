# Troubleshooting and design notes

For the labs, see [../labs/docs/troubleshooting.md](../labs/docs/troubleshooting.md).

## Where to look

| Question | Command |
|---|---|
| Does everything pass? | `task lint` and `task ci` |
| Is a project missing from the drawer? | `task index-check` |
| Is a link in the docs broken? | `task docs-check` |
| What will GitHub Pages get? | `task site-serve`, then http://127.0.0.1:8000/ |
| Which tasks are there? | `task` |
| Which skills does a cloud session have? | `ls -l ~/.claude/skills` |

## Known behaviour

**`task ci` fails with `is not in the drawer`.** A folder under `apps/`,
`tools/` or `labs/` has a `Taskfile.yml` but no entry in `wires.json`. Add
it: `task wire-add -- NAME PATH KIND "what it does"`.

**`task ci` fails with `no such file` or `no heading with that anchor`.** A
markdown link points at a file or heading that is not there. The message
names the file and line. Anchors are made the way GitHub makes them: lower
case, punctuation dropped, spaces to dashes.

**The pages workflow fails at deploy.** GitHub Pages is off for the
repository. Turn it on once: Settings, Pages, Source: GitHub Actions.

**Spillelista says the redirect URI is invalid.** The Spotify app accepts
only the redirect URIs listed in its dashboard. Add the page's address, such
as `https://beddari.github.io/claude-github/apps/spillelista/`, and
`http://127.0.0.1:8000/` for `task spillelista:serve`.

**A cloud session has no dataverket skills.** The session hook clones them
from `git.dataverket.org`. When that host cannot be reached the hook logs
`!!! Could not fetch` and the session starts without them. A session that
already has the clone keeps the last copy.

**`task new-go` adds the project but `task NAME:run` is unknown.**
`includes:` must stay the last key in the root `Taskfile.yml`, because
`bin/new-go` appends to the end of the file.

## Why it is built this way

**Go and Task, with bash only in `bin/`.** The tools are Go, which builds
the same on Linux, macOS and Windows and needs nothing at run time. The
tasks are commands; anything with logic is a short bash script beside them,
checked by shellcheck. No Python or Node.

**One index, `wires.json`.** The landing page, the check and the site build
all read the same file, so a project is either everywhere or caught by
`task ci`.

**Tool pages publish binaries, not source.** The site is for using a tool;
the source stays on GitHub. Binaries are built by the pages workflow on every
push to `main`, so they always match it.

**Docs checked like code.** A broken link fails `task ci`, and versions in
the docs are kept with Renovate. [writing-docs.md](writing-docs.md) has the
rest.

**The style of incusdev-vm.** The Taskfile, the `bin/` scripts and the docs
follow [incusdev-vm](https://git.dataverket.org/dataverket/incusdev-vm), so
what you learn in one repository applies in the other.

## What is tested

On every pull request, by the `ci` workflow on GitHub's `ubuntu-latest`:
shellcheck on all bash, the drawer's index and the docs' links, the tests of
drawer and skill-audit, skill-audit's check of its sample skills, the
Spotify app's configuration, and the labs' templates. The `labs` workflow
runs both labs end to end when `labs/` changes.

Not tested:

- The tasks on macOS. The scripts use only bash 3, which macOS has, but no
  workflow runs on a Mac, and some commands may differ from GNU's.
- The Spotify app in a browser. Only its configuration is checked; logging
  in needs a Spotify account.
- `task tools` itself. The Brewfile was installed once on Linux with
  Linuxbrew, before `shellcheck` was added to it; no workflow runs it.
