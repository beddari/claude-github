package audit

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func load(t *testing.T, name string) *Skill {
	t.Helper()
	s, err := Load(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func statuses(r Report) map[string]Status {
	m := map[string]Status{}
	for _, res := range r.Results {
		m[res.ID] = res.Status
	}
	return m
}

func TestGoodSkill(t *testing.T) {
	rep := Run(load(t, "good-skill"), DefaultOptions())
	for _, res := range rep.Results {
		if res.Status >= Warn {
			t.Errorf("%s: got %s (%s) %v", res.ID, res.Status, res.Summary, res.Details)
		}
	}
	got := statuses(rep)
	for id, want := range map[string]Status{
		"frontmatter": Pass, "contents": Pass, "nesting": Pass, "links": Pass,
		"freedom": Pass, "models": Pass, "checklist": Pass, "feedback": Pass, "deps": Pass,
	} {
		if got[id] != want {
			t.Errorf("%s = %s, want %s", id, got[id], want)
		}
	}
}

func TestBadSkill(t *testing.T) {
	rep := Run(load(t, "bad-skill"), DefaultOptions())
	got := statuses(rep)
	want := map[string]Status{
		"frontmatter": Fail, // invalid name
		"contents":    Fail, // details.md 140+ lines, no TOC
		"nesting":     Fail, // details.md only via advanced.md
		"links":       Fail, // missing.md
		"freedom":     Warn, // production deletes, refunds, no script
		"models":      Warn,
		"emphasis":    Warn,
		"checklist":   Warn,
		"feedback":    Warn,
		"deps":        Fail, // pandas, jq
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("%s = %s, want %s", id, got[id], w)
		}
	}
	find := func(id string) Result {
		for _, r := range rep.Results {
			if r.ID == id {
				return r
			}
		}
		t.Fatalf("no result %s", id)
		return Result{}
	}
	if d := strings.Join(find("nesting").Details, "\n"); !strings.Contains(d, "SKILL.md → advanced.md → details.md") || !strings.Contains(d, "orphan.md") {
		t.Errorf("nesting details missing chain/orphan:\n%s", d)
	}
	if d := strings.Join(find("contents").Details, "\n"); !strings.Contains(d, "after line 100") {
		t.Errorf("contents should flag late rules: %s", d)
	}
	if d := strings.Join(find("deps").Details, "\n"); !strings.Contains(d, "pandas") || !strings.Contains(d, "jq") {
		t.Errorf("deps details: %s", d)
	}
}

func TestLengthLimit(t *testing.T) {
	opt := DefaultOptions()
	opt.MaxSkillLines = 10
	if st := statuses(Run(load(t, "good-skill"), opt))["length"]; st != Fail {
		t.Errorf("length = %s, want fail", st)
	}
}

func TestOnly(t *testing.T) {
	opt := DefaultOptions()
	opt.Only = map[string]bool{"deps": true}
	rep := Run(load(t, "bad-skill"), opt)
	if len(rep.Results) != 1 || rep.Results[0].ID != "deps" {
		t.Errorf("got %+v", rep.Results)
	}
}

func TestDiscover(t *testing.T) {
	dirs, err := Discover([]string{"testdata"})
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 2 {
		t.Fatalf("got %v", dirs)
	}
	one, err := Discover([]string{filepath.Join("testdata", "good-skill", "SKILL.md")})
	if err != nil || len(one) != 1 {
		t.Fatalf("SKILL.md path: %v %v", one, err)
	}
}

func TestFrontmatter(t *testing.T) {
	fm := parseFrontmatter(strings.Split(`---
name: x
description: >
  Folded text
  over lines.
metadata:
  tested-models:
    - haiku
    - opus
---
body`, "\n"))
	if fm.Fields["description"] != "Folded text over lines." {
		t.Errorf("description = %q", fm.Fields["description"])
	}
	if fm.Fields["metadata.tested-models"] != "haiku, opus" {
		t.Errorf("models = %q", fm.Fields["metadata.tested-models"])
	}
	if fm.BodyStart != 10 {
		t.Errorf("BodyStart = %d", fm.BodyStart)
	}
}

func TestFindTOC(t *testing.T) {
	d := parseDoc("x.md", "# T\n\n**Contents**\n- [Auth and setup](#auth-and-setup)\n- Core methods\n\n## Auth and setup\n## Core methods\n")
	toc, ok := findTOC(d, 100)
	if !ok {
		t.Fatal("no toc")
	}
	for _, h := range sectionHeadings(d, toc) {
		if !strings.Contains(toc, norm(h)) {
			t.Errorf("heading %q not covered by %q", h, toc)
		}
	}
	if _, ok := findTOC(parseDoc("y.md", "# T\n\n## A\ntext\n## B\n"), 100); ok {
		t.Error("false positive toc")
	}
}

func TestNegatedFragile(t *testing.T) {
	if lastWords("Never", 4) != "Never" || !negatedRe.MatchString(lastWords("you should never", 4)) {
		t.Error("negation helper")
	}
}

func TestPromptAndRenderers(t *testing.T) {
	reps := []Report{Run(load(t, "bad-skill"), DefaultOptions()), Run(load(t, "good-skill"), DefaultOptions())}
	p := CombinedPrompt(reps, false)
	for _, want := range []string{"Bad_Skill", "before changing anything", "SKILL.md → advanced.md → details.md", "Model fit"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
	if strings.Contains(p, "`good-skill` skill") {
		t.Error("prompt should skip clean skills")
	}
	var buf bytes.Buffer
	WriteText(&buf, reps, TextOptions{})
	WriteMarkdown(&buf, reps)
	if err := WriteJSON(&buf, reps); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"status": "fail"`) {
		t.Error("json should render status names")
	}
}
