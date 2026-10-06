// Package audit inspects Claude skill directories against Anthropic's
// skill-authoring best practices (contents lists, degrees of freedom,
// model coverage, one-level references, checklists, feedback loops and
// explicit dependencies).
package audit

// Status is the outcome of a single check.
type Status int

const (
	NA Status = iota // check does not apply to this skill
	Pass
	Info
	Warn
	Fail
)

func (s Status) String() string {
	switch s {
	case Pass:
		return "pass"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Fail:
		return "fail"
	default:
		return "n/a"
	}
}

// MarshalText makes Status render as its name in JSON.
func (s Status) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// Result is the outcome of one check on one skill.
type Result struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Status  Status   `json:"status"`
	Summary string   `json:"summary"`
	Details []string `json:"details,omitempty"`
	Fix     string   `json:"fix,omitempty"`
}

// Report is the audit of one skill.
type Report struct {
	Name    string   `json:"name"`
	Dir     string   `json:"dir"`
	Results []Result `json:"results"`
}

// Worst returns the most severe status in the report.
func (r Report) Worst() Status {
	w := NA
	for _, res := range r.Results {
		if res.Status > w {
			w = res.Status
		}
	}
	return w
}

// Count returns how many results have the given status.
func (r Report) Count(s Status) int {
	n := 0
	for _, res := range r.Results {
		if res.Status == s {
			n++
		}
	}
	return n
}

// Options tunes the thresholds used by the checks.
type Options struct {
	// HeadLines is how many lines Claude previews of a reference file
	// (it runs `head -100`); longer files need a contents list.
	HeadLines int
	// MaxSkillLines is the SKILL.md body limit.
	MaxSkillLines int
	// Only restricts the run to these check IDs (empty = all).
	Only map[string]bool
}

// DefaultOptions mirrors Anthropic's published guidance.
func DefaultOptions() Options {
	return Options{HeadLines: 100, MaxSkillLines: 500}
}
