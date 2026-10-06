# shellcheck shell=bash
#
# Shared by the scripts in bin/. Sourced, never run.
#

root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd)" || return $?

# shellcheck source=labs/lab.sh
source "$root/../lab.sh" || return $?

cluster="${CLUSTER:-acme-lab}"
subnet="${SUBNET:-10.5.0.0/24}"
workers="${WORKERS:-1}"
docker="${DOCKER:-docker}"

# Extra arguments for "talosctl cluster create docker", for example
# TALOS_ARGS="--config-patch @my.yaml". A host without IPv6 needs the
# hidden "--disable-ipv6".
read -r -a talos_args <<< "${TALOS_ARGS:-}"

net="${subnet%.*}"
cp_ip="$net.2"
worker_ip="$net.3"
host_ip="$(lab_host_ip)"
run_dir="$root/.run"
root_ca="$root/../acme-env/.run/root_ca.crt"
# The cluster's Kubernetes CA, signed by acme-env's root.
ca_dir="$run_dir/ca"
# The secret of the TSIG key that cert-manager signs DNS-01 updates with.
tsig_file="$root/../acme-env/.run/tsig.secret"

cert_manager_version="v1.21.2"
external_dns_version="1.23.0"
traefik_version="41.6.1"

export KUBECONFIG="$run_dir/kubeconfig"
export TALOSCONFIG="$run_dir/talosconfig"

#
# Tells whether the cluster's control plane container exists.
#
function cluster_exists()
{
	"$docker" inspect "$cluster-controlplane-1" >/dev/null 2>&1
}

#
# Fails unless a command is on PATH.
#
function need()
{
	local tool

	for tool in "$@"; do
		command -v "$tool" >/dev/null ||
			fail "$tool is missing: task tools (brew bundle)"
	done
}

#
# Retries a command every two seconds, up to a number of tries.
#
function retry()
{
	local tries="$1"
	shift
	local i

	for i in $(seq "$tries"); do
		"$@" && return
		sleep 2
	done
	return 1
}
