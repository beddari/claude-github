package audit

import (
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	mdLinkRe = regexp.MustCompile(`\[[^\]]*\]\(\s*<?([^)\s>]+)>?(?:\s+"[^"]*")?\s*\)`)
	codeSpan = regexp.MustCompile("`[^`]*`")
	barePath = regexp.MustCompile("(?:^|[\\s`'\"(\\[:=}/])((?:\\.{1,2}/)?[A-Za-z0-9_][A-Za-z0-9_./-]*\\.(?:md|markdown|py|sh|bash|js|mjs|cjs|ts|rb|txt|json|ya?ml|csv|xsd|html|toml))\\b")
)

// Refs holds the outgoing references from one document.
type Refs struct {
	Targets map[string]bool // existing files referenced, by Rel
	Broken  []string        // markdown links to local files that don't exist
}

// refs extracts references from doc to other files in the skill. Paths
// are resolved relative to the doc's directory first, then the skill root.
func (s *Skill) refs(d *Doc) Refs {
	r := Refs{Targets: map[string]bool{}}
	base := path.Dir(d.Rel)
	resolve := func(t string) (string, bool) {
		t = strings.TrimPrefix(t, "./")
		for _, cand := range []string{path.Join(base, t), path.Clean(t)} {
			if strings.HasPrefix(cand, "../") {
				continue
			}
			if s.Files[cand] && cand != d.Rel {
				return cand, true
			}
			// Case-insensitive fallback: Claude will find README.md for readme.md.
			for f := range s.Files {
				if strings.EqualFold(f, cand) && f != d.Rel {
					return f, true
				}
			}
		}
		return "", false
	}
	for i, l := range d.Lines {
		if !d.InCode[i] {
			for _, m := range mdLinkRe.FindAllStringSubmatch(codeSpan.ReplaceAllString(l, ""), -1) {
				t := m[1]
				if strings.Contains(t, "://") || strings.HasPrefix(t, "#") ||
					strings.HasPrefix(t, "mailto:") || strings.HasPrefix(t, "/") {
					continue
				}
				if j := strings.IndexByte(t, '#'); j >= 0 {
					t = t[:j]
				}
				if u, err := url.PathUnescape(t); err == nil {
					t = u
				}
				if t == "" {
					continue
				}
				if f, ok := resolve(t); ok {
					r.Targets[f] = true
				} else if strings.Contains(t, "/") || path.Ext(t) != "" {
					// Extension-less, slash-less targets are placeholders like (URL).
					r.Broken = append(r.Broken, t)
				}
			}
		}
		for _, m := range barePath.FindAllStringSubmatch(l, -1) {
			if f, ok := resolve(m[1]); ok {
				r.Targets[f] = true
			}
		}
	}
	sort.Strings(r.Broken)
	return r
}

// depthMap runs a BFS over markdown-to-markdown references from SKILL.md.
// It returns each reachable doc's depth and the doc that first reached it.
func (s *Skill) depthMap() (depth map[string]int, via map[string]string) {
	depth = map[string]int{s.Main.Rel: 0}
	via = map[string]string{}
	queue := []string{s.Main.Rel}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		targets := sortedKeys(s.refs(s.Docs[cur]).Targets)
		for _, t := range targets {
			if _, isDoc := s.Docs[t]; !isDoc {
				continue
			}
			if _, seen := depth[t]; seen {
				continue
			}
			depth[t] = depth[cur] + 1
			via[t] = cur
			queue = append(queue, t)
		}
	}
	return depth, via
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
