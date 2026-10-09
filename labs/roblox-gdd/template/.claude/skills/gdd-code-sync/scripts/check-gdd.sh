#!/usr/bin/env bash
#
# Checks that the design vault and the code agree. Run it from the root of
# the game project, or give the root as the first argument. Needs bash and
# the usual text tools: find, grep, sed, awk, sort, uniq, comm, tr and wc,
# which Git for Windows has too. Reads files with CRLF line endings as LF.
#
# Usage: check-gdd.sh [PROJECT_ROOT]
#

project="${1:-.}"
design="$project/design"
remotes_file="$project/src/shared/Remotes.luau"
networking="$design/Game/Networking.md"
problems=0

#
# Prints a log message.
#
function log()
{
	echo ">>> $1"
}

#
# Prints one problem and counts it.
#
function problem()
{
	echo "!!! $1" >&2
	problems=$((problems + 1))
}

#
# Prints an error message and exits.
#
function fail()
{
	echo "!!! $1" >&2
	exit 1
}

#
# Prints every page of the vault, relative to design/, one per line. Paths
# with a segment that starts with "." are skipped, as PageTree does.
#
function pages()
{
	(cd "$design" && find . -name '*.md' -type f) |
		sed 's#^\./##' | grep -Ev '(^|/)\.' | sort
}

#
# Prints the remotes that Remotes.luau declares, one per line: every quoted
# name between the "{" after "local REMOTE_EVENTS ... =" or
# "local REMOTE_FUNCTIONS ... =" and its "}", outside "--" comments.
#
function code_remotes()
{
	awk 'function scan(text,    closed) {
		sub(/--.*/, "", text)
		closed = index(text, "}")
		if (closed) text = substr(text, 1, closed - 1)
		while (match(text, /"[A-Za-z0-9_]+"/)) {
			print substr(text, RSTART + 1, RLENGTH - 2)
			text = substr(text, RSTART + RLENGTH)
		}
		if (closed) inside = 0
	     }
	     { sub(/\r$/, "") }
	     inside { scan($0); next }
	     /^local REMOTE_(EVENTS|FUNCTIONS)[^=]*=[ \t]*\{/ {
		text = $0
		sub(/^[^=]*=[ \t]*\{/, "", text)
		inside = 1
		scan(text)
	     }' "$remotes_file" | sort
}

#
# Prints the "### Name" headings under "## Remotes" in Networking.md.
#
function documented_remotes()
{
	awk '{ sub(/\r$/, "") }
	     /^## / { inside = ($0 ~ /^## Remotes[ \t]*$/) }
	     inside && /^### / { sub(/^### +/, ""); sub(/[ \t]+$/, ""); print }' \
		"$networking" | sort
}

#
# Checks that the remotes in the code and in Networking.md are the same.
#
function check_remotes()
{
	local name

	while IFS= read -r name; do
		problem "Remote $name is in Remotes.luau but has no '### $name' in Networking.md"
	done < <(comm -23 <(code_remotes) <(documented_remotes))
	while IFS= read -r name; do
		problem "Networking.md documents $name, which Remotes.luau does not declare"
	done < <(comm -13 <(code_remotes) <(documented_remotes))
}

#
# Checks that every folder holding pages has its page next to it: PageTree
# keeps the children of X.md in X/.
#
function check_pairs()
{
	local folder

	while IFS= read -r folder; do
		[[ -f "$design/$folder.md" ]] ||
			problem "design/$folder/ holds pages but design/$folder.md is missing"
	done < <(pages | grep / | sed 's#/[^/]*$##' | sort -u |
		awk -F/ '{ path = $1; print path
			   for (i = 2; i <= NF; i++) { path = path "/" $i; print path } }' |
		sort -u)
}

#
# Checks that no two pages share a file name, ignoring case as Obsidian
# does, so every [[link]] has one target.
#
function check_unique_names()
{
	local name

	while IFS= read -r name; do
		problem "More than one page is named '$name'; [[$name]] is ambiguous"
	done < <(pages | sed 's#.*/##; s#\.md$##' |
		tr '[:upper:]' '[:lower:]' | sort | uniq -d)
}

#
# Prints the frontmatter of a page: the lines between the first two "---".
#
function frontmatter()
{
	awk '{ sub(/\r$/, "") }
	     NR == 1 && $0 != "---" { exit }
	     NR > 1 && $0 == "---" { exit }
	     NR > 1 { print }' "$1"
}

#
# Prints the items of one list field of a page's frontmatter, one per line.
# Reads "key: [a, b]", "key:" followed by "  - a" lines, and "key: a".
#
function frontmatter_list()
{
	local page="$1"
	local key="$2"

	frontmatter "$page" | awk -v key="$key" '
		function trim(s) { gsub(/^[ \t"'"'"']+|[ \t"'"'"']+$/, "", s); return s }
		/^[A-Za-z_-]+:/ { inside = 0 }
		$0 ~ "^" key ":" {
			value = $0
			sub("^" key ":[ \t]*", "", value)
			if (value ~ /^\[.*\][ \t]*$/) {
				gsub(/^\[|\][ \t]*$/, "", value)
				n = split(value, items, ",")
				for (i = 1; i <= n; i++) if (trim(items[i]) != "") print trim(items[i])
			} else if (trim(value) != "") {
				print trim(value)
			} else {
				inside = 1
			}
			next
		}
		inside && /^[ \t]*- / { sub(/^[ \t]*- /, ""); if (trim($0) != "") print trim($0) }'
}

#
# Checks the code paths and the remotes that each page's frontmatter names.
#
function check_frontmatter()
{
	local page path remote known

	known="$(code_remotes)"
	while IFS= read -r page; do
		while IFS= read -r path; do
			[[ -e "$project/$path" ]] ||
				problem "design/$page lists code: $path, which does not exist"
		done < <(frontmatter_list "$design/$page" code)
		while IFS= read -r remote; do
			grep -qx "$remote" <<< "$known" ||
				problem "design/$page lists remote $remote, which Remotes.luau does not declare"
		done < <(frontmatter_list "$design/$page" remotes)
	done < <(pages)
}

[[ -d "$design" ]]       || fail "No design/ folder in $project!"
[[ -f "$remotes_file" ]] || fail "No $remotes_file!"
[[ -f "$networking" ]]   || fail "No $networking!"

check_remotes
check_pairs
check_unique_names
check_frontmatter

if (( problems > 0 )); then
	fail "$problems problem(s) between design/ and src/!"
fi
log "design/ and src/ agree: $(code_remotes | wc -l | tr -d ' ') remote(s), $(pages | wc -l | tr -d ' ') page(s)."
