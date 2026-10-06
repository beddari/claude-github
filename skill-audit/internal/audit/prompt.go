package audit

import (
	"fmt"
	"strings"
)

// NeedsWork reports whether a report has anything at Warn or above
// (or at Info when includeInfo is set).
func NeedsWork(r Report, includeInfo bool) bool {
	floor := Warn
	if includeInfo {
		floor = Info
	}
	return r.Worst() >= floor
}

// Prompt builds a prompt asking Claude to fix one skill. It follows the
// "show me what you'd change before changing anything" audit style.
func Prompt(r Report, includeInfo bool) string {
	floor := Warn
	if includeInfo {
		floor = Info
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Update the `%s` skill to current skill-authoring best practice\n\n", r.Name)
	fmt.Fprintf(&b, "The skill lives at `%s`. An automated audit found the issues below. ", r.Dir)
	b.WriteString("Read SKILL.md and every file it references before proposing anything.\n\n")

	b.WriteString("## How to work\n\n")
	b.WriteString("1. Read the whole skill, then for each finding below confirm or reject it — the audit is heuristic, so say so when a finding is a false positive and skip it.\n")
	b.WriteString("2. Show me a file-by-file plan with the exact edits (diff-style) **before changing anything**, and wait for my go-ahead.\n")
	b.WriteString("3. Keep the skill's behaviour and intent. Most fixes are about layout, structure and checks, not rewording; only change wording a finding calls for.\n")
	b.WriteString("4. After applying, re-run `skill-audit " + r.Dir + "` and repeat until nothing at warn/fail remains (or each remaining item is a justified false positive).\n\n")

	b.WriteString("## Findings\n\n")
	n := 0
	for _, res := range r.Results {
		if res.Status < floor {
			continue
		}
		n++
		c, _ := LookupCheck(res.ID)
		fmt.Fprintf(&b, "### %d. [%s] %s — %s\n\n", n, strings.ToUpper(res.Status.String()), c.Title, res.Summary)
		fmt.Fprintf(&b, "**Rule:** %s\n\n", c.Rule)
		if len(res.Details) > 0 {
			b.WriteString("**Evidence:**\n")
			for _, d := range res.Details {
				fmt.Fprintf(&b, "- %s\n", d)
			}
			b.WriteString("\n")
		}
		if res.Fix != "" {
			fmt.Fprintf(&b, "**Suggested fix:** %s\n\n", res.Fix)
		}
	}
	if n == 0 {
		b.WriteString("No warn/fail findings. Only do the manual review below.\n\n")
	}

	b.WriteString("## Also review (the audit can't check these)\n\n")
	b.WriteString("- **Degrees of freedom per step:** for each step ask \"what happens if Claude does this differently?\" Nothing much → plain goal (high freedom). A preferred shape → template with options (medium). Consequential (money, deleting, migrations, sending) → an exact script with few/no parameters (low). One skill can mix levels.\n")
	b.WriteString("- **Model fit:** Haiku — enough guidance? Sonnet — clear and efficient? Opus — not over-explaining? Tell me which instructions you'd cut for stronger models and which steps need a script so smaller models can't skip them; I'll test on each model.\n")
	b.WriteString("- **Important rules early:** anything critical in a reference file must sit in its first 100 lines or be reachable from its contents list.\n")
	b.WriteString("- **Split by area:** if a reference mixes domains (finance/sales/clients…), propose one file per area so unrelated questions never load it.\n")
	b.WriteString("- **Self-improvement:** where the skill checks output against a guide, add a final step: \"propose any new rule this run revealed for the guide\" (I approve before it's added).\n")
	return b.String()
}

// CombinedPrompt joins prompts for every report that needs work.
func CombinedPrompt(reps []Report, includeInfo bool) string {
	var parts []string
	for _, r := range reps {
		if NeedsWork(r, includeInfo) {
			parts = append(parts, Prompt(r, includeInfo))
		}
	}
	if len(parts) == 0 {
		return "All audited skills pass; nothing to fix.\n"
	}
	head := fmt.Sprintf("# Skill fixes (%d skill(s))\n\nWork through the skills below one at a time. For each, propose the plan and wait for approval before editing.\n\n---\n\n", len(parts))
	return head + strings.Join(parts, "\n---\n\n")
}
