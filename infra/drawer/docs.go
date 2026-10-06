package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Folders never scanned for markdown: build output, caches, and test
// fixtures that hold broken links on purpose.
var skipDocDirs = map[string]bool{
	".git": true, "node_modules": true, "bin": true, ".run": true, ".task": true,
	"_site": true, "testdata": true, "prompts": true,
}

var (
	mdLink     = regexp.MustCompile(`\]\(\s*<?([^)\s>]+)>?(?:\s+"[^"]*")?\s*\)`)
	mdRefDef   = regexp.MustCompile(`^\s*\[[^\]]+\]:\s*(\S+)`)
	inlineCode = regexp.MustCompile("`[^`]*`")
	fenceLine  = regexp.MustCompile("^\\s*(```|~~~)")
	headingRe  = regexp.MustCompile(`^#{1,6}\s+(.+?)\s*#*\s*$`)
	slugDrop   = regexp.MustCompile(`[^\p{L}\p{N}\- ]`)
)

// checkDocs returns every broken relative link in the markdown files under
// root, as "file:line: message", and how many files it read.
func checkDocs(root string) ([]string, int, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && p != root && skipDocDirs[d.Name()] {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(p), ".md") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Strings(files)
	anchors := map[string]map[string]bool{}
	var problems []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, 0, err
		}
		rel, _ := filepath.Rel(root, f)
		inFence := false
		for i, line := range strings.Split(string(b), "\n") {
			if fenceLine.MatchString(line) {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}
			line = inlineCode.ReplaceAllString(line, "")
			var targets []string
			for _, m := range mdLink.FindAllStringSubmatch(line, -1) {
				targets = append(targets, m[1])
			}
			if m := mdRefDef.FindStringSubmatch(line); m != nil {
				targets = append(targets, m[1])
			}
			for _, t := range targets {
				if msg := checkLink(f, t, anchors); msg != "" {
					problems = append(problems, fmt.Sprintf("%s:%d: %s", filepath.ToSlash(rel), i+1, msg))
				}
			}
		}
	}
	return problems, len(files), nil
}

// checkLink returns "" when target is fine or not local, else a message.
func checkLink(from, target string, anchors map[string]map[string]bool) string {
	if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "/") {
		return ""
	}
	path, frag, _ := strings.Cut(target, "#")
	if p, err := url.PathUnescape(path); err == nil {
		path = p
	}
	file := from
	if path != "" {
		file = filepath.Join(filepath.Dir(from), filepath.FromSlash(path))
		if _, err := os.Stat(file); err != nil {
			return fmt.Sprintf("link to %q: no such file", target)
		}
	}
	if frag == "" || !strings.EqualFold(filepath.Ext(file), ".md") {
		return ""
	}
	set, ok := anchors[file]
	if !ok {
		set = headingAnchors(file)
		anchors[file] = set
	}
	if !set[strings.ToLower(frag)] {
		return fmt.Sprintf("link to %q: no heading with that anchor", target)
	}
	return ""
}

// headingAnchors lists the anchors GitHub makes for a file's headings:
// lower case, punctuation dropped, spaces to dashes, -1, -2 for repeats.
func headingAnchors(file string) map[string]bool {
	set := map[string]bool{}
	b, err := os.ReadFile(file)
	if err != nil {
		return set
	}
	seen := map[string]int{}
	inFence := false
	for _, line := range strings.Split(string(b), "\n") {
		if fenceLine.MatchString(line) {
			inFence = !inFence
			continue
		}
		m := headingRe.FindStringSubmatch(line)
		if inFence || m == nil {
			continue
		}
		text := strings.NewReplacer("`", "", "*", "", "_", "").Replace(m[1])
		slug := strings.ReplaceAll(slugDrop.ReplaceAllString(strings.ToLower(text), ""), " ", "-")
		if n := seen[slug]; n > 0 {
			set[fmt.Sprintf("%s-%d", slug, n)] = true
		} else {
			set[slug] = true
		}
		seen[slug]++
	}
	return set
}
