# shellcheck shell=bash
#
# Shared by the scripts in bin/. Sourced, never run.
#

root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd)" || return $?

# The folders that hold projects, each with a Taskfile.yml of its own.
project_dirs=(infra apps tools labs)

#
# Prints a log message.
#
function log()
{
	if [[ -t 1 ]]; then
		echo -e "\x1b[1m\x1b[32m>>>\x1b[0m \x1b[1m$1\x1b[0m"
	else
		echo ">>> $1"
	fi
}

#
# Prints an error message.
#
function error()
{
	if [[ -t 2 ]]; then
		echo -e "\x1b[1m\x1b[31m!!!\x1b[0m \x1b[1m$1\x1b[0m" >&2
	else
		echo "!!! $1" >&2
	fi
}

#
# Prints an error message and exits.
#
function fail()
{
	error "$1"
	exit 1
}

#
# Prints every project folder with a Taskfile.yml, relative to the root.
#
function projects()
{
	local parent dir

	for parent in "${project_dirs[@]}"; do
		for dir in "$root/$parent"/*/; do
			[[ -f "$dir/Taskfile.yml" ]] && echo "${dir#"$root"/}"
		done
	done
	return 0
}

#
# Runs drawer, the Go tool for the repository's plumbing.
#
function drawer()
{
	(cd "$root" && go run ./infra/drawer "$@")
}
