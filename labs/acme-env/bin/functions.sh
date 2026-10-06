# shellcheck shell=bash
#
# Shared by the scripts in bin/. Sourced, never run.
#

root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd)" || return $?

# shellcheck source=labs/lab.sh
source "$root/../lab.sh" || return $?

name="acme-env"
run_dir="$root/.run"
ca_name="Lab Internal CA"

img_step_ca="docker.io/smallstep/step-ca:0.30.2"
img_coredns="docker.io/coredns/coredns:1.14.7"
img_etcd="quay.io/coreos/etcd:v3.7.2"
img_lego="docker.io/goacme/lego:v5.5.2"

# PODMAN="sudo podman" runs everything rootful, as CI does for port 80.
read -r -a podman <<< "${PODMAN:-podman}"

host_ip="$(lab_host_ip)"
acme_url="https://$host_ip:$ca_port/acme/acme/directory"

#
# Runs podman, or what PODMAN names.
#
function pod()
{
	"${podman[@]}" "$@"
}

#
# Runs a throwaway container with the CA's volume, as its own user.
#
function in_ca()
{
	pod run --rm -v "$name-ca:/home/step" "$img_step_ca" "$@"
}

#
# Runs a shell command as root in a throwaway step-ca container that sees
# .run/ at /lab. Files lego wrote with rootful podman are root's.
#
function in_run_dir()
{
	pod run --rm --user 0 --entrypoint sh -v "$run_dir:/lab:Z" \
		"$img_step_ca" -ec "$1"
}

#
# Runs etcdctl in the etcd container.
#
function etcd()
{
	pod exec "$name-etcd" etcdctl \
		--endpoints "http://127.0.0.1:$etcd_port" "$@"
}

#
# Prints the etcd key of a name in the zone, the way the CoreDNS etcd plugin
# and external-dns write it: smoke.lab.test is /skydns/test/lab/smoke.
#
function record_key()
{
	local fqdn="$1.$zone"
	local key=""
	local part

	while [[ -n "$fqdn" ]]; do
		part="${fqdn##*.}"
		key="$key/$part"
		[[ "$fqdn" == "$part" ]] && break
		fqdn="${fqdn%.*}"
	done
	echo "/skydns$key"
}

#
# Tells whether step-ca answers on its health endpoint.
#
function ca_healthy()
{
	curl -fsk --noproxy '*' "https://127.0.0.1:$ca_port/health" \
		>/dev/null 2>&1
}
