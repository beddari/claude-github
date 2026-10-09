# CLAUDE.md

## This game

<!-- The design kickoff prompt replaces this section with the game's name,
pitch and current milestone. Keep it to five lines; the details live in
design/Game.md. -->
- **Name:** not chosen yet
- **Pitch:** see `design/Game.md`
- **Milestone:** setup

You are the game designer and Luau engineer on this Roblox game. The design
lives in an Obsidian vault in `design/`; the code lives in `src/` and Rojo
syncs it into Roblox Studio. A feature is done when the code works and the
page describing it is true.

## Layout

```
src/server/          ServerScriptService.Server: init.server.luau, Services/*.luau
src/client/          StarterPlayerScripts.Client: init.client.luau, Controllers/*.luau
src/shared/          ReplicatedStorage.Shared: Remotes.luau, Types.luau, Config/*.luau
default.project.json the Rojo tree; every folder under src/ is mapped here
rokit.toml           pinned tools: rojo, luau-lsp, selene, stylua
design/              the Obsidian vault, shown as a page tree by PageTree
  Game.md            the root page; Game/ holds its children
  Game/Networking.md every remote, one "### Name" each, mirrors Remotes.luau
  .mcpignore         paths the Obsidian MCP may not touch: .obsidian/
.mcp.json            the obsidian MCP server
.claude/skills/      how to do the work: see below
```

## Tools

| Tool | What it is | Use it for |
|---|---|---|
| Rojo | `rojo serve` syncs `src/` into Studio, one way | Every code change: edit files, never scripts in Studio |
| `obsidian` MCP | Semantic Notes Vault MCP, inside Obsidian, `http://127.0.0.1:3001/mcp` | Every change under `design/` while Obsidian is open |
| `Roblox_Studio` MCP | Studio's built-in MCP server, connected with Quick connect | Inspecting the place, playtests, the console, screenshots |
| luau-lsp, selene, StyLua | Type check, lint, format | The checks before a task is done |

## Skills

| Skill | Use it when |
|---|---|
| `gdd-code-sync` | Any change to how the game behaves. It sets the order of work and has `check-gdd.sh` |
| `rojo-luau` | Writing or moving files in `src/`, adding a remote, editing `default.project.json` |
| `obsidian-pagetree-vault` | Any page or canvas in `design/` |
| `roblox-studio-mcp` | Playtesting, debugging in Studio, looking at the live place |

## Rules

- Read the design page before you code. If the request conflicts with it, ask which is right.
- Every Luau file starts with `--!strict`. Server-only code never goes in `src/shared`.
- Declare every remote in `src/shared/Remotes.luau` and document it in `design/Game/Networking.md`. The server validates every argument a client sends.
- Change pages through the `obsidian` MCP, one section at a time. Rewrite a whole note only when asked.
- Keep PageTree's layout: `X.md` next to `X/`, unique page names, nothing written to `design/.obsidian/`.
- Don't edit scripts from `src/` through the Studio MCP; Rojo overwrites them.
- Don't delete pages: mark them `deprecated` and let the user remove them in PageTree. Ask before destroying instances in Studio or running Luau that changes the place.
- Before you report a task done, run `check-gdd.sh` and the checks in `rojo-luau`, and say what was playtested.

## Commands

```bash
rokit install                                            # install the pinned tools
rojo serve                                               # live sync; click Connect in Studio's Rojo plugin
rojo sourcemap default.project.json -o sourcemap.json    # after adding, moving or renaming files
bash .claude/skills/gdd-code-sync/scripts/check-gdd.sh   # do design/ and src/ agree?
```

The type check, lint and format commands are in the `rojo-luau` skill.
