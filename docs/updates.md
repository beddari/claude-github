# Updates

How newer versions get in, and how a change is checked before it is merged.

## What Renovate proposes

[Renovate](https://docs.renovatebot.com) reads `renovate.json` and opens a
pull request when a pinned version has a newer release. Each version comes
alone, so that a failing check points at one cause.

| Version | Where it is set | Applies | Checked by |
|---|---|---|---|
| step-ca, CoreDNS, etcd and lego images | `labs/acme-env/bin/functions.sh` | the next `task acme-env:up` or `check` | the `labs` workflow |
| cert-manager, external-dns and Traefik charts | `labs/talos-cluster/bin/functions.sh` | the next `task talos-cluster:up` | the `labs` workflow |
| whoami image | `labs/talos-cluster/manifests/whoami.yaml.tmpl` | the next `task talos-cluster:demo` | the `labs` workflow |
| Go | `go.mod` of each module; `go.work` is changed by hand | the next build | the `ci` workflow |
| GitHub Actions | `.github/workflows/*.yml` | the next run | the workflow itself |

The first three are custom managers in `renovate.json`: each one is a
pattern that finds the version in its file. Go and the Actions are
Renovate's own managers.

Renovate has to be running against the repository for any of this to
happen; `renovate.json` is only its configuration. On GitHub, install the
[Renovate app](https://github.com/apps/renovate) for the repository.

## Check a pull request

Two workflows run on every pull request:

| Workflow | Runs | Takes |
|---|---|---|
| `ci` | `task lint` and `task ci`: shellcheck, the drawer's index, the docs' links, every project's `ci` | about 2 minutes |
| `labs` | `task acme-env:up` and `check`, then `task talos-cluster:up` with `verify`; only when the change needs it, see below | about 6 minutes |

Merge when both are green. `labs` runs its job `e2e` when the change touches
anything under `labs/` but the docs, its workflow, or `bin/labs-changed`,
which decides it. A change to the labs' READMEs or `labs/docs/` skips `e2e`,
and GitHub counts a skipped job as passed. To see what a branch would do:

```sh
task labs-changed               # true or false, against origin/main
```

`labs` does not run again on `main` after a merge: it would test the same
files. `ci` does, because it is quick. That holds when the pull request was
tested on the newest `main`. A ruleset makes sure of it, in the
repository's settings under Rules, Rulesets, `default`:

1. Tick "Require status checks to pass". Under "Show additional settings",
   tick "Require branches to be up to date before merging".
2. "Add checks": `ci` and `e2e`, from GitHub Actions.
3. Under "Bypass list", add Repository admin, "For pull requests only".

The bypass keeps a merge by hand possible: an admin gets "Merge without
waiting for requirements" on a pull request. Such a merge reaches `main`
without `e2e` on it, and nothing runs it afterwards. Run it from the Actions
tab, `labs`, "Run workflow", which always runs `e2e`.

A version of a lab image or chart is tested end to end by `labs`, on a fresh
Talos cluster, before it can be merged.

## What a version does not pin

Some versions follow a stream, and what is installed is decided on the day:

| Set in the repository | Installed |
|---|---|
| nothing | Talos and Kubernetes: the defaults of the talosctl that Homebrew installs, today Talos v1.14 and Kubernetes 1.37 |
| `brew install go-task talosctl` in `labs.yml` | the newest Task and talosctl |
| `go install .../task@latest` in `ci.yml` and `pages.yml` | the newest Task |
| `runs-on: ubuntu-latest` | GitHub's current Ubuntu image, with its podman, Docker and dig |
| the chart versions | the images each chart release names |
| `docker.io/library/alpine:3` in `task talos-cluster:debug` | the newest Alpine 3 |

So two runs of the same commit a week apart can differ, and no pull request
says so. For Talos this is a choice: talosctl and the Talos image must match,
and Homebrew keeps talosctl current. A move to a new Talos release shows up
as a failing `labs` run, not as a pull request.

## See what Renovate would do

Without a running Renovate, or to test a change to `renovate.json`:

```sh
npx --package renovate renovate-config-validator renovate.json
RENOVATE_CONFIG_FILE="$PWD/renovate.json" LOG_LEVEL=debug \
	npx renovate --platform=local | grep -E '"(depName|currentValue|newValue)"'
```

The first command checks the file. The second looks every version up and
prints what it found and what it would change; it changes nothing. It reads
the configuration from `RENOVATE_CONFIG_FILE`, because in this mode Renovate
does not read it from the repository. The GitHub Actions need a token for
their lookups: set `GITHUB_COM_TOKEN`, or they are skipped.

A run on 2026-10-06 found 21: the four acme-env images, the three charts,
whoami, Go in the two Go modules, and in the workflows seven Actions and four
`ubuntu-latest` runners. None had a newer release.

## What nothing updates

- **The tools in the Brewfile.** They are not pinned; `brew upgrade` is
  yours.
- **Talos.** See above.
- **The Spotify app's dependencies.** It loads the Spotify Web Playback SDK
  from Spotify at run time.
