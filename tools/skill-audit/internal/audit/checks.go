package audit

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// CheckInfo describes one rule from the guidance.
type CheckInfo struct {
	ID    string
	Title string
	Rule  string
	run   func(*Skill, Options) Result
}

// Checks lists every rule in the order they are reported.
var Checks = []CheckInfo{
	{"frontmatter", "Name & description", "Frontmatter has a valid `name` (lowercase, hyphens, ≤64 chars, no reserved words) and a third-person `description` (≤1024 chars) saying what the skill does and when to use it.", checkFrontmatter},
	{"length", "SKILL.md size", "Keep the SKILL.md body under 500 lines and treat it as the contents page for everything else; split into reference files as you approach the limit.", checkLength},
	{"contents", "Contents list on long references", "Claude previews reference files with `head -100`. Any reference file over 100 lines needs a contents list at the top that matches its headings, so Claude can read it all or jump to the section it needs.", checkContents},
	{"nesting", "References one level deep", "Every reference file is linked directly from SKILL.md. Files only reachable through another reference get partially read (previewed) at best.", checkNesting},
	{"links", "Broken references", "Every local file SKILL.md or a reference links to exists.", checkLinks},
	{"freedom", "Degrees of freedom", "Match detail to fragility: high freedom (plain goals) for flexible work, templates for medium, and exact scripts for fragile steps (money, deletion, migrations, deploys). Low freedom usually means a script, not more prose.", checkFreedom},
	{"models", "Tested models declared", "Test the skill on every model you'll run it with (Haiku: enough guidance? Sonnet: clear and efficient? Opus: not over-explained?) and record the intended/tested models in the frontmatter.", checkModels},
	{"emphasis", "Not over-prescriptive", "Newer reasoning models do worse with shouty, over-prescriptive instructions written for older models (ALL-CAPS MUST/NEVER/IMPORTANT). Remove instructions the model does better without.", checkEmphasis},
	{"checklist", "Checklist for ordered workflows", "For multi-step work where order matters, give Claude a short checklist it copies into its reply and ticks off, with a short paragraph per step.", checkChecklist},
	{"feedback", "Self-correction loop", "Run a check (script or style guide), fix the errors, and repeat until it passes — e.g. 'if citations are incomplete, return to step 3'. Have Claude propose new rules for the guide at the end of a run.", checkFeedback},
	{"deps", "Explicit dependencies", "Don't assume tools are installed. Every package or CLI the skill's scripts and code examples use has an install line (pip/npm/apt/…) next to it, so the skill works on a teammate's machine on day one.", checkDeps},
}

// Run executes all (or the selected) checks against a skill.
func Run(s *Skill, opt Options) Report {
	rep := Report{Name: s.Name, Dir: s.Dir}
	for _, c := range Checks {
		if len(opt.Only) > 0 && !opt.Only[c.ID] {
			continue
		}
		r := c.run(s, opt)
		r.ID, r.Title = c.ID, c.Title
		rep.Results = append(rep.Results, r)
	}
	return rep
}

// LookupCheck returns the check with the given ID.
func LookupCheck(id string) (CheckInfo, bool) {
	for _, c := range Checks {
		if c.ID == id {
			return c, true
		}
	}
	return CheckInfo{}, false
}

// ---------------------------------------------------------------- frontmatter

var (
	nameRe     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	whenRe     = regexp.MustCompile(`(?i)\b(use (this|it|for)|when|whenever|before|trigger(s|ed)?|invoke|if (the )?(user|person|you))\b`)
	firstPerRe = regexp.MustCompile(`(?i)^(i|i'm|i'll|you|you'll|you can)\b`)
)

