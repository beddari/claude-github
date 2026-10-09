---
name: roblox-studio-mcp
description: Uses the Roblox Studio MCP server, which is built into Studio, to inspect the open place, playtest, read the console, capture the screen and drive a test player, without fighting Rojo. Use when asked to test, playtest or debug the game in Studio, look at the live data model, or check that a change works in play.
---

# Roblox Studio MCP

Roblox Studio has its own MCP server. Turn it on in Studio under
**Assistant → … → Manage MCP Servers → Enable Studio as MCP server**, then
connect Claude Code with **Quick connect**. A green dot in **Manage MCP
Servers** means a client is connected. The server can read and change the
open place, so it is on only while you use it.

Every tool takes a `studio_id`. When more than one Studio window is open,
call `list_roblox_studios` first and pick the one whose place is this project.

## What each tool is for here

| Use | Tools |
|---|---|
| Look at the live tree | `search_game_tree`, `inspect_instance` |
| Read scripts as Studio has them | `script_read`, `script_search`, `script_grep` |
| Playtest | `get_studio_state`, `start_stop_play`, `get_console_output` |
| See the screen | `screen_capture` |
| Act as a player | `character_navigation`, `user_keyboard_input`, `user_mouse_input` |
| Run Luau in Studio | `execute_luau` |
| Assets | `search_asset`, `insert_asset`, `generate_mesh`, `generate_material`, `generate_procedural_model`, `wait_job_finished`, `upload_image`, `store_image` |
| Roblox docs | `http_get`, `skill` |

## Working alongside Rojo

Rojo owns every instance that comes from `src/`. Writes to those instances
through MCP are overwritten at the next sync, and the files no longer match
what runs in Studio.

- **Change code in files.** Edit under `src/` and let `rojo serve` sync it.
  Never use `multi_edit` on a script that comes from `src/`.
- **Use `execute_luau` to read.** Print state, count instances, check that an
  instance exists. Ask before running Luau that changes or destroys anything.
- **Use the place file only for what Rojo doesn't manage.** Map geometry,
  lighting and inserted or generated assets are fine to change with
  `insert_asset`, `generate_*` or `execute_luau`. These changes live in the
  place file, not in git, so tell the user what you changed, and remind them
  to save the place.
- **Check sync first.** If `script_read` shows code that differs from the
  file, the Rojo plugin is not connected. Ask the user to start
  `rojo serve` and click **Connect** in the Rojo plugin.

## Playtest loop

Copy this checklist and tick it off:

- [ ] `rojo serve` is running and the Rojo plugin shows it is connected.
- [ ] `get_studio_state`: Studio is in edit mode, not already playing.
- [ ] `start_stop_play` to start a play session.
- [ ] Wait a few seconds, then `get_console_output`.
- [ ] Look for the lines the change should print, and for any error or
      warning. A stack trace names the script and line; open the file under
      `src/` that maps to that script.
- [ ] If the change is visual or about input, use `screen_capture` or the
      `character_navigation` and `user_*_input` tools.
- [ ] `start_stop_play` to stop. Leave Studio in edit mode.
- [ ] If anything failed, fix it in `src/`, wait for Rojo to sync, and start
      the loop again. Repeat until the console is clean.

A clean start of this project prints `[Server] started` and
`[Server] ClientReady from <player>`.

## When it fails

| Symptom | Likely cause |
|---|---|
| No `Roblox_Studio` tools in the session | MCP is off in Studio, or Claude Code was not restarted after Quick connect |
| `list_roblox_studios` is empty | Studio is closed, or the place is not open |
| A code change has no effect in play | Rojo is not connected, or the file is in a folder `default.project.json` doesn't map |
| `Infinite yield possible on ...Remotes` | `Remotes.luau` was not required on the server before a client asked |
