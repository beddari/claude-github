---
name: obsidian-pagetree-vault
description: Creates, edits, renames and moves pages in the design/ Obsidian vault through the Semantic Notes Vault MCP plugin, keeping the PageTree plugin's file and folder layout intact. Also covers page frontmatter, callouts, wiki-links and JSON Canvas diagrams. Use for any change under design/, any Obsidian or GDD page, or any .canvas file.
metadata:
  tested-models: [haiku, sonnet, opus]
---

# The design vault

`design/` is an Obsidian vault. Two plugins work on it:

- **PageTree** (`page-tree`) shows the vault as a tree of pages. Each page
  can have both its own text and child pages.
- **Semantic Notes Vault MCP** (`semantic-vault-mcp`) runs an MCP server
  inside Obsidian, at `http://127.0.0.1:3001/mcp`. In this project the server
  is named `obsidian`. Obsidian has to be running with this vault open.

## PageTree's layout

Every page is a normal `.md` file. A page's children live in a folder next to
it, with the same name:

```
Game/Systems.md                         the Systems page
Game/Systems/Combat System.md           a child of Systems
Game/Systems/Combat System/Damage.md    a child of Combat System
```

PageTree adds nothing to the notes; the paths are the tree. It keeps the
order of siblings in `.obsidian/plugins/page-tree/data.json`. Four rules
follow:

1. **A folder with pages needs its page.** Before creating
   `Game/Systems/Combat System/Damage.md`, check that
   `Game/Systems/Combat System.md` exists; if it doesn't, create it.
2. **Page names are unique across the vault.** `[[Damage]]` must have one
   target. Use a precise name like `Combat Damage`, not `Damage` or
   `Overview`.
3. **Leave `.obsidian/` alone.** New pages go to the end of their siblings,
   and the user reorders them in PageTree. `.mcpignore` blocks `.obsidian/`
   once **Path exclusions** is on in the MCP plugin's settings.
4. **Files other than pages go in `_assets/`.** It has no `.md` files.
   Folders whose names start with `.` are invisible to PageTree.

## The MCP tools

Each tool takes an `action` parameter.

| Tool | Actions to use |
|---|---|
| `vault` | `list`, `read`, `search`, `fragments`, `create`, `update`, `rename`, `move`, `copy` |
| `edit` | `patch` (target a `heading`, `block` or `frontmatter` field), `window` (fuzzy find and replace), `append`, `at_line` |
| `view` | `file`, `window`, `active`, `open_in_obsidian` |
| `graph` | `backlinks`, `forwardlinks`, `neighbors`, `traverse`, `path`, `statistics` |

- **Change a section, not the whole note.** Use `edit.patch` or
  `edit.window`. `vault.update` replaces the whole note, so use it only when
  the user asks for a rewrite.
- **Find dependants first.** Before changing a mechanic, run `graph.backlinks`
  on its page and `vault.search` for its name.
- **Don't use `vault.split`, `vault.combine` or `vault.concatenate`.**
  `split` writes `name-1.md` next to the page, which is outside PageTree's
  layout.
- **Leave removal to the user.** When a page should go, set its `status` to
  `deprecated` and tell the user. They remove it in PageTree, which also
  removes its children.

## Rename and move

`vault.rename` and `vault.move` go through Obsidian's file manager, so
PageTree and Obsidian's link updater see them.

| Change | Steps |
|---|---|
| Rename, same parent | `vault.rename` on `X.md`. PageTree renames `X/` with it. |
| Move to another parent | `vault.move` `X.md` into the new parent's folder. Then `vault.move` each file under `X/`, deepest last. Or ask the user to drag the page in PageTree, which moves the whole subtree. |

Before either change, list the page's backlinks. Afterwards, check that each
of them still resolves. Obsidian rewrites links only when **Settings → Files
& links → Automatically update internal links** is on.

## A system page

Copy this checklist for a new system page:

- [ ] Make sure the parent page exists, usually `Game/Systems.md`.
- [ ] `vault.create` `Game/Systems/<Name> System.md` from the template below.
- [ ] Link the new page from its parent with `edit.append` or `edit.patch`.
      If the parent still has a placeholder `[!TODO]`, replace it with a list
      of its child pages.
- [ ] Link it from `[[Core Loop]]` where it takes part in the loop.
- [ ] If it uses remotes, add their `### Name` sections to `[[Networking]]`.
- [ ] Run the `gdd-code-sync` check script. Fix what it reports, and run it again until it passes.