func checkFrontmatter(s *Skill, _ Options) Result {
	if !s.FM.Present {
		return Result{Status: Fail, Summary: "SKILL.md has no YAML frontmatter",
			Fix: "Add a `---` block with `name:` and `description:` at the top of SKILL.md."}
	}
	var fails, warns []string
	name := s.FM.Fields["name"]
	switch {
	case name == "":
		fails = append(fails, "missing `name`")
	case len(name) > 64:
		fails = append(fails, fmt.Sprintf("`name` is %d chars (max 64)", len(name)))
	case !nameRe.MatchString(name):
		fails = append(fails, fmt.Sprintf("`name` %q must be lowercase letters, digits and hyphens", name))
	case strings.Contains(name, "anthropic") || strings.Contains(name, "claude"):
		fails = append(fails, fmt.Sprintf("`name` %q contains a reserved word (anthropic/claude)", name))
	}
	var infos []string
	if name != "" && name != path.Base(s.Dir) {
		infos = append(infos, fmt.Sprintf("`name` %q differs from directory name %q", name, path.Base(s.Dir)))
	}
	desc := s.FM.Fields["description"]
	switch {
	case desc == "":
		fails = append(fails, "missing `description`")
	case len(desc) > 1024:
		fails = append(fails, fmt.Sprintf("`description` is %d chars (max 1024)", len(desc)))
	default:
		if len(desc) < 40 {
			warns = append(warns, fmt.Sprintf("`description` is only %d chars — too thin for Claude to know when to load it", len(desc)))
		}
		if !whenRe.MatchString(desc) {
			warns = append(warns, "`description` doesn't say *when* to use the skill (e.g. \"Use when …\")")
		}
		if firstPerRe.MatchString(desc) {
			warns = append(warns, "`description` should be third person (\"Processes X…\"), not \"I/You …\"")
		}
	}
	switch {
	case len(fails) > 0:
		return Result{Status: Fail, Summary: fails[0], Details: append(fails[1:], warns...),
			Fix: "Fix the frontmatter fields listed."}
	case len(warns) > 0:
		return Result{Status: Warn, Summary: warns[0], Details: append(warns[1:], infos...),
			Fix: "Rewrite the description in third person, stating what the skill does and the situations/keywords that should trigger it."}
	case len(infos) > 0:
		return Result{Status: Info, Summary: infos[0],
			Fix: "Rename the directory or the `name` so they match; shared/installed copies are usually keyed by directory."}
	}
	return Result{Status: Pass, Summary: fmt.Sprintf("name %q, description %d chars with trigger language", name, len(desc))}
}

// ---------------------------------------------------------------- length

func checkLength(s *Skill, opt Options) Result {
	n := len(s.BodyLines())
	warnAt := opt.MaxSkillLines * 4 / 5
	switch {
	case n > opt.MaxSkillLines:
		return Result{Status: Fail, Summary: fmt.Sprintf("SKILL.md body is %d lines (limit %d)", n, opt.MaxSkillLines),
			Fix: "Move detail into reference files split by area (one file per domain/client/topic) and link each directly from SKILL.md."}
	case n > warnAt:
		return Result{Status: Warn, Summary: fmt.Sprintf("SKILL.md body is %d lines — approaching the %d limit", n, opt.MaxSkillLines),
			Fix: "Start splitting detail into reference files linked from SKILL.md."}
	}
	return Result{Status: Pass, Summary: fmt.Sprintf("SKILL.md body is %d lines", n)}
}

// ---------------------------------------------------------------- contents

var (
	tocHeadRe  = regexp.MustCompile(`(?i)^\W*(table of contents|contents|toc|index|sections|in this (file|document|guide|reference)|quick (navigation|links))\W*$`)
	listItemRe = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+(.*)$`)
	anchorRe   = regexp.MustCompile(`\]\(#([^)]+)\)`)
	shoutRe    = regexp.MustCompile(`\b(MUST|NEVER|ALWAYS|IMPORTANT|CRITICAL|REQUIRED|MANDATORY)\b`)
)

