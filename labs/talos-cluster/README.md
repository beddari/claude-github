# talos-cluster

A small [Talos](https://www.talos.dev) Kubernetes cluster in Docker that
consumes [`../acme-env`](../acme-env). One Ingress shows the whole chain:

- **external-dns** (coredns provider) publishes the hostname into acme-env's
  etcd, so acme-env's CoreDNS answers for it.
- **cert-manager** orders a certificate from acme-env's step-ca over ACME
  HTTP-01, solved through Traefik.
- **Traefik** listens on the worker's ports 80 and 443 and serves the app
  with that certificate.

```sh
task tools                  # once, from the repo root: brew bundle
task acme-env:up            # the DNS + CA this cluster consumes
task talos-cluster:up       # cluster + Traefik + external-dns + cert-manager + issuer + demo, then verify
export KUBECONFIG=$PWD/labs/talos-cluster/.run/kubeconfig
task talos-cluster:down     # remove the demo (cleans its DNS), destroy the cluster
```

`up` finishes with `verify`, which checks the whole chain:

1. `Certificate/whoami-tls` reaches Ready, issued by step-ca.
2. `whoami.lab.test` resolves through acme-env's CoreDNS to the worker's IP,
   as published by external-dns.
3. `https://whoami.lab.test` answers from the host, verified against the lab
   root CA.

## How it's wired

| piece | setting | why |
|---|---|---|
| cluster | `talosctl cluster create docker`, 1 control plane + `WORKERS` (1), subnet `10.5.0.0/24` | Smallest real Talos setup; nodes are routable from the host on Linux |
| in-cluster DNS | CoreDNS gets a `lab.test:53 { forward . HOST_IP:1053 }` block | Pods (cert-manager's self-check, the ACME client) resolve lab names and `ca.lab.test` |
| ingress | Traefik DaemonSet with hostPorts 80/443, `ingressEndpoint.ip` = first worker | No LoadBalancer in Docker, so ingresses report the worker IP for external-dns to publish |
| DNS | external-dns `provider: coredns`, `ETCD_URLS=http://HOST_IP:23790`, `domainFilters: [lab.test]`, txt registry owned by the cluster name | Writes records acme-env serves; `policy: sync` removes them again |
| certificates | cert-manager with `ClusterIssuer/step-ca` (ACME, `caBundle` = lab root, HTTP-01 via the `traefik` class) | Per-name certificates from the internal CA, with no DNS-01 or wildcards |

Settings shared with acme-env (zone, ports, host IP) come from
[`../lab.env`](../lab.env). Chart versions are pinned in the Taskfile.

## Tasks

| task | does |
|---|---|
| `up` | everything below in order, then `verify` |
| `cluster:create` / `kubeconfig` | Talos in Docker; write `.run/kubeconfig` and wait for nodes |
| `dns:forward` | add the lab zone forward to cluster CoreDNS (idempotent) |
| `addons` / `issuer` / `demo` | Helm releases; the ClusterIssuer; whoami + Ingress |
| `verify` / `status` | check the chain; nodes, pods, ingresses, certificates |
| `down` | delete the demo, wait for external-dns to clean up, destroy the cluster |
| `ci` | render manifests and parse all YAML (no cluster needed) |

Variables: `CLUSTER` (`acme-lab`), `SUBNET`, `WORKERS`, `DOCKER`, and
`TALOS_ARGS` for extra `talosctl cluster create docker` arguments, such as
`--config-patch @file.yaml`. Hosts without IPv6 also need the hidden
`--disable-ipv6`.

## Notes

- **Linux:** node IPs (10.5.0.x) are reachable from the host directly, and
  so is acme-env from the nodes, through `HOST_IP`.
- **macOS:** Docker runs in a VM (colima, Docker Desktop), so 10.5.0.x isn't
  routable from the host. Run the labs inside the VM, or set `HOST_IP` to an
  address both sides can reach, and use `-p 80:80/tcp,443:443/tcp` in
  `TALOS_ARGS` with workers set to 0 to expose ingress on the control plane.
  This setup hasn't been tested yet.
- The Traefik namespace is labelled `pod-security.kubernetes.io/enforce=privileged`,
  because hostPorts aren't allowed under Talos' default baseline policy.
- The CoreDNS change lives in the `coredns` ConfigMap. `talosctl upgrade-k8s`
  may reset it; re-run `task talos-cluster:dns:forward` afterwards.
