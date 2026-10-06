# shellcheck shell=bash disable=SC2153
# SC2153: ZONE, DNS_PORT and the rest are set in lab.env.
#
# Shared by the labs: logging, the settings in lab.env and the host address.
# Sourced by each lab's bin/functions.sh, never run.
#

lab_dir="${BASH_SOURCE[0]%/*}"

# shellcheck source=labs/lab.env
source "$lab_dir/lab.env" || return $?

zone="$ZONE"
dns_port="$DNS_PORT"
ca_port="$CA_PORT"
etcd_port="$ETCD_PORT"
etcd_peer_port="$ETCD_PEER_PORT"
dns_update_port="$DNS_UPDATE_PORT"
# The TSIG key that signs DNS-01 updates, by name. Its secret is made by
# acme-env in acme-env/.run/tsig.secret.
tsig_name="lab-dns01"
upstream_dns="$UPSTREAM_DNS"

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
# Prints a warning message.
#
function warn()
{
	if [[ -t 2 ]]; then
		echo -e "\x1b[1m\x1b[33m***\x1b[0m \x1b[1m$1\x1b[0m" >&2
	else
		echo "*** $1" >&2
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
# Prints the address the labs advertise: HOST_IP from lab.env, or else the
# source address of the default route.
#
function lab_host_ip()
{
	if   [[ -n "$HOST_IP" ]]; then
		echo "$HOST_IP"
	elif command -v ip >/dev/null; then
		ip -4 route get 1.1.1.1 |
			awk '{ for (i = 1; i < NF; i++) if ($i == "src") print $(i + 1) }'
	elif hostname -I >/dev/null 2>&1; then
		read -r first _ <<< "$(hostname -I)"
		echo "$first"
	else
		# macOS: podman runs in a VM; set HOST_IP in lab.env.
		ipconfig getifaddr en0
	fi
}

#
# Tells whether something accepts TCP connections on a host and port.
#
function answers()
{
	local host="$1"
	local port="$2"

	(exec 3<>"/dev/tcp/$host/$port") 2>/dev/null
}
