# Troubleshooting and design notes

## Where to look

| Question | Command |
|---|---|
| Does acme-env work end to end? | `task acme-env:check` |
| Does the cluster work end to end? | `task talos-cluster:verify` |
| Which acme-env service answers? | `task acme-env:status` |
| What does the CA say? | `task acme-env:logs -- ca` |
| Is a name in the zone? | `task acme-env:dig -- NAME`, then `task acme-env:records` |
| Why is the cluster stuck? | `task talos-cluster:debug`: Helm, pods, events, logs, the cluster's DNS |
| Why is a certificate not Ready? | `kubectl describe certificate,order,challenge -A` |
| What did external-dns do? | `kubectl -n external-dns logs deploy/external-dns` |
| What does a node say? | `talosctl -n 10.5.0.2 dmesg`, with `TALOSCONFIG` set |

## Known behaviour

**`task acme-env:check` cannot bind port 80.** Rootless podman cannot listen
below port 1024. Run `PODMAN="sudo podman" task acme-env:check`, or allow
it once with `sudo sysctl net.ipv4.ip_unprivileged_port_start=80`.

**The ClusterIssuer is not Ready: `lookup ca.lab.test: i/o timeout`.** The
cluster cannot resolve the lab zone. On Linux, first check that
`br_netfilter` is loaded (`task talos-cluster:debug` says so at the top); see
below. Then that acme-env is up, and that the cluster's Corefile has the
block from `task talos-cluster:dns-forward`. The last part of
`task talos-cluster:debug` asks acme-env's DNS from the cluster's network,
over UDP and over TCP. `task talos-cluster:verify` resolves `ca.lab.test`
from a pod on every node, and names the node that fails.

**DNS fails only from some pods.** Without the kernel module `br_netfilter`
on the host, a pod cannot reach a service whose pod runs on its own node.
On GitHub's runners, lookups through kube-dns timed out in the two runs where
both CoreDNS pods were on the worker with cert-manager, and worked in the run
where they were on the control plane. Load it with `sudo modprobe br_netfilter`
before `task talos-cluster:up`; the labs workflow does.

**HTTPS shows Traefik's default certificate for a few seconds.** Traefik
loads a new TLS Secret a few seconds after cert-manager writes it. Until
then it serves its own self-signed certificate. `task talos-cluster:verify`
retries for up to a minute.

**Traefik logs `secret whoami/whoami-tls does not exist`.** Normal while the
certificate is being ordered; it stops once the Secret is there.

**lego's files in `.run/lego` belong to root.** With `PODMAN="sudo podman"`
lego runs as root and writes its key as root only. The tasks read and remove
them through a container; to look yourself, use `sudo`.

**The CA changed address.** The CA's TLS certificate names `HOST_IP` as it
was when `task acme-env:up` first made the CA. After a move to another
network, clients refuse it. Run `task acme-env:destroy`, `task acme-env:up`
and trust the new root.

**Stale records after a cluster is gone.** `task talos-cluster:down` deletes
whoami first and waits for external-dns to remove its records. A cluster
destroyed another way leaves them; `task acme-env:record-rm -- whoami`
removes them.

**`talosctl cluster create` fails on `net.ipv6.conf.all.disable_ipv6`.** The
host has no IPv6. Add the hidden flag:
`TALOS_ARGS=--disable-ipv6 task talos-cluster:up`.

## Why it is built this way

**Split DNS with CoreDNS.** Names in the lab zone resolve through the lab,
everything else as before, without `/etc/hosts`. CoreDNS is the same engine
and `Corefile` as Kubernetes' own DNS, `forward` makes the split one block
per zone, and its `etcd` plugin gives external-dns a supported way in.
dnsmasq does too little and Unbound only recurses.

**Certificates per name over HTTP-01, not wildcards.** A wildcard needs
DNS-01, and a self-hosted DNS-01 needs a server that takes RFC 2136 updates
with TSIG, such as Knot or BIND, beside CoreDNS. Per-name certificates need
none of that, limit a leaked key to one service, and leave one entry per
name in the CA's log. With cert-manager renewing them, there are no more of
them to look after.

**step-ca, not a public CA.** Only internal clients need to trust these
certificates. step-ca speaks standard ACME, runs inside the network, so it
can reach names a public CA cannot, and issues for 24 hours by default.

**acme-env on the host network.** The CA's certificate, the DNS answers and
external-dns all use `HOST_IP`, which works the same from the host, from
containers and from the cluster. With a container network, each would need a
different address.

**step-ca resolves through the lab's CoreDNS.** `--resolver 127.0.0.1:1053`
makes it find names in the zone that no public resolver knows.

**Containers without the host's proxy.** `--http-proxy=false` keeps a
corporate `HTTPS_PROXY` out of the lab's containers. Their traffic is all
local, and a proxy would carry it away.

**Traefik on hostPorts, with the worker as the Ingress address.** Talos in
Docker has no LoadBalancer. hostPorts put Traefik on the worker's own ports
80 and 443, and `ingressEndpoint.ip` gives every Ingress that address, which
external-dns publishes. The namespace allows hostPorts with the label
`pod-security.kubernetes.io/enforce=privileged`.

**`kubectl rollout status`, not `helm --wait`.** On a GitHub runner `helm
--wait` waited 5 minutes for Traefik while its pod had been Ready for 2
seconds. Waiting on the workloads themselves says what is not ready.

**Traefik, not ingress-nginx.** ingress-nginx is retired.

## What is tested

The `labs` workflow runs on GitHub's `ubuntu-latest` runners for every
pull request that changes the labs, but not their docs: `task acme-env:up`, `task acme-env:check`,
`task talos-cluster:up` with its `verify`, then `down` for both. A run takes
about 6 minutes. acme-env and `task acme-env:check` also run on a Linux host
with rootful podman.

Not tested:

- macOS. podman and Docker run in VMs there: `HOST_IP` must be set to an
  address that both VMs reach, and the worker's address, `10.5.0.3`, is not
  reachable from the Mac itself.
- Rootless podman with `task acme-env:check`.
- arm64.
- More than one worker. Ingresses would still point at the first.
- Upgrades of Talos or the charts in a running cluster; `down` and `up` is
  the tested way.
