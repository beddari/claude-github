package audit

import "strings"

// Frontmatter is a deliberately small YAML subset parser: top-level and
// one-level nested scalars, block scalars (| and >) and simple lists.
// Nested keys are flattened as "parent.child".
type Frontmatter struct {
	Present   bool
	Fields    map[string]string
	BodyStart int // index of the first body line
}

func parseFrontmatter(lines []string) Frontmatter {
	fm := Frontmatter{Fields: map[string]string{}}
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return fm
	}
	fm.Present = true
	fm.BodyStart = end + 1

	var parent, lastKey, blockKey string
	blockIndent := -1
	var block []string
	flush := func() {
		if blockKey != "" {
			fm.Fields[blockKey] = strings.TrimSpace(strings.Join(block, " "))
			blockKey, block = "", nil
		}
	}
	for _, l := range lines[1:end] {
		t := strings.TrimSpace(l)
		indent := len(l) - len(strings.TrimLeft(l, " \t"))
		if blockKey != "" {
			if t == "" || indent > blockIndent {
				block = append(block, t)
				continue
			}
			flush()
		}
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if strings.HasPrefix(t, "- ") {
			item := unquote(strings.TrimSpace(t[2:]))
			if lastKey != "" {
				if fm.Fields[lastKey] != "" {
					fm.Fields[lastKey] += ", "
				}
				fm.Fields[lastKey] += item
			}
			continue
		}
		k, v, ok := strings.Cut(t, ":")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		full := k
		if indent == 0 {
			parent = ""
		} else if parent != "" {
			full = parent + "." + k
		}
		lastKey = full
		switch {
		case v == "":
			if indent == 0 {
				parent = k
			}
			fm.Fields[full] = ""
		case strings.HasPrefix(v, "|") || strings.HasPrefix(v, ">"):
			blockKey, blockIndent = full, indent
		default:
			fm.Fields[full] = unquote(v)
		}
	}
	flush()
	return fm
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}
