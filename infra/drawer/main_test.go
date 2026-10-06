package main

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// repo builds a fake monorepo: one app, one Go tool with a page, wires.json.
func repo(t *testing.T) string {
	t.Helper()
	r := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(r, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("index.html", "<h1>drawer</h1>")
	write("assets/bg.webp", "img")
	write("wires.json", `{"repo":"https://example.com/r","wires":[
  {"name":"app-one","path":"apps/app-one","kind":"app","blurb":"An app.","href":"apps/app-one/"},
  {"name":"tool-one","path":"tools/tool-one","kind":"tool","blurb":"A tool.","href":"tools/tool-one/"}]}`)
	write("apps/app-one/Taskfile.yml", "version: '3'\n")
	write("apps/app-one/index.html", "app")
	write("apps/app-one/config.json", `{"a":{"b":"x"}}`)
	write("apps/app-one/README.md", "not published")
	write("tools/tool-one/Taskfile.yml", "version: '3'\n")
	write("tools/tool-one/index.html", "tool page")
	write("tools/tool-one/go.mod", "module example.com/toolone\n\ngo 1.24\n")
	write("tools/tool-one/main.go", "package main\n\nfunc main() {}\n")
	return r
}

func runOK(t *testing.T, args ...string) string {
	t.Helper()
	var out, errb bytes.Buffer
	if code := run(args, &out, &errb); code != 0 {
		t.Fatalf("drawer %v: exit %d\n%s%s", args, code, out.String(), errb.String())
	}
	return out.String()
}

func TestCheckAndAdd(t *testing.T) {
	r := repo(t)
	if out := runOK(t, "-root", r, "check"); !strings.Contains(out, "OK (2 wires)") {
		t.Errorf("check: %s", out)
	}

	// A new project without an entry fails the check…
	if err := os.MkdirAll(filepath.Join(r, "tools/new-thing"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(r, "tools/new-thing/Taskfile.yml"), []byte("version: '3'\n"), 0o644)
	var out, errb bytes.Buffer
	if code := run([]string{"-root", r, "check"}, &out, &errb); code != 1 || !strings.Contains(errb.String(), "tools/new-thing is not in the drawer") {
		t.Errorf("check with unlisted project: %d %s", code, errb.String())
	}
	// …until it's added; adding twice is a no-op.
	runOK(t, "-root", r, "add", "new-thing", "tools/new-thing", "tool", "Shiny.")
	if out := runOK(t, "-root", r, "add", "new-thing", "tools/new-thing", "tool"); !strings.Contains(out, "already listed") {
		t.Errorf("second add: %s", out)
	}
	runOK(t, "-root", r, "check")

	if code := run([]string{"-root", r, "add", "Bad Name", "x", "tool"}, &out, &errb); code != 1 {
		t.Errorf("bad name accepted")
	}
	if code := run([]string{"-root", r, "add", "x", "tools/x", "gizmo"}, &out, &errb); code != 1 {
		t.Errorf("bad kind accepted")
	}
}

func TestCheckBadHref(t *testing.T) {
	r := repo(t)
	os.Remove(filepath.Join(r, "tools/tool-one/index.html"))
	var out, errb bytes.Buffer
	if code := run([]string{"-root", r, "check"}, &out, &errb); code != 1 || !strings.Contains(errb.String(), "no index.html") {
		t.Errorf("missing page not caught: %d %s", code, errb.String())
	}
}

func TestFindRoot(t *testing.T) {
	r := repo(t)
	got, err := findRoot(filepath.Join(r, "apps/app-one"))
	if err != nil || got != r {
		t.Errorf("findRoot = %q, %v", got, err)
	}
	if _, err := findRoot(t.TempDir()); err == nil {
		t.Error("expected error outside a repo")
	}
}

func TestJSON(t *testing.T) {
	r := repo(t)
	f := filepath.Join(r, "apps/app-one/config.json")
	if out := runOK(t, "json", f, "a.b"); strings.TrimSpace(out) != "x" {
		t.Errorf("json = %q", out)
	}
	var out, errb bytes.Buffer
	for _, key := range []string{"a.c", "a", "a.b.c"} {
		if code := run([]string{"json", f, key}, &out, &errb); code != 1 {
			t.Errorf("json %s: exit %d", key, code)
		}
	}
}

func TestSite(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-compiles")
	}
	r := repo(t)
	out := filepath.Join(t.TempDir(), "site")
	runOK(t, "-root", r, "site", out)
	for _, f := range []string{
		"index.html", "wires.json", ".nojekyll", "assets/bg.webp",
		"apps/app-one/index.html", "apps/app-one/config.json",
		"tools/tool-one/index.html", "tools/tool-one/dl/SHA256SUMS",
		"tools/tool-one/dl/tool-one-darwin-arm64", "tools/tool-one/dl/tool-one-windows-amd64.exe",
	} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	for _, f := range []string{"apps/app-one/README.md", "apps/app-one/Taskfile.yml", "tools/tool-one/main.go"} {
		if _, err := os.Stat(filepath.Join(out, f)); err == nil {
			t.Errorf("%s should not be published", f)
		}
	}
	sums, _ := os.ReadFile(filepath.Join(out, "tools/tool-one/dl/SHA256SUMS"))
	if n := strings.Count(string(sums), "\n"); n != len(targets) {
		t.Errorf("SHA256SUMS has %d lines, want %d", n, len(targets))
	}
}

func TestServe(t *testing.T) {
	r := repo(t)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	var out syncBuf
	go run([]string{"serve", "-port", itoa(port), r}, &out, &out)
	url := "http://127.0.0.1:" + itoa(port) + "/apps/app-one/config.json"
	var resp *http.Response
	for i := 0; i < 50; i++ {
		if resp, err = http.Get(url); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status %d", resp.StatusCode)
	}
}
