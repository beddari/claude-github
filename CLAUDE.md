# CLAUDE.md

A monorepo of small experiments ("wires"). `README.md` has the layout.

## Skills

The SessionStart hook in `.claude/hooks/session-start.sh` links the skills
of https://git.dataverket.org/dataverket/skills into `~/.claude/skills` and
installs Task and shellcheck. Use `bash-style` for every shell script.

## Taskfiles and bin/

The shape of https://git.dataverket.org/dataverket/incusdev-vm:

- A `Taskfile.yml` lists commands. A task that needs a conditional, a loop
  or a computed value calls a script in `bin/` with `{{.CLI_ARGS}}`, one
  script per task, named after it. Shared functions and logging are in
  `bin/functions.sh`; the labs share `labs/lab.sh`.
- Scripts follow the `bash-style` skill. `task lint` runs its shellcheck.
- Tasks have a `desc`, are `silent: true` when they call a script, take
  arguments after `--` (`task acme-env:record-add -- smoke 10.5.0.3`) and
  read settings from the environment. Destructive tasks have a `prompt`.
- Task names use `-`; `:` only separates a project from its task.
- Plain `task` lists the tasks.

## Projects

- Every project has a `Taskfile.yml` with a `ci` task. `task ci` and
  `task lint` at the root pass before a push.
- A new project goes into `wires.json` (`task wire-add` or `task new-go`)
  and into `includes:` at the end of the root `Taskfile.yml`.
- Repository plumbing is Go, in `infra/drawer`. No Python or Node.
- Host tools come from the `Brewfile`.

## Comments

- State the choice a line makes, not what the code already says.
- No history and no evidence of how it was found: the current stance only.
- Reasons go in the docs; the comment points there by path from the root,
  `labs/docs/troubleshooting.md`.
- bash-style's one-line comment above each function stays.

## Docs

Write docs as `docs/writing-docs.md` says, after incusdev-vm: plain short
sentences, real numbers and output, tables for tasks, ports, settings and
versions, and what is not tested. A change to how something runs updates its
README or `docs/` in the same commit.