```markdown
---
status: draft
code: []
remotes: []
---
# <Name> System

> [!NOTE] Design intent
> What the player feels and why this system exists, in two sentences.

## Rules
- Each rule as one line, with numbers from `src/shared/Config`.

## Data
- What the server stores per player, and where.

> [!WARNING] Exploits
> What a modified client could try, and the server check that stops it.

> [!TODO] Not built yet
> - The parts the code doesn't do yet.
```

**Frontmatter fields**

- `status`: one of `draft`, `in-progress`, `implemented` or `deprecated`.
- `code`: the files under `src/` that implement the page, as a flow list on
  one line, for example `[src/server/Services/ShopService.luau]`.
- `remotes`: the remote names the page uses, as a flow list on one line,
  for example `[BuyItem, ItemBought]`.

Keep both lists on one line. `edit.patch` on a `frontmatter` field replaces
only the line with the key, so it would leave the `- item` lines of a block
list behind as broken YAML. If a list has become a block list (Obsidian's
Properties view writes them that way), change it with `edit.window`, or turn
it back into a flow list. `check-gdd.sh` reads both forms.

**Callouts**

- `[!NOTE]` for intent and loops.
- `[!WARNING]` for trust boundaries and exploits.
- `[!TODO]` for designed but unbuilt parts.

**Links**

- `[[Page]]` links to a page.
- `[[Page#Heading]]` links to a section.
- `![[File.canvas]]` embeds a canvas.

## A remote's section in Networking

Each remote in `src/shared/Remotes.luau` has exactly one section under
`## Remotes` in `Game/Networking.md`:

```markdown
### BuyItem
- **Kind:** RemoteEvent
- **Direction:** client → server
- **Payload:** `itemId: string`
- **Server checks:** item exists; player can afford it; at most 2 per second
- **Code:** `src/client/Controllers/ShopController.luau` fires it, `src/server/Services/ShopService.luau` handles it
```

A RemoteFunction also gets a `- **Returns:**` line, such as
`number?, nil when refused`, after **Payload**.

## Canvas

A `.canvas` file is JSON Canvas, a feature built into Obsidian. PageTree
doesn't show `.canvas` files, so embed each one in the page it belongs to.

```json
{
  "nodes": [
    {"id": "client", "type": "group", "label": "Client", "x": -480, "y": -160, "width": 400, "height": 280},
    {"id": "server", "type": "group", "label": "Server", "x": 80, "y": -160, "width": 400, "height": 280},
    {"id": "shop-ctl", "type": "text", "text": "**ShopController**", "x": -440, "y": -80, "width": 320, "height": 80},
    {"id": "shop-svc", "type": "text", "text": "**ShopService**", "x": 120, "y": -80, "width": 320, "height": 80},
    {"id": "page", "type": "file", "file": "Game/Systems/Shop System.md", "x": 120, "y": 160, "width": 320, "height": 200}
  ],
  "edges": [
    {"id": "buy", "fromNode": "shop-ctl", "fromSide": "right", "toNode": "shop-svc", "toSide": "left", "toEnd": "arrow", "label": "BuyItem (C→S)"}
  ]
}
```

**Nodes**

- Every node has `id`, `type`, `x`, `y`, `width` and `height`.
- `type` is one of `text`, `file`, `link` or `group`.
- `text` nodes take `text`, `file` nodes take a vault path in `file`, and `group` nodes take `label`.

**Edges**

- Every edge has `id`, `fromNode` and `toNode`.
- `fromSide` and `toSide` are each one of `top`, `right`, `bottom` or `left`.
- `toEnd` is `arrow` or `none`, and `label` is optional.

**In this project**

`Game/Networking.canvas` has a **Client** group and a **Server** group, one
node per script, and one edge per remote, labelled with its name and
direction.

Every `fromNode` and `toNode` must name a node that exists. To edit a
canvas, read it with `vault.read`, change the JSON, and write it back with
`vault.update`; a canvas has no headings to patch.

## When Obsidian is closed

The `obsidian` MCP tools fail when Obsidian isn't running. Ask the user to
open the vault. If they can't, edit the files in `design/` directly and follow
the layout rules above by hand. Edit only the section that changes, as
`edit.patch` would. Use `grep -rn` in place of `graph.backlinks`. When you rename or move `X.md`, also rename or
move `X/`, and update the `[[links]]` to it yourself. When Obsidian next
opens, PageTree sees each moved file as new and puts it at the end of its
siblings.
