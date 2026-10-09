---
name: rojo-luau
description: Writes and checks Luau code in a Rojo-synced Roblox project. Covers which file name becomes which instance, default.project.json, strict typing, remotes, server-side validation, and the rojo, luau-lsp, selene and StyLua checks. Use when creating, moving or reviewing files under src/, adding a remote, or editing default.project.json.
metadata:
  tested-models: [haiku, sonnet, opus]
---

# Rojo and Luau

Rojo (the file-sync tool for Roblox) builds the instance tree from
`default.project.json` and the files under `src/`. It syncs one way, from
files to Studio. Studio edits to synced scripts are lost at the next sync, so
all code changes happen in files.

## How files become instances

| File | Instance |
|---|---|
| `Name.server.luau` | `Script` |
| `Name.client.luau` | `LocalScript` |
| `Name.luau` | `ModuleScript` |
| `init.server.luau`, `init.client.luau`, `init.luau` in a folder | The folder becomes that Script, LocalScript or ModuleScript, and its other files become children |
| `Name.meta.json`, `init.meta.json` | Properties and attributes of the sibling file or the folder |
| `Name.model.json` | A tree of instances, such as Folders or Values |
| `Name.json` | A ModuleScript that returns the JSON as a table |
| a folder | `Folder` |

`default.project.json` maps paths to services with `$path`. A folder of
`src/` that no `$path` reaches is not synced. A new top-level folder needs
its own `$path` entry in the same change.

Use `.luau` for every file. Do not mix in `.lua`.

## Where code goes

| Folder | Instance | Runs on | Rule |
|---|---|---|---|
| `src/server/` | `ServerScriptService.Server` | server | Secrets, data stores, authority |
| `src/client/` | `StarterPlayer.StarterPlayerScripts.Client` | each client | Input, UI, effects |
| `src/shared/` | `ReplicatedStorage.Shared` | both | Every client can read it, so no secrets and no server logic |

The project has one entry script per side: `src/server/init.server.luau` and
`src/client/init.client.luau`. They `require` every module in
`Server/Services` and `Client/Controllers` in name order and call `:start()`.
New logic is a new ModuleScript there, not a new Script.

The modules start one after another. If `start` yields, it holds up every
module after it. Yielding calls include `InvokeServer`, `task.wait` and
`WaitForChild` on something that may never arrive. Put them in
`task.spawn`.

## Luau rules

- The first line of every file is `--!strict`. `.luaurc` also sets strict mode.
- Annotate the parameters and the return type of every function.
- Put types that both sides use in `src/shared/Types.luau` as `export type`.
- Write requires as instance paths that follow the tree:
  `require(ReplicatedStorage.Shared.Remotes)`.
- Put tuning numbers in `src/shared/Config/<System>.luau`, one file per
  system, frozen with `table.freeze`. `Config/Game.luau` holds values that no
  single system owns. The design pages quote these values, so changing one
  means updating its page too.

## Remotes

`src/shared/Remotes.luau` is the only place a RemoteEvent or RemoteFunction is
declared. Each declaration is a quoted name on its own line, inside
`REMOTE_EVENTS` or `REMOTE_FUNCTIONS`. Get a remote with
`Remotes.event("Name")` or `Remotes.func("Name")`.

To add a remote:

- [ ] Add the quoted name to `REMOTE_EVENTS` or `REMOTE_FUNCTIONS`.
- [ ] In the server handler, check each argument before acting on it. See the list below.
- [ ] Add a `### Name` section to `design/Game/Networking.md`. The
      `obsidian-pagetree-vault` skill has the fields it needs.
- [ ] Run the checks below.

The server trusts nothing a client sends. A handler checks, in this order:

1. **Rate.** Has the player called this too often? Keep a per-player
   timestamp, and clear it on `Players.PlayerRemoving`. It comes first
   because it is cheapest and stops a client from flooding the log.
2. **Count and type.** `typeof(x) == "number"`. Reject NaN with `x ~= x`.
3. **Range.** Is the value within the limits in `Config`?
4. **Ownership.** Does the player own the thing they name?

If any check fails, return straight away and never error inside the
handler. Warn when the call can only come from a modified client, such as a
wrong type. Stay silent when it is rate-limited.

- A RemoteEvent handler returns nothing.
- A RemoteFunction returns `nil` when it refuses. Its return type is
  optional (`number?`), and the client handles `nil`.

Choose the rate limit, put it in `Config`, and write it in the remote's
section of `Networking.md`. `ReadyService.luau` is the worked example.

## Checks

Run these from the project root. Install the tools first with
`rokit install`; the versions are pinned in `rokit.toml`.

```bash
rojo sourcemap default.project.json -o sourcemap.json
lsp="$(sed -n 's/^luau-lsp = ".*@\(.*\)"/\1/p' rokit.toml)"
mkdir -p .luau-lsp && curl -sSfL -o .luau-lsp/globalTypes.d.luau \
  "https://raw.githubusercontent.com/JohnnyMorganz/luau-lsp/$lsp/scripts/globalTypes.d.luau"
luau-lsp analyze --platform=roblox --sourcemap=sourcemap.json \
  --definitions=@roblox=.luau-lsp/globalTypes.d.luau --base-luaurc=.luaurc src
selene src
stylua --check src
mkdir -p build && rojo build -o build/Game.rbxl
```

- Regenerate the sourcemap after you add, move or rename a file, or edit
  `default.project.json`. Otherwise luau-lsp cannot resolve requires.
- Download the definitions once per luau-lsp version: they come from the
  version `rokit.toml` pins. The folder is git-ignored.
- On the first run, selene downloads the Roblox API.
- Run `stylua src` to fix formatting.

Fix every error and run the checks again until all five pass.

## Common mistakes

- **Server code in `src/shared`.** It is visible to every client.
- **A new folder that isn't synced.** If a folder under `src/` doesn't show
  in Studio, its path is missing from `default.project.json`.
- **A client file named `.server.luau`, or the other way round.** It runs
  on the wrong side.
- **`WaitForChild` on the server for an instance Rojo creates.** Rojo
  instances exist before scripts run, so index them directly.
