# Writing docs

How the docs in this repository are written. The model is
[incusdev-vm](https://git.dataverket.org/dataverket/incusdev-vm): its README
and its four documents under `docs/`. When in doubt, read those and do the
same.

## The reader

Someone who has the repository and wants to run the thing, use it, or find
out why it broke. They may not have English as their first language. They
read the README first and a document under `docs/` only when they need it.

## The words

- Short sentences in plain English. One idea per sentence.
- Say what is, in the present tense: "`task up` starts the VM", not "will
  start" or "should start".
- Concrete facts over adjectives: times ("about 6 minutes"), sizes
  ("2 GiB"), ports, paths, versions. No "simple", "easy", "powerful",
  "seamless" or "just".
- Name a thing the same way every time. Explain a term once, in brackets,
  where it first appears: "the object gateway, RadosGW".
- "You" for the reader. Tell them what to do, then what they will see.
- Say why when a choice is not obvious. A reader who knows why can tell
  when it no longer holds.
- Say what is not tested, and what does not work. Never claim more than was
  run.
- No emoji, no exclamation marks, no marketing. Bold only for the lead-in of
  a paragraph that explains one item, as in the lists below.

## A README

A README answers, in this order:

1. **What it is**, in one paragraph: what it does and what it runs on.
2. **What you get**: a short list, and a table of the services with port and
   login, when there are services.
3. **How to run it**: what you need first (tools, ports, CPU, memory), then
   the commands, each with a comment saying what it does and how long it
   takes. Show the output where it helps the reader know it worked.
4. **The tasks**: a table of `task` names and what each one does.
5. **Use it**: the first things to do once it runs.
6. **Settings**: a table of variables with their default and meaning.
7. **How it is built**: the file tree, one line per file or folder.
8. **The documents**: a table linking each file under `docs/` with one line
   on what is in it.

Leave out a section that has nothing to say. Keep the README to what most
readers need; move details to `docs/`.

## The documents

A project with more than one moving part has a `docs/` folder with these
files. Each starts with one or two sentences on what it covers.

| File | What it holds |
|---|---|
| `architecture.md` | A mermaid diagram of the parts and how traffic flows, then tables of networks and ports, then the flows, one paragraph each with a bold lead-in |
| `access.md` | How to log in to or reach each service, one section per service |
| `troubleshooting.md` | "Where to look" (a table of question and command), "Known behaviour", "Why it is built this way", and "What is tested" with what is not |
| `updates.md` | Which versions are pinned, where, what moves them (Renovate) and how to check a change before merging it |

A small project can do with its README alone. The README of the repository
keeps `docs/` for what is common to all projects.

## Tables, code and diagrams

- A table when there are three or more items with the same fields: tasks,
  ports, settings, versions. Otherwise a list or a sentence.
- Every command in a code block can be copied and run as it stands. Put the
  explanation in a `#` comment at the end of the line, or in the sentence
  before the block.
- Show real output from a real run, trimmed, never invented.
- Diagrams in mermaid, so they live in the text and render on GitHub. Label
  every arrow with what flows over it and on which port.

## Keep docs true

- A change to how something runs changes its docs in the same commit.
- `task ci` checks that every relative link in every markdown file points to
  a file that exists.
- Versions mentioned in docs are listed in `renovate.json`, so an update
  changes the code and the docs together.
