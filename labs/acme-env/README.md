# acme-env

Split DNS and an internal ACME certificate authority for local clusters,
running in podman and driven by Task. It's the runnable version of the
design in the `claude/split-dns-acme-wildcard-Whz6b` branch: CoreDNS for
split DNS, and step-ca issuing per-name certificates over ACME HTTP-01. It
uses no wildcard certificates and no RFC 2136.

```sh
task tools               # once, from the repo root: brew bundle (podman, step, dig, …)
task acme-env:up         # start etcd + CoreDNS + step-ca (creates the CA on first run)
task acme-env:smoke      # prove it: issue smoke.lab.test over HTTP-01 (binds port 80)
task acme-env:trust      # trust the lab root CA on this machine
```

[`../talos-cluster`](../talos-cluster) is the downstream consumer: a Talos
Kubernetes cluster where external-dns publishes into this DNS and
cert-manager gets certificates from this CA.

## What runs

All three containers use host networking on unprivileged ports, set in
[`../lab.env`](../lab.env):

| container | port | role |
|---|---|---|
| `acme-env-etcd` | 23790 | record store; external-dns writes here (SkyDNS format) |
| `acme-env-dns` | 1053 | CoreDNS: serves `lab.test` from etcd, pins `ca.lab.test` to the host, forwards everything else upstream |
| `acme-env-ca` | 8443 | step-ca with an ACME provisioner; resolves names through CoreDNS when validating |

```
external-dns ──writes──▶ etcd ◀──reads── CoreDNS ◀──resolves── step-ca ──HTTP-01──▶ your ingress :80
                                            ▲                     ▲
cert-manager ──────────── ACME order ───────┼─────────────────────┘
any client ── split DNS (lab.test) ─────────┘
```

The CA and its keys live in the `acme-env-ca` volume, and the records in
`acme-env-etcd`. `task down` keeps both. `task nuke` deletes them, after
which you'll need to re-trust the new root.

## Tasks

| task | does |
|---|---|
| `up` / `down` / `nuke` | start; stop (keep data); stop and delete the CA and records |
| `env` / `status` / `logs C=ca\|dns\|etcd` | settings, containers, logs |
| `root` / `trust` / `untrust` | export the root to `.run/root_ca.crt`; install it with `step`; remove it |
| `record:add NAME= IP=` / `record:rm NAME=` / `records` | manage records by hand (external-dns does this for clusters) |
| `dig NAME=` | resolve `NAME.lab.test` through the lab CoreDNS |
| `smoke` / `e2e` | issue a real certificate over HTTP-01 with lego; `up` then `smoke` |
| `ci` | render the Corefile (no podman needed) |

Everything is configured in `../lab.env`. `HOST_IP` is detected from the
default route; set it explicitly if that picks the wrong interface, or on
macOS, where podman runs in a VM. `PODMAN="sudo podman"` runs rootful,
which is what `smoke` needs if rootless podman can't bind port 80.

## Pointing your machine at it (optional)

Clusters use it directly. To resolve `lab.test` from your own shell and
browser as well, route just that zone to `127.0.0.1:1053`:

- **macOS:** `/etc/resolver/lab.test` containing `nameserver 127.0.0.1` and `port 1053`
- **Linux (systemd-resolved):** `/etc/systemd/resolved.conf.d/lab.conf`
  ```ini
  [Resolve]
  DNS=127.0.0.1:1053
  Domains=~lab.test
  ```
  then `sudo systemctl restart systemd-resolved`

## Decisions

These are carried over from the design doc.

**Why split DNS.** Some names (the lab zone) should resolve through an
internal resolver while everything else uses normal upstream DNS, without
editing `/etc/hosts` per project.

**Why CoreDNS rather than dnsmasq or Unbound.** It's the same engine and
`Corefile` as Kubernetes' in-cluster DNS. `forward` makes split DNS one
block per zone, and its `etcd` plugin gives external-dns a supported
write path through the coredns provider. dnsmasq is too limited, and
Unbound is recursive only.

**Why no wildcard certificates and no RFC 2136.** ACME wildcards need
DNS-01, which self-hosted means an RFC 2136/TSIG-capable authoritative
server (Knot or BIND) just for `_acme-challenge` records. Per-name
certificates over HTTP-01 avoid that:

- the blast radius is one service
- every issuance is auditable per hostname
- no shared wildcard key sits in cluster secrets

With cert-manager automating renewals, per-name certificates cost nothing
extra to run.

**Why step-ca and not a public CA.** Internal services only need to be
trusted by internal clients. step-ca speaks standard ACME, so cert-manager,
lego, certbot and acme.sh all work against it. It validates HTTP-01 and
TLS-ALPN-01 from inside the network, defaults to short-lived (24h)
certificates, and `step certificate install` handles trust on a laptop.

**Out of scope (future directions):**
- **Public, tenant-facing certificates** need a public CA. The best fit
  for "issue on first request" is Caddy on-demand TLS with an `ask`
  endpoint.
- **Gateway API with on-demand TLS:** no mainstream implementation does
  per-SNI issuance. Create `Certificate` resources at tenant signup instead,
  or put Caddy in front of the Gateway.
- **Public wildcards via RFC 2136:** if ever needed, pair CoreDNS with a
  tiny Knot instance authoritative only for `_acme-challenge.<zone>`,
  reached through a CNAME delegation.

## Changes from the original doc

- Everything runs as containers (podman, host network) driven by Task,
  instead of brew-installed daemons wrapped in launchd or systemd units.
- An etcd-backed zone was added so external-dns can publish records.
  CoreDNS still forwards everything else upstream.
- Upstream DNS is explicit (`1.1.1.1 9.9.9.9`), never `/etc/resolv.conf`,
  which avoids the systemd-resolved forwarding loop the doc warned about.
- The cluster example uses Traefik rather than ingress-nginx, which has
  been retired. See `../talos-cluster`.

## Operational notes

- Keep the `acme-env-ca` volume if you've trusted its root. Losing it means
  re-trusting on every machine.
- The CA's TLS certificate covers `ca.lab.test`, `HOST_IP`, `localhost` and
  `127.0.0.1`. If `HOST_IP` changes, run `task nuke up` and re-trust.
- Containers run with `--http-proxy=false`, so a corporate `HTTPS_PROXY`
  on the host never captures the lab's internal traffic.
