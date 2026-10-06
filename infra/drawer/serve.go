package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// serve is a static file server for local previews.
func serve(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	port := fs.Int("port", 8000, "port to listen on (127.0.0.1 only)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "error: %s is not a directory\n", dir)
		return 2
	}
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	fmt.Fprintf(stdout, "serving %s at http://%s/ (Ctrl-C to stop)\n", abs, addr)
	h := http.FileServer(http.Dir(abs))
	logged := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(stdout, "%s %s\n", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
	if err := http.ListenAndServe(addr, logged); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}
