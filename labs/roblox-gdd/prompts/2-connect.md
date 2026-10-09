# 2. Check the connections

Run this after you have opened the vault in Obsidian, turned on Studio's MCP
server, set `OBSIDIAN_MCP_API_KEY`, and started a new session. It checks
both MCP servers and runs the first playtest. It takes about 3 minutes.
When the session first uses the project's `.mcp.json`, Claude Code asks you
to approve the `obsidian` server: approve it.

**Where:** a new Code-tab session on the same folder. Environment variables
and MCP servers are read when a session starts, so a session that was open
before you set them doesn't see them.

**Before you start:**

- Obsidian is open on `design/`.
- Studio is open on a place, such as a new **Baseplate**.
- `rojo serve` is running in the session's terminal (**Ctrl+\`**), and you
  have clicked **Connect** in Studio's Rojo plugin.

```text
Check that this project's tools are connected. Use the skills in
.claude/skills/. Report each step as pass or fail with what you saw, and
stop at the first failure with what I should do about it.

1. Obsidian: run vault list on the obsidian MCP server, and read
   Game/Networking.md through it. The vault should list Game.md and the
   Game/ folder. Then ask me to confirm that "Path exclusions" is on in the
   Semantic Notes Vault MCP settings, and that its ".mcpignore file
   management" section reads "Current exclusions: 1 patterns active". If
   it reads 0, I close and reopen the settings: the file loads a moment
   after the toggle.
2. PageTree: ask me to confirm that the PageTree view shows Game with Core
   Loop, Systems and Networking under it.
3. Studio: call list_roblox_studios. If more than one Studio is open, ask me
   which one is this project.
4. Rojo: use search_game_tree to find ServerScriptService.Server.Services.ReadyService
   and ReplicatedStorage.Shared.Remotes. If they are missing, Rojo isn't
   connected.
5. Playtest: follow the playtest loop in the roblox-studio-mcp skill. The
   console must show "[Server] started" and "[Server] ClientReady from"
   followed by my player name, with no errors. Stop the playtest afterwards.
6. Give me a table of the five checks. If they all passed, tell me the next
   prompt to run is 3-design-kickoff.
