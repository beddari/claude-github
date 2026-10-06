// Command skill-audit checks Claude skills against Anthropic's current
// skill-authoring guidance and can emit a prompt asking Claude to fix them.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beddari/claude-github/tools/skill-audit/internal/audit"
)

const usage = `skill-audit — audit Claude skills against current best practice

Usage:
  skill-audit [flags] [path ...]

Each path may be a skills folder (searched recursively for SKILL.md), a
single skill directory, or a SKILL.md file. With no paths it audits
~/.claude/skills and ./.claude/skills (whichever exist).

Flags:
`

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("skill-audit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "text", "report format: text | markdown | json")
	prompt := fs.Bool("prompt", false, "print a Claude prompt to fix the findings instead of a report")
	promptDir := fs.String("prompt-dir", "", "also write one <skill>.prompt.md per skill needing work into this dir")
	withInfo := fs.Bool("include-info", false, "include info-level findings in prompts")
	only := fs.String("only", "", "comma-separated check IDs to run (see -list-checks)")
	strict := fs.Bool("strict", false, "exit 1 if any check fails (2 on warnings with -strict-warn)")
	strictWarn := fs.Bool("strict-warn", false, "with -strict, also exit 2 on warnings")
	noColor := fs.Bool("no-color", false, "disable ANSI colour")
	verbose := fs.Bool("v", false, "show details for passing checks too")
	headLines := fs.Int("head-lines", 100, "lines Claude previews of a reference file; longer files need a contents list")
	maxLines := fs.Int("max-lines", 500, "SKILL.md body line limit")
	list := fs.Bool("list-checks", false, "list the checks and the rule each enforces")
	fs.Usage = func() { fmt.Fprint(stderr, usage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if *list {
		for _, c := range audit.Checks {
			fmt.Fprintf(stdout, "%-11s %s\n            %s\n", c.ID, c.Title, c.Rule)
		}
		return 0
	}

	opt := audit.DefaultOptions()
	opt.HeadLines, opt.MaxSkillLines = *headLines, *maxLines
	if *only != "" {
		opt.Only = map[string]bool{}
		for _, id := range strings.Split(*only, ",") {
			id = strings.TrimSpace(id)
			if _, ok := audit.LookupCheck(id); !ok {
				fmt.Fprintf(stderr, "unknown check %q (see -list-checks)\n", id)
				return 2
			}
			opt.Only[id] = true
		}
	}

	roots := fs.Args()
	if len(roots) == 0 {
		roots = defaultRoots()
		if len(roots) == 0 {
			fmt.Fprintln(stderr, "no paths given and neither ~/.claude/skills nor ./.claude/skills exists")
			return 2
		}
	}
	dirs, err := audit.Discover(roots)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}

	var reps []audit.Report
	for _, d := range dirs {
		s, err := audit.Load(d)
		if err != nil {
			fmt.Fprintln(stderr, "skip:", err)
			continue
		}
		reps = append(reps, audit.Run(s, opt))
	}

	if *promptDir != "" {
		if err := writePrompts(*promptDir, reps, *withInfo); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
	}

	switch {
	case *prompt:
		fmt.Fprint(stdout, audit.CombinedPrompt(reps, *withInfo))
	case *format == "json":
		if err := audit.WriteJSON(stdout, reps); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
	case *format == "markdown" || *format == "md":
		audit.WriteMarkdown(stdout, reps)
	case *format == "text":
		audit.WriteText(stdout, reps, audit.TextOptions{Color: useColor(stdout, *noColor), Verbose: *verbose})
	default:
		fmt.Fprintf(stderr, "unknown -format %q\n", *format)
		return 2
	}

	if *strict {
		worst := audit.NA
		for _, r := range reps {
			worst = max(worst, r.Worst())
		}
		if worst == audit.Fail {
			return 1
		}
		if *strictWarn && worst == audit.Warn {
			return 2
		}
	}
	return 0
}

func defaultRoots() []string {
	var roots []string
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".claude", "skills"))
	}
	roots = append(roots, filepath.Join(".claude", "skills"))
	var out []string
	for _, r := range roots {
		if info, err := os.Stat(r); err == nil && info.IsDir() {
			out = append(out, r)
		}
	}
	return out
}

func writePrompts(dir string, reps []audit.Report, withInfo bool) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, r := range reps {
		if !audit.NeedsWork(r, withInfo) {
			continue
		}
		name := strings.Map(func(c rune) rune {
			if c == '/' || c == '\\' || c == ' ' {
				return '_'
			}
			return c
		}, r.Name)
		if err := os.WriteFile(filepath.Join(dir, name+".prompt.md"), []byte(audit.Prompt(r, withInfo)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func useColor(w io.Writer, disabled bool) bool {
	if disabled || os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
