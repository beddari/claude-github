package audit

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
)

// dependency is one third-party package or CLI the skill relies on.
type dependency struct {
	Name    string   // import/command name as used
	Install []string // names that count as installing it (pip/npm/apt names)
	Kind    string   // "python", "node", "cli"
	Where   string   // first place it was seen
}

var (
	pyImportRe   = regexp.MustCompile(`^\s*import\s+([A-Za-z_][\w.]*(?:\s+as\s+\w+)?(?:\s*,\s*[A-Za-z_][\w.]*(?:\s+as\s+\w+)?)*)\s*(?:#.*)?$`)
	pyFromRe     = regexp.MustCompile(`^\s*from\s+([A-Za-z_][\w.]*)\s+import\b`)
	jsRequireRe  = regexp.MustCompile(`require\(\s*['"]([^'"./][^'"]*)['"]\s*\)`)
	jsImportRe   = regexp.MustCompile(`(?:^|\s)(?:import|export)\s[^'"]*?from\s+['"]([^'"./][^'"]*)['"]|^\s*import\s+['"]([^'"./][^'"]*)['"]`)
	installLnRe  = regexp.MustCompile(`(?i)\b(install|requires?|dependenc|prerequisite|npx|uvx|pipx)`)
	pep723Re     = regexp.MustCompile(`(?m)^# /// script`)
	cliToolNames = map[string][]string{
		"jq": {"jq"}, "yq": {"yq"}, "pandoc": {"pandoc"}, "ffmpeg": {"ffmpeg"}, "ffprobe": {"ffmpeg"},
		"magick": {"imagemagick"}, "pdftotext": {"poppler-utils", "poppler"}, "pdftoppm": {"poppler-utils", "poppler"},
		"pdfimages": {"poppler-utils", "poppler"}, "qpdf": {"qpdf"}, "soffice": {"libreoffice"}, "libreoffice": {"libreoffice"},
		"rg": {"ripgrep"}, "gh": {"gh"}, "aws": {"awscli", "aws-cli"}, "gcloud": {"google-cloud-sdk", "gcloud"},
		"kubectl": {"kubectl"}, "terraform": {"terraform"}, "tesseract": {"tesseract-ocr", "tesseract"},
		"wkhtmltopdf": {"wkhtmltopdf"}, "playwright": {"playwright"}, "mmdc": {"@mermaid-js/mermaid-cli", "mermaid-cli"},
		"exiftool": {"exiftool", "libimage-exiftool-perl"}, "pdftk": {"pdftk"}, "ghostscript": {"ghostscript"}, "gs": {"ghostscript"},
	}
	pipNames = map[string]string{
		"PIL": "pillow", "yaml": "pyyaml", "cv2": "opencv-python", "docx": "python-docx", "pptx": "python-pptx",
		"fitz": "pymupdf", "bs4": "beautifulsoup4", "sklearn": "scikit-learn", "dateutil": "python-dateutil",
		"dotenv": "python-dotenv", "magic": "python-magic", "attr": "attrs", "Crypto": "pycryptodome",
		"jwt": "pyjwt", "serial": "pyserial", "usb": "pyusb", "pytesseract": "pytesseract", "pdf2image": "pdf2image",
		"googleapiclient": "google-api-python-client", "OpenSSL": "pyopenssl", "Levenshtein": "python-levenshtein",
	}
	nodeBuiltins = set("assert", "async_hooks", "buffer", "child_process", "cluster", "console", "crypto", "dgram", "dns",
		"events", "fs", "fs/promises", "http", "http2", "https", "module", "net", "os", "path", "perf_hooks", "process",
		"querystring", "readline", "stream", "string_decoder", "timers", "tls", "tty", "url", "util", "v8", "vm",
		"worker_threads", "zlib")
	pyStdlib = set("__future__", "_thread", "abc", "aifc", "argparse", "array", "ast", "asyncio", "atexit", "base64", "bdb",
		"binascii", "bisect", "builtins", "bz2", "calendar", "cgi", "cgitb", "chunk", "cmath", "cmd", "code", "codecs",
		"codeop", "collections", "colorsys", "compileall", "concurrent", "configparser", "contextlib", "contextvars",
		"copy", "copyreg", "cProfile", "csv", "ctypes", "curses", "dataclasses", "datetime", "dbm", "decimal", "difflib",
		"dis", "doctest", "email", "encodings", "ensurepip", "enum", "errno", "faulthandler", "fcntl", "filecmp",
		"fileinput", "fnmatch", "fractions", "ftplib", "functools", "gc", "getopt", "getpass", "gettext", "glob",
		"graphlib", "grp", "gzip", "hashlib", "heapq", "hmac", "html", "http", "imaplib", "importlib", "inspect", "io",
		"ipaddress", "itertools", "json", "keyword", "linecache", "locale", "logging", "lzma", "mailbox", "marshal",
		"math", "mimetypes", "mmap", "multiprocessing", "netrc", "numbers", "operator", "optparse", "os", "pathlib",
		"pdb", "pickle", "pickletools", "pkgutil", "platform", "plistlib", "poplib", "posix", "posixpath", "ntpath", "genericpath", "nturl2path", "opcode", "sre_compile", "sre_parse", "sre_constants", "pprint", "profile",
		"pstats", "pty", "pwd", "py_compile", "pyclbr", "pydoc", "queue", "quopri", "random", "re", "readline",
		"reprlib", "resource", "rlcompleter", "runpy", "sched", "secrets", "select", "selectors", "shelve", "shlex",
		"shutil", "signal", "site", "smtplib", "socket", "socketserver", "sqlite3", "ssl", "stat", "statistics",
		"string", "stringprep", "struct", "subprocess", "symtable", "sys", "sysconfig", "syslog", "tabnanny", "tarfile",
		"tempfile", "termios", "textwrap", "threading", "time", "timeit", "tkinter", "token", "tokenize", "tomllib",
		"trace", "traceback", "tracemalloc", "tty", "turtle", "types", "typing", "unicodedata", "unittest", "urllib",
		"uuid", "venv", "warnings", "wave", "weakref", "webbrowser", "winreg", "wsgiref", "xml", "xmlrpc", "zipapp",
		"zipfile", "zipimport", "zlib", "zoneinfo")
	cliRe = func() *regexp.Regexp {
		var names []string
		for k := range cliToolNames {
			names = append(names, regexp.QuoteMeta(k))
		}
		sort.Strings(names)
		return regexp.MustCompile(`(?:^|[|;&(]\s*|\$\(\s*|^\s*(?:sudo\s+)?)(` + strings.Join(names, "|") + `)\s`)
	}()
)

