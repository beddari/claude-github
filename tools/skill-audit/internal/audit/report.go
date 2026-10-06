package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// TextOptions controls the terminal report.
type TextOptions struct {
	Color   bool
	Verbose bool // also show details for passing checks
}

func (o TextOptions) paint(code, s string) string {
	if !o.Color {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (o TextOptions) badge(s Status) string {
	switch s {
	case Pass:
		return o.paint("32", "PASS")
	case Info:
		return o.paint("36", "INFO")
	case Warn:
		return o.paint("33", "WARN")
	case Fail:
		return o.paint("31;1", "FAIL")
	}
	return o.paint("2", " N/A")
}

// WriteText renders reports for a terminal.
func WriteText(w io.Writer, reps []Report, o TextOptions) {
	for _, r := range reps {
		fmt.Fprintf(w, "%s  %s\n", o.paint("1", r.Name), o.paint("2", r.Dir))
		for _, res := range r.Results {
			fmt.Fprintf(w, "  %s  %-11s %s\n", o.badge(res.Status), res.ID, res.Summary)
			if res.Status >= Info || o.Verbose {
				for _, d := range res.Details {
					fmt.Fprintf(w, "                    %s %s\n", o.paint("2", "·"), d)
				}
			}
			if res.Status >= Warn && res.Fix != "" {
				fmt.Fprintf(w, "                    %s %s\n", o.paint("36", "fix:"), res.Fix)
			}
		}
		fmt.Fprintln(w)
	}
	WriteSummary(w, reps, o)
}

// WriteSummary prints a one-line-per-skill table plus totals.
func WriteSummary(w io.Writer, reps []Report, o TextOptions) {
	if len(reps) == 0 {
		fmt.Fprintln(w, "No skills found.")
		return
	}
	width := 5
	for _, r := range reps {
		width = max(width, len(r.Name))
	}
	fmt.Fprintf(w, "%s\n", o.paint("1", "Summary"))
	fmt.Fprintf(w, "  %-*s  %4s %4s %4s %4s  %s\n", width, "skill", "pass", "info", "warn", "fail", "needs work")
	var tp, ti, tw, tf int
	for _, r := range reps {
		p, i, wa, f := r.Count(Pass), r.Count(Info), r.Count(Warn), r.Count(Fail)
		tp, ti, tw, tf = tp+p, ti+i, tw+wa, tf+f
		var ids []string
		for _, res := range r.Results {
			if res.Status >= Warn {
				ids = append(ids, res.ID)
			}
		}
		fmt.Fprintf(w, "  %-*s  %4d %4d %4d %4d  %s\n", width, r.Name, p, i, wa, f, strings.Join(ids, ", "))
	}
	fmt.Fprintf(w, "  %-*s  %4d %4d %4d %4d\n", width, "TOTAL", tp, ti, tw, tf)
}

// WriteMarkdown renders reports as a markdown document.
func WriteMarkdown(w io.Writer, reps []Report) {
	fmt.Fprintln(w, "# Skill audit")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| skill | pass | info | warn | fail | needs work |")
	fmt.Fprintln(w, "|---|---:|---:|---:|---:|---|")
	for _, r := range reps {
		var ids []string
		for _, res := range r.Results {
			if res.Status >= Warn {
				ids = append(ids, "`"+res.ID+"`")
			}
		}
		fmt.Fprintf(w, "| %s | %d | %d | %d | %d | %s |\n", r.Name, r.Count(Pass), r.Count(Info), r.Count(Warn), r.Count(Fail), strings.Join(ids, " "))
	}
	for _, r := range reps {
		fmt.Fprintf(w, "\n## %s\n\n`%s`\n\n", r.Name, r.Dir)
		for _, res := range r.Results {
			fmt.Fprintf(w, "- **%s** `%s` — %s\n", strings.ToUpper(res.Status.String()), res.ID, res.Summary)
			for _, d := range res.Details {
				fmt.Fprintf(w, "  - %s\n", d)
			}
			if res.Status >= Warn && res.Fix != "" {
				fmt.Fprintf(w, "  - *Fix:* %s\n", res.Fix)
			}
		}
	}
}

// WriteJSON renders reports as JSON.
func WriteJSON(w io.Writer, reps []Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(reps)
}
