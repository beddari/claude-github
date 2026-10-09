# 4. Build the first system

Run this for each system you build. Claude builds it end to end: it reads
the page, writes the code, updates the pages and the network contract, runs
the checks and playtests in Studio. Replace `<System>` with a page name from
the vault, such as `Coins System`.

**Where:** a Code-tab session on the project folder, with Obsidian open,
Studio open, `rojo serve` running and the Rojo plugin connected.

```text
Build [[<System>]] as its page describes. Follow the gdd-code-sync skill's
checklist from Read to Report, and use rojo-luau for the code,
obsidian-pagetree-vault for the pages and roblox-studio-mcp for the
playtest.

Before you write code, list the rules you will implement, the remotes you
will add with their server checks, and the Config values. Wait for me to say
go. If the page is missing something you need, ask instead of inventing it.

When you are done, give me the report the skill describes. Include the
playtest console lines that show the system working.