func set(xs ...string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return m
}

// localModules returns names that resolve to code inside the skill.
func (s *Skill) localModules() map[string]bool {
	m := map[string]bool{}
	for f := range s.Files {
		parts := strings.Split(f, "/")
		for _, p := range parts[:len(parts)-1] {
			m[p] = true
		}
		base := parts[len(parts)-1]
		if ext := path.Ext(base); ext == ".py" || ext == ".js" || ext == ".ts" {
			m[strings.TrimSuffix(base, ext)] = true
		}
	}
	return m
}

func (s *Skill) dependencies() []dependency {
	local := s.localModules()
	found := map[string]*dependency{}
	var order []string
	add := func(kind, name, where string, install ...string) {
		key := kind + ":" + name
		if _, ok := found[key]; ok {
			return
		}
		found[key] = &dependency{Name: name, Kind: kind, Where: where, Install: append([]string{name}, install...)}
		order = append(order, key)
	}
	scanPython := func(line, where string) {
		var mods []string
		if m := pyFromRe.FindStringSubmatch(line); m != nil {
			mods = append(mods, m[1])
		} else if m := pyImportRe.FindStringSubmatch(line); m != nil {
			for _, part := range strings.Split(m[1], ",") {
				mods = append(mods, strings.Fields(part)[0])
			}
		}
		for _, mod := range mods {
			top := strings.SplitN(mod, ".", 2)[0]
			if pyStdlib[top] || local[top] {
				continue
			}
			var alt []string
			if p, ok := pipNames[top]; ok {
				alt = append(alt, p)
			}
			add("python", top, where, alt...)
		}
	}
	scanNode := func(line, where string) {
		var mods []string
		for _, m := range jsRequireRe.FindAllStringSubmatch(line, -1) {
			mods = append(mods, m[1])
		}
		for _, m := range jsImportRe.FindAllStringSubmatch(line, -1) {
			mods = append(mods, m[1]+m[2])
		}
		for _, mod := range mods {
			if strings.HasPrefix(mod, "node:") || nodeBuiltins[mod] {
				continue
			}
			pkg := mod
			parts := strings.Split(mod, "/")
			if strings.HasPrefix(mod, "@") && len(parts) > 1 {
				pkg = parts[0] + "/" + parts[1]
			} else {
				pkg = parts[0]
			}
			if nodeBuiltins[pkg] || local[pkg] {
				continue
			}
			add("node", pkg, where)
		}
	}
	scanCLI := func(line, where string) {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "#") || strings.HasPrefix(t, "//") || installLnRe.MatchString(t) {
			return
		}
		for _, m := range cliRe.FindAllStringSubmatch(line+" ", -1) {
			add("cli", m[1], where, cliToolNames[m[1]]...)
		}
	}

	for _, sc := range s.Scripts {
		src, ok := s.Sources[sc]
		if !ok || pep723Re.MatchString(src) {
			continue // PEP 723 inline metadata declares its own deps
		}
		ext := path.Ext(sc)
		for i, l := range strings.Split(src, "\n") {
			where := fmt.Sprintf("%s:%d", sc, i+1)
			switch ext {
			case ".py":
				scanPython(l, where)
				if strings.Contains(l, "subprocess") || strings.Contains(l, "run(") {
					scanQuotedCLI(l, where, add)
				}
			case ".js", ".mjs", ".cjs", ".ts":
				scanNode(l, where)
			case ".sh", ".bash":
				scanCLI(l, where)
			}
		}
	}
	for _, d := range append([]*Doc{s.Main}, s.SortedDocs()...) {
		for i, l := range d.Lines {
			if !d.InCode[i] {
				continue
			}
			where := fmt.Sprintf("%s:%d", d.Rel, i+1)
			switch d.CodeLang[i] {
			case "python", "py":
				scanPython(l, where)
			case "javascript", "js", "typescript", "ts", "node":
				scanNode(l, where)
			case "bash", "sh", "shell", "console", "zsh", "":
				scanCLI(strings.TrimPrefix(strings.TrimSpace(l), "$ "), where)
			}
		}
	}
	out := make([]dependency, 0, len(order))
	for _, k := range order {
		out = append(out, *found[k])
	}
	return out
}