func checkContents(s *Skill, opt Options) Result {
	var fails, warns, ok []string
	for _, d := range s.SortedDocs() {
		if len(d.Lines) <= opt.HeadLines {
			continue
		}
		block, found := findTOC(d, opt.HeadLines)
		late := lateRules(d, opt.HeadLines)
		if !found {
			msg := fmt.Sprintf("%s (%d lines) has no contents list in its first %d lines", d.Rel, len(d.Lines), opt.HeadLines)
			if late > 0 {
				msg += fmt.Sprintf("; %d emphasised rule(s) sit after line %d and may never be seen", late, opt.HeadLines)
			}
			fails = append(fails, msg)
			continue
		}
		secs := sectionHeadings(d, block)
		if len(secs) == 0 {
			ok = append(ok, d.Rel)
			continue
		}
		var missing []string
		for _, h := range secs {
			if !strings.Contains(block, norm(h)) {
				missing = append(missing, h)
			}
		}
		covered := len(secs) - len(missing)
		if float64(covered) < 0.6*float64(len(secs)) {
			warns = append(warns, fmt.Sprintf("%s: contents list covers %d/%d sections; missing: %s",
				d.Rel, covered, len(secs), strings.Join(truncList(missing, 5), "; ")))
		} else {
			ok = append(ok, d.Rel)
		}
	}
	fix := "Add a `## Contents` list at the very top of each file listing every `##` section (as anchor links), matching the headings exactly."
	switch {
	case len(fails) > 0:
		return Result{Status: Fail, Summary: fmt.Sprintf("%d long reference file(s) without a contents list", len(fails)),
			Details: append(fails, warns...), Fix: fix}
	case len(warns) > 0:
		return Result{Status: Warn, Summary: fmt.Sprintf("%d contents list(s) don't match the headings", len(warns)),
			Details: warns, Fix: fix}
	case len(ok) == 0:
		return Result{Status: NA, Summary: fmt.Sprintf("no reference files over %d lines", opt.HeadLines)}
	}
	return Result{Status: Pass, Summary: fmt.Sprintf("%d long reference file(s) have contents lists", len(ok)), Details: ok}
}

// findTOC looks for a contents list in the first `window` lines and
// returns its normalised text.
func findTOC(d *Doc, window int) (string, bool) {
	lim := min(window, len(d.Lines))
	for i := 0; i < lim; i++ {
		if d.InCode[i] {
			continue
		}
		l := strings.TrimSpace(d.Lines[i])
		label := strings.TrimSpace(strings.TrimLeft(l, "#"))
		label = strings.Trim(label, "*_: ")
		isHead := strings.HasPrefix(l, "#") || strings.HasPrefix(l, "**") || strings.HasSuffix(l, ":")
		if isHead && tocHeadRe.MatchString(label) {
			if items := listAfter(d, i+1); len(items) >= 2 {
				return norm(strings.Join(items, " ")), true
			}
		}
	}
	// Fallback: a run of ≥3 list items that link to in-page anchors.
	run := []string{}
	for i := 0; i < lim; i++ {
		if !d.InCode[i] {
			if m := listItemRe.FindStringSubmatch(d.Lines[i]); m != nil && anchorRe.MatchString(m[1]) {
				run = append(run, m[1])
				if len(run) >= 3 && (i+1 >= lim || !anchorRe.MatchString(d.Lines[i+1])) {
					return norm(strings.Join(run, " ")), true
				}
				continue
			}
		}
		if len(run) >= 3 {
			return norm(strings.Join(run, " ")), true
		}
		run = run[:0]
	}
	return "", false
}

func listAfter(d *Doc, from int) []string {
	var items []string
	for i := from; i < len(d.Lines); i++ {
		l := d.Lines[i]
		if strings.TrimSpace(l) == "" {
			if len(items) > 0 && (i+1 >= len(d.Lines) || listItemRe.FindStringSubmatch(d.Lines[i+1]) == nil) {
				break
			}
			continue
		}
		m := listItemRe.FindStringSubmatch(l)
		if m == nil {
			break
		}
		// Include anchor slugs so `[Setup](#auth-and-setup)` matches either form.
		item := m[1]
		for _, a := range anchorRe.FindAllStringSubmatch(item, -1) {
			item += " " + strings.ReplaceAll(a[1], "-", " ")
		}
		items = append(items, item)
	}
	return items
}

