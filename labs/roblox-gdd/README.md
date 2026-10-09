# roblox-gdd

A starting point for making a Roblox game with Claude. The code lives in
Luau files that Rojo (the file-sync tool for Roblox) syncs into Roblox
Studio. The design lives in an Obsidian vault shown as a page tree. Claude
reaches Studio and Obsidian through two MCP servers, and four skills tell it
how to keep the design and the code in agreement. It runs in the Code tab
of the Claude Desktop app, on Windows and macOS.

## What you get

- **A game template** in [template/](template). It is a Rojo project with
  one server entry script and one client entry script, and a typed remote
  that fires when a client loads. It also holds a design vault of four pages
  that matches the code.
- **A project [CLAUDE.md](template/CLAUDE.md)** that says what this game is,
  which tools it has, and which skill to use when. It holds no how-to.
- **Four skills** in [template/.claude/skills/](template/.claude/skills).
  They hold the how-to:

| Skill | What it covers |
|---|---|
| [gdd-code-sync](template/.claude/skills/gdd-code-sync/SKILL.md) | The order of work for a change, and `check-gdd.sh`. The script finds drift between `design/` and `src/` in remotes, code paths, page names and the PageTree layout |
| [rojo-luau](template/.claude/skills/rojo-luau/SKILL.md) | Which file becomes which instance, `default.project.json`, strict Luau, remotes and server checks, and the rojo, luau-lsp, selene and StyLua commands |
| [obsidian-pagetree-vault](template/.claude/skills/obsidian-pagetree-vault/SKILL.md) | PageTree's `X.md` + `X/` layout, the Semantic Notes Vault MCP tools, renames and moves, page frontmatter, callouts, and JSON Canvas |
| [roblox-studio-mcp](template/.claude/skills/roblox-studio-mcp/SKILL.md) | Studio's built-in MCP tools, how to use them alongside Rojo, and the playtest loop |

- **Four prompts** in [prompts/](prompts) that take you from an empty
  folder to the first built system:

| Prompt | What Claude does | Time |
|---|---|---|
| [1-setup](prompts/1-setup.md) | Copies the template, installs the pinned tools, installs the Rojo plugin, and runs every check | about 5 minutes |
| [2-connect](prompts/2-connect.md) | Checks both MCP servers, PageTree and Rojo, and runs the first playtest | about 3 minutes |
| [3-design-kickoff](prompts/3-design-kickoff.md) | Interviews you, then writes the pitch, the core loop and one draft page per system | 15 to 30 minutes |
| [4-first-feature](prompts/4-first-feature.md) | Builds one system end to end: page, code, network contract, checks, playtest | depends on the system |

## What you need