var quotedCmdRe = regexp.MustCompile(`\[\s*['"]([\w-]+)['"]`)

// scanQuotedCLI catches subprocess.run(["pandoc", ...]) style calls.
func scanQuotedCLI(line, where string, add func(kind, name, where string, install ...string)) {
	for _, m := range quotedCmdRe.FindAllStringSubmatch(line, -1) {
		if alt, ok := cliToolNames[m[1]]; ok {
			add("cli", m[1], where, alt...)
		}
	}
}

// installCorpus collects every line that could document an install.
func (s *Skill) installCorpus() (lines []string, manifests string) {
	for rel, src := range s.Sources {
		if isManifest(path.Base(rel)) {
			manifests += "\n" + strings.ToLower(src)
			continue
		}
		for _, l := range strings.Split(src, "\n") {
			if installLnRe.MatchString(l) {
				lines = append(lines, strings.ToLower(l))
			}
		}
	}
	return lines, manifests
}

func wordIn(hay, needle string) bool {
	needle = strings.ToLower(needle)
	re := regexp.MustCompile(`(?:^|[^a-z0-9_-])` + regexp.QuoteMeta(needle) + `(?:[^a-z0-9_-]|$)`)
	return re.MatchString(hay)
}

func checkDeps(s *Skill, _ Options) Result {
	deps := s.dependencies()
	if len(deps) == 0 {
		return Result{Status: NA, Summary: "no third-party packages or CLI tools detected"}
	}
	lines, manifests := s.installCorpus()
	var missing, ok []string
	for _, d := range deps {
		satisfied := false
		for _, name := range d.Install {
			if wordIn(manifests, name) {
				satisfied = true
				break
			}
			for _, l := range lines {
				if wordIn(l, name) {
					satisfied = true
					break
				}
			}
			if satisfied {
				break
			}
		}
		label := fmt.Sprintf("%s `%s`", d.Kind, d.Name)
		if len(d.Install) > 1 {
			label += " (install: " + d.Install[1] + ")"
		}
		if satisfied {
			ok = append(ok, label)
		} else {
			missing = append(missing, label+" — first used at "+d.Where)
		}
	}
	if len(missing) > 0 {
		st := Fail
		if len(missing) <= len(deps)/4 {
			st = Warn
		}
		return Result{Status: st,
			Summary: fmt.Sprintf("%d of %d dependencies have no install instruction", len(missing), len(deps)),
			Details: missing,
			Fix:     "Next to each script/code example, add the install line with the real package name (e.g. `pip install pypdf`, `npm install docx`, `apt-get install poppler-utils`) — Claude skips it if already installed. Or add requirements.txt / PEP 723 metadata and tell Claude to install from it."}
	}
	return Result{Status: Pass, Summary: fmt.Sprintf("all %d dependencies have install instructions", len(deps)),
		Details: truncList(ok, 8)}
}
