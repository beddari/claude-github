package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// targets are the platforms Go tools with a page are cross-compiled for.
var targets = [][2]string{
	{"linux", "amd64"}, {"linux", "arm64"}, {"darwin", "arm64"}, {"darwin", "amd64"}, {"windows", "amd64"},
}

// skipInApps are not published when an app folder is copied.
var skipInApps = map[string]bool{"bin": true, ".task": true, "Taskfile.yml": true, "README.md": true}

// buildSite assembles the GitHub Pages site:
//   - index.html, wires.json and assets/ at the root: the landing page
//   - apps/*: the whole folder, its page included
//   - tools/* and labs/*: the index.html of the href; a Go tool also gets
//     binaries and SHA256SUMS in dl/
func buildSite(root, out string, log io.Writer) error {
	ix, err := load(root)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, f := range []string{"index.html", "wires.json"} {
		if err := copyFile(filepath.Join(root, f), filepath.Join(out, f)); err != nil {
			return err
		}
	}
	// Shared static files for the landing page (e.g. the background photo).
	if fi, err := os.Stat(filepath.Join(root, "assets")); err == nil && fi.IsDir() {
		if err := copyTree(filepath.Join(root, "assets"), filepath.Join(out, "assets")); err != nil {
			return err
		}
	}
	// Serve files as-is; no Jekyll processing.
	if err := os.WriteFile(filepath.Join(out, ".nojekyll"), nil, 0o644); err != nil {
		return err
	}
	for _, w := range ix.Wires {
		if !isLocal(w.Href) {
			continue
		}
		src, dst := filepath.Join(root, w.Path), filepath.Join(out, filepath.FromSlash(w.Href))
		fmt.Fprintf(log, "wire %s -> %s\n", w.Name, w.Href)
		if strings.HasPrefix(w.Path, "apps/") {
			if err := copyTree(src, filepath.Join(out, filepath.FromSlash(w.Path))); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(filepath.Join(root, filepath.FromSlash(w.Href), "index.html"), filepath.Join(dst, "index.html")); err != nil {
			return err
		}
		if _, err := os.Stat(filepath.Join(src, "go.mod")); err == nil {
			if err := buildGo(src, w.Name, filepath.Join(dst, "dl"), log); err != nil {
				return err
			}
		}
	}
	fmt.Fprintf(log, "site ready in %s\n", out)
	return nil
}

func buildGo(src, name, dl string, log io.Writer) error {
	if err := os.MkdirAll(dl, 0o755); err != nil {
		return err
	}
	var sums strings.Builder
	for _, t := range targets {
		bin := fmt.Sprintf("%s-%s-%s", name, t[0], t[1])
		if t[0] == "windows" {
			bin += ".exe"
		}
		outPath, err := filepath.Abs(filepath.Join(dl, bin))
		if err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", outPath, ".")
		cmd.Dir = src
		cmd.Env = append(os.Environ(), "GOOS="+t[0], "GOARCH="+t[1], "CGO_ENABLED=0")
		cmd.Stdout, cmd.Stderr = log, log
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build %s: %w", bin, err)
		}
		b, err := os.ReadFile(outPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(b), bin)
		fmt.Fprintf(log, "  built %s\n", path.Join(filepath.Base(filepath.Dir(dl)), "dl", bin))
	}
	return os.WriteFile(filepath.Join(dl, "SHA256SUMS"), []byte(sums.String()), 0o644)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if skipInApps[d.Name()] && p != src {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		return copyFile(p, filepath.Join(dst, rel))
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
