# skill-audit

Checks Claude skills, the folders with a `SKILL.md`, against Anthropic's
current guidance on writing them, and writes the prompt that asks Claude to
fix what it finds. One Go binary with no dependencies beyond the standard
library, for Linux, macOS and Windows.

## What you get

- 11 checks, from the frontmatter to the dependencies of scripts, each with
  pass, info, warn or fail.
- A report as text, markdown or JSON.
- A prompt for Claude per skill, or one for all, that asks it to confirm
  each finding and show a plan before it changes anything.

| Check | Rule |
|---|---|
| `frontmatter` | A valid `name`, and a third-person `description` that says when to use the skill |
| `length` | The body of `SKILL.md` under 500 lines; a warning from 400 |
| `contents` | A reference file over 100 lines starts with a contents list that matches its headings; Claude previews such files with `head -100` |
| `nesting` | Every reference is linked from `SKILL.md` itself, one level deep; files nothing links to are flagged |
| `links` | Local links resolve |
| `freedom` | Risky steps, such as money, deletion, migrations or deploys, run an exact script, not loose prose |
| `models` | The models the skill was tested on are in the frontmatter |
| `emphasis` | Not many ALL-CAPS MUST, NEVER or IMPORTANT, which newer models handle worse |
| `checklist` | An ordered workflow gives Claude a `- [ ]` checklist to copy and tick off |
| `feedback` | A validation step loops: check, fix, repeat until it passes |
| `deps` | Every package and command-line tool has an install line, a manifest entry or PEP 723 metadata |

## Run it

From the root of the repository; Go and Task are enough.

```sh
task skill-audit:install                    # go install: skill-audit on your PATH
skill-audit                                 # audit ~/.claude/skills and ./.claude/skills
skill-audit -prompt > fix.md                # the prompt that fixes them
```

Or without installing, with the arguments after `--`:

```sh
task skill-audit:audit -- ~/.claude/skills
task skill-audit:audit -- -format json path/to/one-skill
```

The text report ends with one line per skill. On the bundled sample skills,
one good and one written to break every rule:

```
Summary
  skill       pass info warn fail  needs work
  Bad_Skill      1    0    5    5  frontmatter, contents, nesting, links, freedom, models, emphasis, checklist, feedback, deps
  good-skill    11    0    0    0
  TOTAL         12    0    5    5
```

On the 13 skills of a Claude Code cloud session in October 2026 it found
that none recorded tested models, and that `pdf` had install lines for 2 of
its 17 packages and tools. The tool's page, [index.html](index.html), has
the whole result; GitHub Pages publishes it with binaries to download.

| Task | What it does |
|---|---|
| `task skill-audit:audit -- [flags] [paths]` | Audit skills and print the report |
| `task skill-audit:report -- [paths]` | Write the report as markdown to `skill-audit.md` |
| `task skill-audit:prompt -- [paths]` | Print one prompt that fixes every skill that needs work |
| `task skill-audit:prompts -- [paths]` | Write one `prompts/<skill>.prompt.md` per skill, and `ALL.prompt.md` |
| `task skill-audit:fix -- [paths]` | Start Claude Code with the prompt as its first message |
| `task skill-audit:checks` | List the checks and their rules |
| `task skill-audit:install` | `go install` the binary into `$GOBIN` or `~/go/bin` |
| `task skill-audit:ci` | vet, the tests, a build, and a check against the sample skills |

The tasks that take paths read them relative to where you run `task`.

## Settings

| Flag | Default | Meaning |
|---|---|---|
| `-format` | `text` | `text`, `markdown` or `json` |
| `-prompt` | off | Print the prompt for Claude in place of the report |
| `-prompt-dir DIR` | none | Also write one prompt per skill into `DIR` |
| `-include-info` | off | Put info-level findings in the prompts too |
| `-only ID,ID` | all | Run only these checks |
| `-strict` | off | Exit 1 when a check fails |
| `-strict-warn` | off | With `-strict`, exit 2 on warnings |
| `-head-lines` | `100` | Lines Claude previews of a reference file |
| `-max-lines` | `500` | The limit for the body of `SKILL.md` |
| `-v`, `-no-color`, `-list-checks` | | More detail, no colour, list the checks |

## How it is built

```
main.go                     the command line
internal/audit/             loading a skill, the checks, the report and the prompt
internal/audit/testdata/    a good and a bad sample skill, for the tests
index.html                  the tool's page on GitHub Pages
Taskfile.yml                the tasks
```

The checks are patterns, so a finding can be wrong; the prompt tells Claude
to say so and skip it. Two things the guidance asks for cannot be checked by
a program, and the prompt lists them for a person: testing the skill on
Haiku, Sonnet and Opus, and how much freedom each step should have.
