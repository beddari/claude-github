# 1. Set up the project

Run this first. It makes a new game project from the template, installs the
toolchain and checks that everything builds. It takes about 5 minutes,
mostly downloads.

**Where:** the Claude Desktop app, **Code** tab, a new session with the
environment set to **Local** and the folder set to a new, empty folder for
your game.

**Before you start:** the steps under "What you need" in the README are done.
Roblox Studio and Obsidian don't need to be open yet.

```text
Set up a new Roblox game project in this folder from the roblox-gdd kit.
Work through these steps in order, tick them off as a task list, and stop
and tell me if a step fails.

1. Check that git and bash are available. On Windows, bash comes with Git
   for Windows. If this folder already holds the template (CLAUDE.md and
   .claude/skills/gdd-code-sync/ exist), skip step 2's copy and go on with
   its `git init`, unless .git/ exists too.
2. Otherwise check that this folder is empty apart from dotfiles, and copy
   the kit's template into it. Clone with LF line endings, also on Windows:
     tmp="$(mktemp -d)"
     git -c core.autocrlf=false clone --depth 1 https://github.com/beddari/claude-github "$tmp/kit"
     cp -R "$tmp/kit/labs/roblox-gdd/template/." .
     rm -rf "$tmp"
   Then `git init` and make a first commit,
   "Start from the roblox-gdd template".
3. Read CLAUDE.md and the four skills in .claude/skills/ before you go on.
4. Check for Rokit with `rokit --version`. If it is missing, stop here.
   Tell me the install command for my OS from
   https://github.com/rojo-rbx/rokit; don't pipe an installer into a shell
   yourself. Tell me that after installing I must quit and reopen Claude
   Desktop, because the app reads PATH only when it starts, and then run
   this prompt again in a new session on this folder.
5. Show me the tools pinned in rokit.toml, then ask me to approve trusting
   them. After I say yes, run `rokit install --no-trust-check`. Check each
   tool with `--version`.
6. Install the Rojo plugin into Roblox Studio with `rojo plugin install`.
7. Run every check in the rojo-luau skill (sourcemap, luau-lsp analyze,
   selene, stylua --check, rojo build) and then check-gdd.sh from the
   gdd-code-sync skill. Fix anything that fails and run them again until
   all pass. Commit nothing that .gitignore lists.
8. Show me a table of each check and its result, and then the steps I do by
   hand next, from the "Connect Obsidian" and "Connect Roblox Studio"
   sections of the kit's README:
   https://github.com/beddari/claude-github/blob/main/labs/roblox-gdd/README.md
