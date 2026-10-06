# skill-audit

Audits Claude skills (`SKILL.md` folders) against Anthropic's updated
skill-authoring best practices, and can generate a prompt asking Claude to
make the fixes. It's a single binary with no dependencies beyond the Go
standard library.

```sh
task install          # or: go build -o bin/skill-audit .
skill-audit           # audits ~/.claude/skills and ./.claude/skills
skill-audit path/to/skills path/to/one-skill path/to/SKILL.md
skill-audit -prompt > fix.md      # a Claude prompt covering every skill that needs work
```

## Checks

| id | rule |
|---|---|
| `frontmatter` | Valid `name`, plus a third-person `description` that says when to use the skill |
| `length` | SKILL.md body under 500 lines (warns from 400) |
| `contents` | Reference files over 100 lines need a contents list in their first 100 lines that matches their headings (Claude previews them with `head -100`) |
| `nesting` | Every reference is linked directly from SKILL.md (one level deep); flags orphans |
| `links` | Local links resolve |
| `freedom` | Fragile operations (money, deletion, migrations, deploys) are backed by a script or an exact command |
| `models` | Tested/intended models are recorded in the frontmatter (e.g. `metadata.tested-models`) |
| `emphasis` | Not over-prescriptive (heavy use of ALL-CAPS MUST/NEVER/IMPORTANT hurts newer models) |
| `checklist` | Ordered multi-step workflows give Claude a copyable `- [ ]` checklist |
| `feedback` | Validation steps follow a fix-and-repeat-until-it-passes loop |
| `deps` | Every Python/Node package and CLI tool used by scripts or code examples has an install line, a manifest entry, or PEP 723 metadata |

The checks are heuristics. The generated prompt tells Claude to confirm
or reject each finding and to show a plan before it edits anything.
Testing on Haiku, Sonnet and Opus, and judging the right degree of
freedom for each step, can't be automated, so the prompt lists those as
manual review items.

## Taskfile

`task --list` shows every target. The main ones:

- `task`: runs vet, tests, build and the fixture self-check, then audits your skills
- `task audit SKILLS="dir1 dir2"`: audits specific folders; pass flags with `task audit -- -v`
- `task prompt` / `task prompts`: prints one combined fix prompt, or writes one `prompts/<skill>.prompt.md` per skill
- `task fix`: starts `claude` with the fix prompt as the first message
- `task audit:md` / `task audit:json`: writes the report as markdown or JSON

## Flags

`-format text|markdown|json`, `-prompt`, `-prompt-dir DIR`,
`-include-info`, `-only id,id`, `-strict` (exits 1 on any fail),
`-strict-warn` (exits 2 on warnings), `-head-lines 100`, `-max-lines 500`,
`-v`, `-no-color`, `-list-checks`.
