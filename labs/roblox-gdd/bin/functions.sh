# shellcheck shell=bash
#
# Shared by the scripts in bin/. Sourced, never run.
#

lab="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd)" || return $?
repo="$(cd "$lab/../.." && pwd)" || return $?
template="$lab/template"
run_dir="$lab/.run"
tools_dir="$run_dir/bin"
defs_file="$run_dir/globalTypes.d.luau"

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
# Prints the "owner/repo version" that template/rokit.toml pins for a tool.
#
function pinned()
{
	local tool="$1"

	sed -n "s#^$tool = \"\\([^@\"]*\\)@\\([^\"]*\\)\"#\\1 \\2#p" \
		"$template/rokit.toml"
}

#
# Prints the release asset name of a tool for this computer.
#
function asset_name()
{
	local tool="$1"
	local version="$2"
	local os arch

	case "$(uname -s)" in
		Linux)	os="linux" ;;
		Darwin)	os="macos" ;;
		*)	return 1 ;;
	esac
	case "$(uname -m)" in
		x86_64|amd64)	arch="x86_64" ;;
		arm64|aarch64)	arch="aarch64" ;;
		*)		return 1 ;;
	esac

	case "$tool" in
		rojo)		echo "rojo-$version-$os-$arch.zip" ;;
		stylua)		echo "stylua-$os-$arch.zip" ;;
		luau-lsp)
			if [[ "$os" == "macos" ]]; then
				echo "luau-lsp-macos.zip"
			elif [[ "$arch" == "aarch64" ]]; then
				echo "luau-lsp-linux-arm64.zip"
			else
				echo "luau-lsp-linux-x86_64.zip"
			fi
			;;
		*)		return 1 ;;
	esac
}
