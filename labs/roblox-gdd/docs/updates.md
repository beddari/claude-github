# Updates

Which versions the kit pins, where, and how to check an update.

## What is pinned

| What | Version | Where | Moved by |
|---|---|---|---|
| rojo, luau-lsp, selene, StyLua | see the file | `template/rokit.toml` | Renovate, from GitHub releases |
| Roblox type definitions | the luau-lsp version in `template/rokit.toml` | read by `bin/tools` and the skill `rojo-luau` | Renovate, with luau-lsp |
| PageTree, Semantic Notes Vault MCP | the version the skills were checked against | `docs/troubleshooting.md`, README | by hand |
| Roblox Studio's MCP tools | as listed on create.roblox.com/docs/studio/mcp | the skill `roblox-studio-mcp` | by hand |

`bin/tools` reads the versions from `template/rokit.toml`, so CI checks the
template with exactly what a new project installs.

## Check an update

A Renovate pull request that changes `template/rokit.toml` runs
`task roblox-gdd:ci` in the `ci` workflow. It downloads the new versions
and checks the template with them. If it passes, merge it.

For a plugin or Studio update, read its release notes for:

- **PageTree.** Changes to how a page and its folder pair, or to where the
  order is stored. The `obsidian-pagetree-vault` skill and `check-gdd.sh`
  depend on both.
- **Semantic Notes Vault MCP.** Changes to tool or action names, the
  default port, or `.mcpignore`.
- **Studio MCP.** Renamed or new tools. Update the skill's table.

Then run prompt 2 on a test project.