// sectionHeadings returns the headings a contents list should cover: the
// shallowest level below the title that appears at least twice.
func sectionHeadings(d *Doc, tocText string) []string {
	counts := map[int]int{}
	for _, h := range d.Headings {
		if !tocHeadRe.MatchString(h.Text) {
			counts[h.Level]++
		}
	}
	level := 0
	for l := 2; l <= 4; l++ {
		if counts[l] >= 2 {
			level = l
			break
		}
	}
	if level == 0 && counts[1] >= 2 {
		level = 1
	}
	if level == 0 {
		return nil
	}
	var out []string
	for _, h := range d.Headings {
		if h.Level == level && !tocHeadRe.MatchString(h.Text) {
			out = append(out, h.Text)
		}
	}
	return out
}

func lateRules(d *Doc, after int) int {
	n := 0
	for i := after; i < len(d.Lines); i++ {
		if !d.InCode[i] {
			n += len(shoutRe.FindAllString(d.Lines[i], -1))
		}
	}
	return n
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func norm(s string) string {
	return strings.TrimSpace(nonAlnum.ReplaceAllString(strings.ToLower(s), " "))
}

func truncList(xs []string, n int) []string {
	if len(xs) <= n {
		return xs
	}
	return append(xs[:n:n], fmt.Sprintf("… +%d more", len(xs)-n))
}

// ---------------------------------------------------------------- nesting

var orphanOK = regexp.MustCompile(`(?i)(^|/)(readme|changelog|license|contributing|notice)(\.[a-z]+)?$`)

func checkNesting(s *Skill, _ Options) Result {
	docs := s.SortedDocs()
	if len(docs) == 0 {
		return Result{Status: NA, Summary: "no reference files"}
	}
	depth, via := s.depthMap()
	var deep, orphans []string
	direct := 0
	for _, d := range docs {
		dep, ok := depth[d.Rel]
		switch {
		case !ok:
			if !orphanOK.MatchString(d.Rel) {
				orphans = append(orphans, d.Rel)
			}
		case dep == 1:
			direct++
		default:
			chain := []string{d.Rel}
			for cur := d.Rel; via[cur] != ""; cur = via[cur] {
				chain = append([]string{via[cur]}, chain...)
			}
			deep = append(deep, fmt.Sprintf("%s is %d levels deep: %s", d.Rel, dep, strings.Join(chain, " → ")))
		}
	}
	var details []string
	for _, o := range orphans {
		details = append(details, o+" is not linked from SKILL.md or any reference (Claude won't know it exists)")
	}
	switch {
	case len(deep) > 0:
		return Result{Status: Fail, Summary: fmt.Sprintf("%d reference file(s) only reachable through another reference", len(deep)),
			Details: append(deep, details...),
			Fix:     "Link every reference file directly from SKILL.md (one level deep), with a one-line note on when to open it."}
	case len(orphans) > 0:
		return Result{Status: Warn, Summary: fmt.Sprintf("%d markdown file(s) not linked from SKILL.md", len(orphans)),
			Details: details,
			Fix:     "Link each file from SKILL.md with when-to-read guidance, or delete it if it's dead."}
	}
	return Result{Status: Pass, Summary: fmt.Sprintf("all %d reference file(s) linked directly from SKILL.md", direct)}
}

// ---------------------------------------------------------------- links

func checkLinks(s *Skill, _ Options) Result {
	var broken []string
	for _, d := range append([]*Doc{s.Main}, s.SortedDocs()...) {
		for _, b := range s.refs(d).Broken {
			broken = append(broken, fmt.Sprintf("%s → %s", d.Rel, b))
		}
	}
	if len(broken) > 0 {
		return Result{Status: Fail, Summary: fmt.Sprintf("%d link(s) to missing files", len(broken)),
			Details: broken, Fix: "Fix the paths or create the missing files."}
	}
	return Result{Status: Pass, Summary: "all local links resolve"}
}

// ---------------------------------------------------------------- freedom

var (
	highRiskRe = regexp.MustCompile(`(?i)rm -rf|drop|truncate|migrat|refund|invoic|payment|billing|charg|transfer|deploy|production|prod |force`)
	negatedRe  = regexp.MustCompile(`(?i)\b(never|don'?t|do not|avoid|without|not)\b`)
	fragileRe  = regexp.MustCompile(`(?i)\b(rm -rf|drop (table|database)|truncate table|delet(e|es|ing|ion)|purg(e|ing)|(db|database|schema|data) migrations?|migrat(e|ing) (the )?(db|database|schema|data)|run (the )?migrations?|refund(s|ing)?|invoic(e|es|ing)|payments?|billing|charg(e|es|ing) (the )?(card|customer)|transfer(s|ring)? (funds|money)|deploy(s|ing|ment)?|production|prod (db|database)|force[- ]push|revoke|remove (users?|members?|accounts?))\b`)
	exactRe    = regexp.MustCompile(`(?i)(run exactly|exact(ly this)? command|exactly as (written|shown)|do not (modify|change|edit) (the|this) (command|script)|don'?t (add|change) (any )?flags|do not add (any )?(flags|arguments)|without modification)`)
)

func checkFreedom(s *Skill, _ Options) Result {
	body := s.BodyLines()
	hits := map[string]int{}
	var where []string
	high := false
	for i, l := range body {
		for _, loc := range fragileRe.FindAllStringIndex(l, -1) {
			// "Never delete …" is a guardrail, not a fragile operation Claude performs.
			if negatedRe.MatchString(lastWords(l[:loc[0]], 4)) {
				continue
			}
			k := strings.ToLower(l[loc[0]:loc[1]])
			high = high || highRiskRe.MatchString(k)
			ref := fmt.Sprintf("SKILL.md:%d %q", i+1+s.FM.BodyStart, trunc(l, 90))
			if hits[k] == 0 && len(where) < 6 && (len(where) == 0 || where[len(where)-1] != ref) {
				where = append(where, ref)
			}
			hits[k]++
		}
	}
	// Scripts count as wired in when SKILL.md or a first-level reference names them.
	used := s.refs(s.Main).Targets
	depth, _ := s.depthMap()
	for rel, dep := range depth {
		if dep == 1 {
			for t := range s.refs(s.Docs[rel]).Targets {
				used[t] = true
			}
		}
	}
	var usedScripts []string
	for _, sc := range s.Scripts {
		if used[sc] {
			usedScripts = append(usedScripts, sc)
		}
	}
	exact := exactRe.MatchString(strings.Join(body, "\n"))
	lowFreedom := len(usedScripts) > 0 || exact

	terms := sortedKeys(boolSet(hits))
	switch {
	case len(hits) > 0 && !lowFreedom:
		st := Info
		if high || len(hits) >= 3 {
			st = Warn
		}
		return Result{Status: st,
			Summary: fmt.Sprintf("fragile operations mentioned (%s) but no locked-down script or exact command", strings.Join(truncList(terms, 6), ", ")),
			Details: where,
			Fix:     "For each step ask 'what happens if Claude does this differently?'. Where the answer is consequential, move the step into a script under scripts/ and tell Claude to run exactly that command (no extra flags). Leave low-risk steps as plain goals."}
	case len(hits) > 0:
		d := []string{}
		if len(usedScripts) > 0 {
			d = append(d, "scripts referenced from SKILL.md or its references: "+strings.Join(truncList(usedScripts, 6), ", "))
		}
		if exact {
			d = append(d, "uses exact-command language")
		}
		return Result{Status: Pass,
			Summary: fmt.Sprintf("fragile operations (%s) backed by low-freedom scripts/exact commands — confirm every fragile step is covered", strings.Join(truncList(terms, 4), ", ")),
			Details: d}
	case len(s.Scripts) > 0 && len(usedScripts) == 0:
		return Result{Status: Info, Summary: fmt.Sprintf("%d script(s) in the skill but none referenced from SKILL.md or its references", len(s.Scripts)),
			Details: truncList(s.Scripts, 6),
			Fix:     "Reference the scripts from SKILL.md with the exact command to run, or remove them."}
	}
	return Result{Status: Pass, Summary: "no fragile operations detected — high freedom (plain goals) is appropriate"}
}

// lastWords returns the final n words of s.
func lastWords(s string, n int) string {
	f := strings.Fields(s)
	if len(f) > n {
		f = f[len(f)-n:]
	}
	return strings.Join(f, " ")
}

func boolSet(m map[string]int) map[string]bool {
	out := map[string]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}

func trunc(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ---------------------------------------------------------------- models

var (
	modelKeys    = []string{"model", "models", "tested-models", "tested_models", "tested-with", "tested_with", "tested-on", "tested_on", "intended-models", "intended_models", "compatibility"}
	modelWordRe  = regexp.MustCompile(`(?i)\b(haiku|sonnet|opus|fable|claude-[a-z0-9.-]+)\b`)
	testedBodyRe = regexp.MustCompile(`(?i)(tested|verified|evaluated|intended)\s+(with|on|for)\b[^\n]{0,80}\b(haiku|sonnet|opus|fable)\b`)
)

func checkModels(s *Skill, _ Options) Result {
	for k, v := range s.FM.Fields {
		leaf := k
		if i := strings.LastIndexByte(k, '.'); i >= 0 {
			leaf = k[i+1:]
		}
		for _, mk := range modelKeys {
			if leaf == mk && v != "" {
				models := modelWordRe.FindAllString(v, -1)
				if len(models) == 0 && leaf == "compatibility" {
					continue
				}
				res := Result{Status: Pass, Summary: fmt.Sprintf("frontmatter `%s: %s`", k, trunc(v, 60))}
				if len(uniqLower(models)) == 1 {
					res.Status = Info
					res.Summary += " — only one model declared"
					res.Fix = "If the skill will be shared, test it on Haiku, Sonnet and Opus and list all that pass."
				}
				return res
			}
		}
	}
	if m := testedBodyRe.FindString(strings.Join(s.BodyLines(), "\n")); m != "" {
		return Result{Status: Info, Summary: fmt.Sprintf("models mentioned in body (%q) but not in frontmatter", trunc(m, 60)),
			Fix: "Move the tested-models note into the frontmatter (e.g. `tested-models: [haiku, sonnet, opus]` under `metadata:`)."}
	}
	return Result{Status: Warn, Summary: "no tested/intended models recorded",
		Fix: "Run the skill on the same task with Haiku, Sonnet and Opus. If Haiku skips a step, make the step clearer or a script; if Opus does better without the skill, cut instructions. Then record the models in the frontmatter, e.g. `tested-models: [haiku, sonnet, opus]` under `metadata:`."}
}

func uniqLower(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		x = strings.ToLower(x)
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

// ---------------------------------------------------------------- emphasis

func checkEmphasis(s *Skill, _ Options) Result {
	body := s.BodyLines()
	count := 0
	var where []string
	for i, l := range body {
		if s.Main.InCode[i+s.FM.BodyStart] {
			continue
		}
		n := len(shoutRe.FindAllString(l, -1))
		if n > 0 {
			count += n
			if len(where) < 5 {
				where = append(where, fmt.Sprintf("SKILL.md:%d %q", i+1+s.FM.BodyStart, trunc(l, 90)))
			}
		}
	}
	lines := max(len(body), 1)
	density := float64(count) * 100 / float64(lines)
	switch {
	case count >= 8 && density > 3:
		return Result{Status: Warn, Summary: fmt.Sprintf("%d ALL-CAPS directives (%.1f per 100 lines) — likely too prescriptive for newer models", count, density),
			Details: where,
			Fix:     "Rewrite shouted rules as plain statements with the reason behind them, and delete rules the model follows anyway. Test on Opus with and without each rule."}
	case count >= 4:
		return Result{Status: Info, Summary: fmt.Sprintf("%d ALL-CAPS directives (%.1f per 100 lines)", count, density),
			Details: where,
			Fix:     "Check each shouted rule is still needed on current models."}
	}
	return Result{Status: Pass, Summary: fmt.Sprintf("%d ALL-CAPS directives — tone is not over-prescriptive", count)}
}

// ---------------------------------------------------------------- checklist

var (
	numItemRe  = regexp.MustCompile(`^(\d+)[.)]\s+\S`)
	stepHeadRe = regexp.MustCompile(`(?i)^(step|phase|stage)\s*\d+\b|^\d+[.)]\s`)
	checkboxRe = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+\[[ xX]\]\s`)
)

// workflow reports whether SKILL.md describes an ordered multi-step process.
func workflow(s *Skill) (bool, string) {
	d := s.Main
	steps := 0
	for _, h := range d.Headings {
		if stepHeadRe.MatchString(h.Text) {
			steps++
		}
	}
	if steps >= 3 {
		return true, fmt.Sprintf("%d step headings", steps)
	}
	best, run, last := 0, 0, 0
	for i := s.FM.BodyStart; i < len(d.Lines); i++ {
		if d.InCode[i] {
			continue
		}
		l := d.Lines[i]
		if m := numItemRe.FindStringSubmatch(l); m != nil {
			n := atoi(m[1])
			if n == last+1 {
				run, last = run+1, n
			} else {
				run, last = 1, n
			}
			best = max(best, run)
			continue
		}
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") {
			continue // blank or continuation lines don't break a numbered list
		}
		run, last = 0, 0
	}
	if best >= 5 {
		return true, fmt.Sprintf("a %d-step numbered list", best)
	}
	return false, ""
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

func checkChecklist(s *Skill, _ Options) Result {
	ok, why := workflow(s)
	has := ""
	for _, d := range append([]*Doc{s.Main}, s.SortedDocs()...) {
		for i, l := range d.Lines {
			if checkboxRe.MatchString(l) {
				has = fmt.Sprintf("%s:%d", d.Rel, i+1)
				break
			}
		}
		if has != "" {
			break
		}
	}
	switch {
	case !ok && has == "":
		return Result{Status: NA, Summary: "no ordered multi-step workflow detected — a checklist isn't needed"}
	case has != "":
		return Result{Status: Pass, Summary: "copyable checklist found at " + has}
	}
	return Result{Status: Warn, Summary: "ordered workflow (" + why + ") with no copyable checklist",
		Fix: "If the order matters, add a short `- [ ] Step N: goal` checklist Claude copies into its reply and ticks off, with a short paragraph per step. If order doesn't matter, leave the steps as goals instead."}
}

// ---------------------------------------------------------------- feedback

var (
	loopRe     = regexp.MustCompile(`(?i)(\brepeat\b|re-?run|\buntil\b[^.\n]{0,40}\b(pass|passes|valid|clean|succeed|no (errors|issues))|(return|go back|loop back) to step|\biterate\b|fix (the |any |all )?(errors|issues|problems|failures)[^.\n]{0,60}(again|re-?run|repeat|re-?validate|re-?check))`)
	validateRe = regexp.MustCompile(`(?i)\b(validat\w*|verif\w*|lint\w*|check (the|your|that|against))\b`)
)

func checkFeedback(s *Skill, _ Options) Result {
	isFlow, _ := workflow(s)
	var validators []string
	for _, sc := range s.Scripts {
		b := strings.ToLower(path.Base(sc))
		if strings.Contains(b, "valid") || strings.Contains(b, "check") || strings.Contains(b, "verify") || strings.Contains(b, "lint") {
			validators = append(validators, sc)
		}
	}
	body := strings.Join(s.BodyLines(), "\n")
	mentionsCheck := validateRe.MatchString(body)
	if !isFlow && len(validators) == 0 && !mentionsCheck {
		return Result{Status: NA, Summary: "no workflow or validation step — a feedback loop isn't required"}
	}
	for _, d := range append([]*Doc{s.Main}, s.SortedDocs()...) {
		text := strings.Join(d.Lines, "\n")
		if d == s.Main {
			text = body
		}
		if m := loopRe.FindString(text); m != "" {
			return Result{Status: Pass, Summary: fmt.Sprintf("validate→fix→repeat loop found in %s (%q)", d.Rel, trunc(m, 40))}
		}
	}
	r := Result{Status: Warn, Summary: "checks/validation present but no 'fix and repeat until it passes' loop",
		Fix: "After the producing step, add: run the check (validator script or the style/brand guide), note each issue, fix it, and repeat until everything passes — e.g. 'if citations are incomplete, return to step 3'. Optionally end with: 'propose any new rule this run revealed for the guide'."}
	if len(validators) > 0 {
		r.Details = []string{"validator scripts: " + strings.Join(truncList(validators, 5), ", ")}
	}
	if !isFlow && len(validators) == 0 {
		r.Status = Info
	}
	return r
}
