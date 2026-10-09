---
name: gdd-code-sync
description: Keeps the Game Design Document in design/ and the Luau code in src/ in agreement. Reads the design before coding, updates the pages in the same task, and runs scripts/check-gdd.sh to find drift in remotes, code paths, page names and the PageTree layout. Use for any task that changes game behaviour, adds or removes a system or remote, or asks whether the docs match the code.
metadata:
  tested-models: [haiku, sonnet, opus]
---

# Keeping the GDD and the code in sync

A feature is done when the code works and the page describing it is true.
This skill is the order of work. The other skills hold the details:

- `rojo-luau` for code.
- `obsidian-pagetree-vault` for pages.
- `roblox-studio-mcp` for playtests.

## The check script

`scripts/check-gdd.sh` compares `design/` with `src/`. Run it from the
project root:

```bash
bash .claude/skills/gdd-code-sync/scripts/check-gdd.sh
```

It needs bash and the usual text tools: find, grep, sed, awk, sort, uniq,
comm, tr and wc. On Windows these come with Git for Windows, which Claude
Code requires. It reads files with CRLF line endings as LF, and frontmatter
lists in both flow (`[a, b]`) and block (`- a`) form. The script reports:

| Problem | Example |
|---|---|
| A remote in `Remotes.luau` without a `### Name` section in `Networking.md`, or the reverse | `Remote BuyItem is in Remotes.luau but has no '### BuyItem' in Networking.md` |
| A folder of pages without its page next to it | `design/Game/Systems/Combat System/ holds pages but design/Game/Systems/Combat System.md is missing` |
| Two pages with the same name, ignoring case | `More than one page is named 'damage'; [[damage]] is ambiguous` |
| A `code:` path in a page's frontmatter that does not exist | `design/Game/Systems/Shop System.md lists code: src/server/Services/ShopService.luau, which does not exist` |
| A name in `remotes: [...]` that `Remotes.luau` does not declare | `design/Game/Systems/Shop System.md lists remote Ghost, which Remotes.luau does not declare` |

When everything matches, it prints
`>>> design/ and src/ agree: 1 remote(s), 4 page(s).` and exits 0.

## A change to the game

Copy this checklist and tick it off:

- [ ] **Read.** Open the page for the system, its parent, and `[[Networking]]`
      if remotes are involved. Run `graph.backlinks` on the page to see what
      depends on it.
- [ ] **Agree.** If the request conflicts with the page, say how and ask
      which is right before writing code. If no page exists, create a
      `draft` page first and confirm its rules with the user. If no one can
      answer (a scheduled run, or a delegated task), take the request as the
      confirmation and say so in the report.
- [ ] **Code.** Implement under `src/` with the `rojo-luau` skill.
      Tuning numbers go in `src/shared/Config`.
- [ ] **Pages.** Update in the same task:
  - [ ] `status`, `code` and `remotes` in the frontmatter. Set `status`
        to `in-progress` while code exists but has not been playtested, and
        to `implemented` once the playtest passes.
  - [ ] every number or rule that changed, quoted from `Config`
  - [ ] each `[!TODO]` item the change built
- [ ] **Network contract.** For each remote added, renamed or removed,
      change `Remotes.luau`, its `### Name` section in `Networking.md`, and
      its edge in `Networking.canvas` together.
- [ ] **Check.** Run `check-gdd.sh` and the code checks in `rojo-luau`. Fix
      every problem and run them again until both pass.
- [ ] **Playtest.** Run the playtest loop in `roblox-studio-mcp` if Studio is
      connected. If it isn't, say the change was not playtested.
- [ ] **Report.** List the files changed under `src/` and the pages changed
      under `design/`, and say what was tested and what was not.

## An audit

When asked whether the docs match the code, run `check-gdd.sh` first, then
read further. For each page with `status: implemented`:

1. Read the files in its `code:` list.
2. Compare each rule and number on the page with the code and `Config`.
3. List each mismatch with the page, the file and both values.

Report the list. Don't change anything until the user says which side is
right.