| Part | Version | Where to get it |
|---|---|---|
| Claude Desktop | current, with the **Code** tab | [code.claude.com/docs/en/desktop](https://code.claude.com/docs/en/desktop) |
| Git | any; on Windows, Git for Windows, which also gives Claude Code its bash | [git-scm.com](https://git-scm.com) |
| Roblox Studio | current | [create.roblox.com](https://create.roblox.com) |
| Obsidian | 1.13.0 or later, desktop | [obsidian.md](https://obsidian.md) |
| PageTree plugin | 0.3.7 or later, id `page-tree` | Obsidian → Settings → Community plugins → Browse → "PageTree" |
| Semantic Notes Vault MCP plugin | 0.12.9 or later, id `semantic-vault-mcp` | Obsidian → Settings → Community plugins → Browse → "Semantic Notes Vault MCP" |
| Rokit | 1.2.0 or later | [github.com/rojo-rbx/rokit](https://github.com/rojo-rbx/rokit); prompt 1 tells you the command for your OS |

Rokit installs the rest: rojo, luau-lsp, selene and StyLua, at the versions
in [template/rokit.toml](template/rokit.toml).

The kit uses the Code tab because it is Claude Code. It reads the project's
`CLAUDE.md`, `.claude/skills/` and `.mcp.json`, and it can run the tools.
The Chat tab reads none of these.

## Set it up

**1. Make the project.** In Claude Desktop, open the **Code** tab and start
a session. Set the environment to **Local** and the folder to a new, empty
folder. Paste the prompt from [prompts/1-setup.md](prompts/1-setup.md). At
the end, Claude prints a table of the checks, all passing. If Rokit was
missing, Claude stops and gives you its install command: install it, quit
and reopen Claude Desktop so it sees the new PATH, and run the prompt again
in a new session.

**2. Connect Obsidian.** The design vault is the project's `design/` folder.

1. In Obsidian, choose **Open folder as vault** and pick `design/`.
2. Turn on community plugins.
3. Install and enable **PageTree** and **Semantic Notes Vault MCP**.
4. Run **Open PageTree** from the command palette.
5. In the Semantic Notes Vault MCP settings, turn on **Path exclusions**.
   Its `.mcpignore` section then reads "Current exclusions: 1 patterns
   active", which keeps the MCP out of `.obsidian/`. If it reads 0, close
   and reopen the settings.
6. Copy the API key from the same settings page.
7. In Claude Desktop, open the environment dropdown in the prompt box,
   hover over **Local**, and click the gear icon.
8. Add the variable `OBSIDIAN_MCP_API_KEY` with the key as its value.
   The project's `.mcp.json` reads it.
9. In Obsidian, turn on **Settings → Files & links → Automatically update
   internal links**, so renames keep `[[links]]` working.

**3. Connect Roblox Studio.**

1. Open Studio and open a place. A new **Baseplate** is enough.
2. Go to **Assistant → … → Manage MCP Servers** and turn on **Enable Studio
   as MCP server**.
3. Under **Quick connect**, turn on **Claude Code**.

**4. Check the connections.**

1. Start a new Code-tab session on the same folder. A session reads
   environment variables and MCP servers when it starts. In the session
   from step 1, `${OBSIDIAN_MCP_API_KEY}` reaches the plugin as that literal
   text, and the plugin answers 401.
2. When Claude Code asks whether to use the project's `obsidian` MCP
   server from `.mcp.json`, approve it.
3. In this session's terminal (**Ctrl+\`**), run `rojo serve`, and leave it
   running.
4. In Studio, open the **Rojo** plugin and click **Connect**.
5. Paste [prompts/2-connect.md](prompts/2-connect.md). It ends with a table
   of five checks. The playtest step must show these lines in the console:

```
[Server] started
[Server] ClientReady from <your player name>
```

**5. Design, then build.** Paste
[prompts/3-design-kickoff.md](prompts/3-design-kickoff.md), then
[prompts/4-first-feature.md](prompts/4-first-feature.md) once for each
system.

## The tasks

With this repository cloned, you can also make a project from the command
line, and check the kit:

| Task | What it does |
|---|---|
| `task roblox-gdd:new -- DIR` | Copy the template into `DIR`, relative to where you run `task`, and make its first commit. Prompt 1 then skips the copy |
| `task roblox-gdd:tools` | Download rojo, luau-lsp and StyLua at the pinned versions, and the Roblox type definitions, into `.run/` |
| `task roblox-gdd:check` | Check the template's JSON and canvas; run shellcheck on `check-gdd.sh`, then run it on the template and on a copy with planted problems; run skill-audit on the skills; run rojo, luau-lsp and StyLua on a copy of the template |
| `task roblox-gdd:ci` | The same as `check`, what CI runs |
| `task roblox-gdd:clean` | Remove `.run/` |

`task roblox-gdd:check` prints, in about 5 seconds once the tools are
downloaded:

```
>>> JSON OK: 4 files, every canvas edge joins two nodes.
>>> design/ and src/ agree: 1 remote(s), 4 page(s).
>>> check-gdd.sh OK: the template passes, a fixture's 6 planted problems are found.
>>> skill-audit OK: 4 skills, no failures.
>>> Tools in labs/roblox-gdd/.run/bin: rojo, luau-lsp, stylua.
Created sourcemap at sourcemap.json
Built project to Game.rbxl
>>> Luau OK: rojo sourcemap and build, luau-lsp analyze, stylua --check.
```

## Settings

| Setting | Default | Where | Meaning |
|---|---|---|---|
| `OBSIDIAN_MCP_API_KEY` | none | Claude Desktop's local environment editor | The Semantic Notes Vault MCP API key; `.mcp.json` sends it as a bearer token |
| MCP plugin HTTP port | `3001` | Obsidian, plugin settings | If you change it, change `url` in `.mcp.json` too |
| Roblox Studio MCP | off | Studio, Manage MCP Servers | On only while you work with Claude; it can change the open place |

## How it is built

```
README.md                  this file
Taskfile.yml               the tasks above
bin/                       new, tools, check, and functions.sh that they share
prompts/                   the four prompts, one file each
docs/                      architecture, troubleshooting, updates
template/                  what a new game project starts as
  CLAUDE.md                this game, its tools, which skill when, the rules
  .mcp.json                the obsidian MCP server, http://127.0.0.1:3001/mcp
  rokit.toml               rojo, luau-lsp, selene, StyLua, pinned
  default.project.json     src/server, src/client, src/shared into the DataModel
  .luaurc selene.toml stylua.toml   strict mode, lint, format
  src/                     init scripts, ReadyService, ReadyController, Remotes, Types, Config
  design/                  the vault: Game.md, Game/ (Core Loop, Systems, Networking + canvas), .mcpignore
  .claude/skills/          gdd-code-sync, rojo-luau, obsidian-pagetree-vault, roblox-studio-mcp
```

## The documents

| Document | Content |
|---|---|
| [docs/architecture.md](docs/architecture.md) | The parts, how Claude reaches each one, and why changes go the way they do |
| [docs/troubleshooting.md](docs/troubleshooting.md) | Where to look, known behaviour, why it is built this way, what is tested |
| [docs/updates.md](docs/updates.md) | Which versions are pinned, where, and how to check an update |
