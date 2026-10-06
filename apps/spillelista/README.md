# Spillelista

A one-page Spotify player: log in with Spotify, pick one of your playlists,
and play it in the browser. One `index.html` and a `config.json`, with no
build step.

## What you get

- Login with Spotify, in the browser, with no server (OAuth with PKCE).
- Your playlists, and playback through the Spotify Web Playback SDK.
- A page that works from any sub-path, as on GitHub Pages under
  `/claude-github/apps/spillelista/`.

Playback needs a Spotify Premium account; that is Spotify's rule for the
Web Playback SDK.

## Run it

```sh
task spillelista:serve        # http://127.0.0.1:8000/
```

Spotify accepts a login only for redirect URIs that are registered for the
app. The redirect URI is the page's own folder, so register each place the
app is served from in the
[Spotify dashboard](https://developer.spotify.com/dashboard):

| Served by | Redirect URI |
|---|---|
| `task spillelista:serve` | `http://127.0.0.1:8000/` |
| `task serve` at the root | `http://127.0.0.1:8000/apps/spillelista/` |
| GitHub Pages | `https://beddari.github.io/claude-github/apps/spillelista/` |

| Task | What it does |
|---|---|
| `task spillelista:serve` | Serve this folder at http://127.0.0.1:8000/; `PORT` changes the port |
| `task spillelista:ci` | Check that `config.json` parses and has a client ID |

## Settings

`config.json` holds the app's Spotify client ID under
`spotify.clients.spillelista.clientId`. A client ID is not a secret; it is
safe in the repository.

## How it is built

```
index.html      the whole app: markup, style and script
config.json     the Spotify client ID, read at start
Taskfile.yml    the tasks
```

The page finds everything relative to its own folder: the redirect URI,
`config.json` and the address it goes back to after login. So
`.../spillelista/` and `.../spillelista/index.html` behave the same. Its
sessionStorage keys start with `spillelista:`, so other pages on the same
site cannot clash with them.
