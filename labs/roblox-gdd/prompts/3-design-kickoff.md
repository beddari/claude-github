# 3. Design kickoff

Run this once the connections pass. Claude interviews you about the game and
writes the first pages of the design: the pitch, the core loop and a list of
systems. It writes no code. It takes 15 to 30 minutes, depending on how
much you have decided.

**Where:** a Code-tab session on the project folder, with Obsidian open.

```text
Interview me about the game I want to make, then write the first version of
its design in the vault. Use the obsidian-pagetree-vault and gdd-code-sync
skills, and change pages through the obsidian MCP.

Ask me one question at a time, at most 12 in total, and offer 2 or 3
concrete options with each question. Cover, in this order: the genre and a
game it is like; the player fantasy; the minute-to-minute loop; what brings a
player back tomorrow; solo or multiplayer and the server size; how the game
makes money, if it does; the scope of a first playable version in 2 to 4
weeks.

Then:
1. Fill in Game.md: the pitch callout and a short "Pillars" section with 3
   design pillars.
2. Rewrite Core Loop.md as 3 to 6 steps that lead back to the first one,
   each linked to the system that runs it.
3. Under Systems, create one draft child page per system in the first
   playable version, from the skill's template. Write Rules and Exploits
   only as far as we agreed; put open questions in a [!TODO] callout.
4. Replace the "This game" section at the top of CLAUDE.md with the name,
   the pitch in one line and the milestone "first playable".
5. Run check-gdd.sh and fix what it reports. Then show me the page tree and
   suggest which system to build first and why.
