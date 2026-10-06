package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Index is wires.json.
type Index struct {
	Repo  string `json:"repo"`
	Wires []Wire `json:"wires"`
}

// Wire is one experiment in the drawer.
type Wire struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Kind  string `json:"kind"`
	Blurb string `json:"blurb"`
	Href  string `json:"href,omitempty"` // relative page to publish, e.g. apps/x/
}

var nameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func load(root string) (*Index, error) {
	b, err := os.ReadFile(filepath.Join(root, "wires.json"))
	if err != nil {
		return nil, err
	}
	var ix Index
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ix); err != nil {
		return nil, fmt.Errorf("wires.json: %w", err)
	}
	return &ix, nil
}

func save(root string, ix *Index) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(ix); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "wires.json"), buf.Bytes(), 0o644)
}

// projects lists apps/*, tools/* and labs/* directories that have a Taskfile.
func projects(root string) ([]string, error) {
	var out []string
	for _, parent := range []string{"apps", "tools", "labs"} {
		entries, err := os.ReadDir(filepath.Join(root, parent))
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			rel := parent + "/" + e.Name()
			if _, err := os.Stat(filepath.Join(root, rel, "Taskfile.yml")); err == nil {
				out = append(out, rel)
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

func validKind(k string) bool { return k == "app" || k == "tool" || k == "lab" }

func isLocal(href string) bool { return href != "" && !strings.Contains(href, "://") }

// check returns the number of wires and every problem found.
func check(root string) (int, []string, error) {
	ix, err := load(root)
	if err != nil {
		return 0, nil, err
	}
	var errs []string
	listed := map[string]bool{}
	for _, w := range ix.Wires {
		listed[w.Path] = true
		switch {
		case !nameRe.MatchString(w.Name):
			errs = append(errs, fmt.Sprintf("entry %q: name must be lowercase letters, digits and hyphens", w.Name))
		case w.Path == "" || w.Blurb == "":
			errs = append(errs, fmt.Sprintf("entry %q: path and blurb are required", w.Name))
		case !validKind(w.Kind):
			errs = append(errs, fmt.Sprintf("entry %q: kind must be app, tool or lab, got %q", w.Name, w.Kind))
		}
		if fi, err := os.Stat(filepath.Join(root, w.Path)); w.Path != "" && (err != nil || !fi.IsDir()) {
			errs = append(errs, fmt.Sprintf("%s is listed but does not exist", w.Path))
		}
		switch {
		case w.Href == "":
			errs = append(errs, fmt.Sprintf("entry %q: no href; every wire has a page", w.Name))
		case isLocal(w.Href):
			if _, err := os.Stat(filepath.Join(root, w.Href, "index.html")); err != nil {
				errs = append(errs, fmt.Sprintf("entry %q: href %q has no index.html to publish", w.Name, w.Href))
			}
		}
	}
	ps, err := projects(root)
	if err != nil {
		return 0, nil, err
	}
	for _, p := range ps {
		if !listed[p] {
			errs = append(errs, fmt.Sprintf("%s is not in the drawer (task wire-add -- NAME %s KIND BLURB)", p, p))
		}
	}
	return len(ix.Wires), errs, nil
}

// add appends w unless its path is already listed. Its href defaults to its
// path: the folder whose index.html is the wire's page.
func add(root string, w Wire) (bool, error) {
	if !nameRe.MatchString(w.Name) {
		return false, fmt.Errorf("name %q must be lowercase letters, digits and hyphens", w.Name)
	}
	if !validKind(w.Kind) {
		return false, fmt.Errorf("kind must be app, tool or lab, got %q", w.Kind)
	}
	ix, err := load(root)
	if err != nil {
		return false, err
	}
	for _, x := range ix.Wires {
		if x.Path == w.Path {
			return false, nil
		}
	}
	if w.Href == "" {
		w.Href = strings.TrimSuffix(w.Path, "/") + "/"
	}
	ix.Wires = append(ix.Wires, w)
	return true, save(root, ix)
}

// jsonString reads a dotted key path from a JSON file and requires a
// non-empty string there.
func jsonString(file, keyPath string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return "", fmt.Errorf("%s: %w", file, err)
	}
	for _, k := range strings.Split(keyPath, ".") {
		m, ok := v.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s: %s: not an object at %q", file, keyPath, k)
		}
		if v, ok = m[k]; !ok {
			return "", fmt.Errorf("%s: %s: missing key %q", file, keyPath, k)
		}
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s: %s must be a non-empty string", file, keyPath)
	}
	return s, nil
}
