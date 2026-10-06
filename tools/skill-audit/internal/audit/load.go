package audit

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Doc is a parsed markdown file inside a skill.
type Doc struct {
	Rel      string // path relative to the skill root, slash separated
	Lines    []string
	InCode   []bool   // line is inside (or is) a fenced code block
	CodeLang []string // fence language for code lines
	Headings []Heading
}

// Heading is a markdown ATX heading outside code fences.
type Heading struct {
	Level int
	Text  string
	Line  int // 0-based
}

// Skill is a loaded skill directory.
type Skill struct {
	Name    string
	Dir     string
	Main    *Doc // SKILL.md
	FM      Frontmatter
	Docs    map[string]*Doc // all markdown files, keyed by Rel
	Files   map[string]bool // every file, keyed by Rel
	Scripts []string        // executable code files, keyed by Rel
	Sources map[string]string
}

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "__pycache__": true,
	".venv": true, "venv": true, ".mypy_cache": true, ".pytest_cache": true,
}

var scriptExt = map[string]bool{
	".py": true, ".sh": true, ".bash": true, ".js": true, ".mjs": true,
	".cjs": true, ".ts": true, ".rb": true,
}

const maxRead = 2 << 20

// Discover finds every directory under the given roots that contains a
// SKILL.md (case-insensitive). Results are sorted and de-duplicated.
func Discover(roots []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if isSkillFile(filepath.Base(root)) {
				root = filepath.Dir(root)
			} else {
				return nil, errors.New(root + ": not a directory or SKILL.md")
			}
		}
		err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // unreadable entries are skipped, not fatal
			}
			if d.IsDir() && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if !d.IsDir() && isSkillFile(d.Name()) {
				dir, _ := filepath.Abs(filepath.Dir(p))
				if real, err := filepath.EvalSymlinks(dir); err == nil {
					dir = real
				}
				if !seen[dir] {
					seen[dir] = true
					out = append(out, dir)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

func isSkillFile(name string) bool { return strings.EqualFold(name, "SKILL.md") }

// Load reads a skill directory.
func Load(dir string) (*Skill, error) {
	s := &Skill{
		Dir:     dir,
		Docs:    map[string]*Doc{},
		Files:   map[string]bool{},
		Sources: map[string]string{},
	}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p != dir && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		s.Files[rel] = true
		ext := strings.ToLower(filepath.Ext(p))
		isMD := ext == ".md" || ext == ".markdown"
		isScript := scriptExt[ext]
		if !isMD && !isScript && !isManifest(d.Name()) {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > maxRead {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		s.Sources[rel] = string(b)
		if isMD {
			s.Docs[rel] = parseDoc(rel, string(b))
			if isSkillFile(d.Name()) && !strings.Contains(rel, "/") {
				s.Main = s.Docs[rel]
			}
		}
		if isScript {
			s.Scripts = append(s.Scripts, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.Main == nil {
		return nil, errors.New(dir + ": no SKILL.md at skill root")
	}
	sort.Strings(s.Scripts)
	s.FM = parseFrontmatter(s.Main.Lines)
	s.Name = s.FM.Fields["name"]
	if s.Name == "" {
		s.Name = filepath.Base(dir)
	}
	return s, nil
}

func isManifest(name string) bool {
	switch strings.ToLower(name) {
	case "requirements.txt", "requirements-dev.txt", "pyproject.toml",
		"package.json", "pipfile", "go.mod", "gemfile", "cargo.toml":
		return true
	}
	return false
}

var (
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*#*\s*$`)
	fenceRe   = regexp.MustCompile("^\\s*(```+|~~~+)\\s*([\\w+-]*)")
)

func parseDoc(rel, text string) *Doc {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	d := &Doc{Rel: rel, Lines: lines, InCode: make([]bool, len(lines)), CodeLang: make([]string, len(lines))}
	fence, lang := "", ""
	fmEnd := -1
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				fmEnd = i
				break
			}
		}
	}
	for i, l := range lines {
		if i <= fmEnd {
			d.InCode[i] = true // frontmatter is not prose
			continue
		}
		if m := fenceRe.FindStringSubmatch(l); m != nil {
			marker := m[1][:3]
			if fence == "" {
				fence, lang = marker, strings.ToLower(m[2])
				d.InCode[i], d.CodeLang[i] = true, lang
				continue
			}
			if marker == fence && strings.TrimSpace(l) == strings.TrimSpace(m[1]) {
				d.InCode[i], d.CodeLang[i] = true, lang
				fence, lang = "", ""
				continue
			}
		}
		if fence != "" {
			d.InCode[i], d.CodeLang[i] = true, lang
			continue
		}
		if m := headingRe.FindStringSubmatch(l); m != nil {
			d.Headings = append(d.Headings, Heading{Level: len(m[1]), Text: m[2], Line: i})
		}
	}
	return d
}

// BodyLines returns SKILL.md lines after the frontmatter.
func (s *Skill) BodyLines() []string { return s.Main.Lines[s.FM.BodyStart:] }

// SortedDocs returns markdown docs other than SKILL.md, sorted by path.
func (s *Skill) SortedDocs() []*Doc {
	var out []*Doc
	for _, d := range s.Docs {
		if d != s.Main {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Rel < out[j].Rel })
	return out
}
