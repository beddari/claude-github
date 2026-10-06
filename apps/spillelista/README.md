# Spillelista

A single-page Spotify playlist app with no build step: `index.html` loads
`config.json` at startup.

```sh
task spillelista:serve      # from the repo root, or: task serve in this folder
```

The app runs from a sub-path. It resolves its redirect URI, `config.json`
and history URL against its own directory, so `…/spillelista/` and
`…/spillelista/index.html` behave the same, and its sessionStorage keys
are prefixed `spillelista:` so other experiments on the same origin can't
clash with it. The redirect URI is that directory URL. Every
place the app is served from must be registered as a redirect URI in the
[Spotify dashboard](https://developer.spotify.com/dashboard), for example:

- `http://127.0.0.1:8000/` for `task spillelista:serve` (serves just this folder)
- `http://127.0.0.1:8000/apps/spillelista/` for the root `task serve`
- `https://<user>.github.io/claude-github/apps/spillelista/` on GitHub Pages
