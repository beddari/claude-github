#!/usr/bin/env bash
#
# SessionStart hook for Claude Code cloud sessions. Links the dataverket
# skills into ~/.claude/skills and installs the tools that "task ci" and
# "task lint" need. Runs again on every start and only does what is missing.
#

skills_url="https://git.dataverket.org/dataverket/skills.git"
skills_dir="${DATAVERKET_SKILLS_DIR:-$HOME/.local/share/dataverket-skills}"
gobin="$(go env GOPATH 2>/dev/null)/bin"

#
# Prints a log message.
#
function log()
{
	echo ">>> $1"
}

#
# Prints an error message.
#
function error()
{
	echo "!!! $1" >&2
}

#
# Clones the skills repository, or brings an existing clone up to date.
#
function fetch_skills()
{
	if [[ -d "$skills_dir/.git" ]]; then
		git -C "$skills_dir" pull --ff-only --quiet || return $?
	else
		mkdir -p "${skills_dir%/*}" || return $?
		git clone --quiet --depth 1 "$skills_url" "$skills_dir" || return $?
	fi
}

#
# Symlinks every skill into the agent folders, with the repository's own
# script, which leaves real folders and foreign symlinks alone.
#
function link_skills()
{
	"$skills_dir/bin/link" || return $?
}

#
# Installs Task with go install, unless it is on PATH.
#
function install_task()
{
	command -v task >/dev/null && return
	go install github.com/go-task/task/v3/cmd/task@latest || return $?
	echo "export PATH=\"\$PATH:$gobin\"" >> "${CLAUDE_ENV_FILE:-/dev/null}"
}

#
# Installs shellcheck from the distribution, unless it is on PATH.
#
function install_shellcheck()
{
	command -v shellcheck >/dev/null && return
	apt-get install -y -q shellcheck >/dev/null && return
	apt-get update -q >/dev/null || return $?
	apt-get install -y -q shellcheck >/dev/null || return $?
}

[[ "${CLAUDE_CODE_REMOTE:-}" == "true" ]] || exit 0

log "Linking the dataverket skills from $skills_dir ..."
fetch_skills       || error "Could not fetch $skills_url!"
link_skills        || error "Could not link the skills!"

log "Installing the tools for task ci and task lint ..."
install_task       || error "Could not install Task!"
install_shellcheck || error "Could not install shellcheck!"

exit 0
