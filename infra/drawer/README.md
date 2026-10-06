# drawer

The repository's own plumbing, in Go: it keeps `wires.json` true, checks the
links in the docs, builds the GitHub Pages site and serves files for a
preview. It is not a wire itself, so it is not in the drawer.

## What you get

| Command | What it does |
|---|---|
| `drawer check` | Every project under `apps/`, `tools/` and `labs/` is in `wires.json`, and every entry is valid |
| `drawer add NAME PATH KIND [BLURB]` | Add an entry; `KIND` is `app`, `tool` or `lab` |
| `drawer docs` | Every relative link in every markdown file points to a file, and to a heading that exists |
| `drawer site [OUT]` | Build the site into `OUT`, by default `_site/` |
| `drawer serve [-port 8000] [DIR]` | Serve `DIR` at `http://127.0.0.1:PORT/`, with no cache |
| `drawer json FILE KEY.PATH` | Print a string from a JSON file; fail when it is missing or empty |

`check`, `add`, `docs` and `site` find the root of the repository as the
nearest folder with a `wires.json` above where they run; `-root DIR` sets it.

## Run it

The root tasks call it with `go run`, so there is nothing to install:

```sh
task index-check      # drawer check
task docs-check       # drawer docs
task site-serve       # drawer site, then drawer serve on _site
```

`drawer docs` prints each broken link with its file and line:

```
labs/acme-env/README.md:146: link to "../../docs/updates.md": no such file
labs/talos-cluster/README.md:141: link to "../../docs/updates.md": no such file
2 broken link(s) in 10 markdown files
```

It skips links inside code, and the folders `bin`, `.run`, `_site` and
`testdata`, which hold broken links on purpose. Anchors are compared the way
GitHub makes them from headings.

| Task | What it does |
|---|---|
| `task drawer:ci` | `go vet`, the format check and the tests |
| `task drawer:test` | The tests; one builds a site with real cross-compiles |

## How it is built

```
main.go         the command line
wires.go        wires.json: load, check, add; drawer json
docs.go         the link check
site.go         the site: copies apps and pages, builds tool binaries
serve.go        the file server
main_test.go    the tests, on a made-up repository in a temporary folder
```

`drawer site` builds a Go tool for Linux on amd64 and arm64, macOS on arm64
and amd64, and Windows on amd64, and writes `SHA256SUMS` beside them. It
publishes a tool's page and binaries, never its source. `drawer check`
fails for a wire without a page.
