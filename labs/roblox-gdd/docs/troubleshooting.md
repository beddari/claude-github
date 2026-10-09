# Troubleshooting and design notes

Where to look when something does not work, what is known to behave
oddly, why the kit is built this way, and what has been tested.

## Where to look

| Question | Where |
|---|---|
| Does `design/` match `src/`? | `bash .claude/skills/gdd-code-sync/scripts/check-gdd.sh` in the project |
| Is the obsidian MCP server connected? | `/mcp` in the Code session; the plugin's settings show "Server running" |
| Is the Studio MCP server connected? | Studio → Assistant → … → Manage MCP Servers: a green dot per connected client |
| Is Rojo syncing? | The Rojo plugin in Studio says **Connected**; `rojo serve` prints the client |
| Does the kit itself pass? | `task roblox-gdd:check` in this repository |
| Is the API key reaching the session? | Ask Claude to run `echo ${OBSIDIAN_MCP_API_KEY:+set}`; it prints `set` |

## Known behaviour

**The obsidian server fails with 401.** The session started before
`OBSIDIAN_MCP_API_KEY` was set, or the key changed. Claude Code still loads
`.mcp.json` with the variable unset, and sends the literal text
`${OBSIDIAN_MCP_API_KEY}` as the key. Set it in the local environment editor
and start a new session. On macOS, a variable exported in `~/.zshrc` does
not reach the app when it starts from the Dock.

**The obsidian server is missing from `/mcp`.** Claude Code asks once
before it uses a server from a project's `.mcp.json`. If it was declined,
run `claude mcp reset-project-choices` in the session's terminal and start
a new session.

**The obsidian server worked yesterday and now does not connect.** When
port 3001 is taken, the plugin moves to 3002 for that run and says so only
in an Obsidian notice. Free the port, or set the port in the plugin and in
`.mcp.json` to the same free one.

**`rokit` is not found right after installing it.** Claude Desktop reads
PATH when it starts. Quit and reopen it, then start a new session.

**check-gdd.sh fails with `$'\r': command not found`, or StyLua flags every
file.** The files have CRLF line endings. The template's `.gitattributes`
keeps them LF; a clone made before it, or with `core.autocrlf=true`, needs
`git add --renormalize .` and a fresh checkout.

**The obsidian server fails to connect.** Obsidian is closed, or the vault
open in it is not this project's `design/`. Each vault runs its own server;
two vaults with the plugin need two ports.

**A page moved to another parent left its children behind.** PageTree moves
the paired folder only for a rename in the same folder. `vault.move` moves
one file. The `obsidian-pagetree-vault` skill moves the children one by one,
or asks you to drag the page in the PageTree view.

**A moved page is at the end of its siblings.** PageTree saw it as new.
Drag it into place.

**`selene src` fails on the first run without network.** Its `roblox`
standard library downloads the Roblox API dump. Run it once online.

**`rojo build -o build/Game.rbxl` fails with "No such file or directory".**
The `build/` folder must exist: `mkdir -p build`.

**An `InvokeServer` call in a controller's `start` holds up the client.**
The controllers start one after another, and `InvokeServer` waits for the
server's answer. The `rojo-luau` skill puts yielding calls in `task.spawn`.

## Why it is built this way

**CLAUDE.md is short and the skills are long.** CLAUDE.md is read in every
session, so it holds what is true of this project: the game, the tools and
the rules. The how-to is in the skills, which Claude loads only when a task
needs them.

**One `.mcp.json` entry, not two.** The Studio MCP server is started by a
different command on Windows and on macOS, and Studio's Quick connect
writes it for you. The Obsidian server is the same everywhere, so it is in
the project file, with the key read from the environment.

**HTTP, not HTTPS, to Obsidian.** The plugin binds to loopback only. Its
HTTPS port uses a self-signed certificate that Claude Code trusts only
through `NODE_EXTRA_CA_CERTS`.

**Remotes declared as a list of names.** `check-gdd.sh` reads them with awk.
One quoted name per line keeps that reliable without a Luau parser.

**Pages are deprecated, not deleted.** Deleting a page in PageTree deletes
its children too. The skill leaves removal to you.

**The kit is a lab.** It has no binary to download, like a tool, and it is
not a web app. Its end-to-end check needs Studio, which CI cannot run.

## What is tested

| What | How | Where |
|---|---|---|
| Template JSON and canvas edges | `jq` | `task roblox-gdd:check`, CI |
| `check-gdd.sh` | shellcheck; a run on the template; a run on a copy with CRLF line endings, a one-line remote table, a commented-out remote, flow and block frontmatter lists, and 6 planted problems, each of which must be reported | CI |
| The four skills | skill-audit: no failures; `roblox-studio-mcp` warns that it has no tested models | CI |
| Template Luau | rojo 7.7.1 sourcemap and build, luau-lsp 1.70.1 analyze in strict mode with the Roblox types, StyLua 2.5.2 | CI |
| The skills with Haiku, Sonnet and Opus | Each model added the same Coins system to a copy of the template, with Obsidian and Studio closed. All three passed `check-gdd.sh`, luau-lsp, StyLua and rojo build, rerun independently. The gaps they reported are now in the skills | by hand, October 2026 |

Not tested:

- selene, which needs Roblox's API servers.
- The Studio MCP server and the playtest loop, which need Roblox Studio on
  Windows or macOS.
- The Semantic Notes Vault MCP plugin and PageTree in a running Obsidian.
  The skills follow their source code (PageTree 0.3.7, the MCP plugin
  0.12.9), not a live run.
- The prompts, end to end, in Claude Desktop.
- `check-gdd.sh` with macOS's awk and sed, and with Git for Windows. It is
  written for bash 3 and POSIX tools, and runs in CI with mawk and GNU sed.
