// Command drawer maintains the monorepo's drawer of experiments ("wires"):
// the wires.json index, the GitHub Pages site built from it, and a local
// static server. It is repo infrastructure, not a wire itself.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usage = `drawer — keep the wire drawer in order

Usage:
  drawer [-root DIR] check                           every apps/*, tools/* and labs/* project is in wires.json, valid, with a page
  drawer [-root DIR] add NAME PATH KIND [BLURB [HREF]] append a wire (KIND: app | tool | lab; HREF: its page, default PATH/)
  drawer [-root DIR] site [OUT]                      build the Pages site into OUT (default _site)
  drawer serve [-port 8000] [DIR]                    serve DIR (default .) at http://127.0.0.1:PORT/
  drawer json FILE KEY.PATH                          print a JSON string value; fail if missing or empty
  drawer [-root DIR] docs                            every relative link in every markdown file points to a file and heading that exist

-root defaults to the nearest parent directory containing wires.json.
`

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("drawer", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", "", "repo root (default: nearest parent with wires.json)")
	fs.Usage = func() { fmt.Fprint(stderr, usage) }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	args = fs.Args()
	if len(args) == 0 {
		fs.Usage()
		return 2
	}
	cmd, args := args[0], args[1:]

	// Commands that don't need the repo root.
	switch cmd {
	case "serve":
		return serve(args, stdout, stderr)
	case "json":
		if len(args) != 2 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		v, err := jsonString(args[0], args[1])
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		fmt.Fprintln(stdout, v)
		return 0
	}

	r := *root
	if r == "" {
		var err error
		if r, err = findRoot("."); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
	}
	var err error
	switch {
	case cmd == "check" && len(args) == 0:
		var errs []string
		var n int
		if n, errs, err = check(r); err == nil {
			for _, e := range errs {
				fmt.Fprintln(stderr, "wires.json:", e)
			}
			if len(errs) > 0 {
				return 1
			}
			fmt.Fprintf(stdout, "wires.json OK (%d wires)\n", n)
		}
	case cmd == "add" && len(args) >= 3 && len(args) <= 5:
		blurb, href := "TODO: describe this wire.", ""
		if len(args) >= 4 {
			blurb = args[3]
		}
		if len(args) == 5 {
			href = args[4]
		}
		var added bool
		if added, err = add(r, Wire{Name: args[0], Path: args[1], Kind: args[2], Blurb: blurb, Href: href}); err == nil {
			if added {
				fmt.Fprintf(stdout, "added %s to wires.json\n", args[1])
			} else {
				fmt.Fprintf(stdout, "%s already listed\n", args[1])
			}
		}
	case cmd == "docs" && len(args) == 0:
		var problems []string
		var n int
		if problems, n, err = checkDocs(r); err == nil {
			for _, p := range problems {
				fmt.Fprintln(stderr, p)
			}
			if len(problems) > 0 {
				fmt.Fprintf(stderr, "%d broken link(s) in %d markdown files\n", len(problems), n)
				return 1
			}
			fmt.Fprintf(stdout, "docs OK (%d markdown files, all relative links resolve)\n", n)
		}
	case cmd == "site" && len(args) <= 1:
		out := filepath.Join(r, "_site")
		if len(args) == 1 {
			out = args[0]
		}
		err = buildSite(r, out, stdout)
	default:
		fs.Usage()
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

// findRoot walks up from dir to the first directory holding wires.json.
func findRoot(dir string) (string, error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "wires.json")); err == nil {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", errors.New("no wires.json found in this directory or any parent (use -root)")
		}
		d = parent
	}
}
