# acme-env

Split DNS and an internal certificate authority (CA) for local clusters. Three
containers in podman on your computer: CoreDNS answers for a lab zone,
`lab.test`, and step-ca issues certificates for names in it over ACME. A
cluster, such as [talos-cluster](../talos-cluster), publishes its names in
the zone and gets its certificates from the CA.

## What you get

- A DNS zone, `lab.test`, whose records anyone can add: external-dns in a
  cluster, or you with `task acme-env:record-add`. Names outside the zone are
  forwarded to public resolvers.
- An ACME CA, the "Lab Internal CA", that issues a certificate for one name
  at a time over HTTP-01. No wildcard certificates.
- A root certificate to trust on your computer and in clusters.

| Service | Address | Port | Login |
|---|---|---|---|
| DNS, CoreDNS | your host address | 1053, UDP and TCP | none |
| Record store, etcd | your host address | 23790 | none |
| ACME directory, step-ca | `https://ca.lab.test:8443/acme/acme/directory` | 8443 | none; clients trust the root |

The host address is the source address of your default route, or `HOST_IP`
in [`../lab.env`](../lab.env). Containers and cluster nodes reach the
services there.

## Run it

You need podman, Task and, for the tasks that test DNS, `dig`. On macOS and
Linux, `task tools` at the root of the repository installs them from the
Brewfile. The ports 1053, 8443, 23790 and 23800 must be free, and port 80
for `task acme-env:check`.

```sh
task acme-env:up        # about 4 seconds; the first run pulls 4 images
task acme-env:check     # the end-to-end test: about 6 seconds
task acme-env:trust     # trust the root on this computer (asks for sudo)
```

`task acme-env:up` ends with the services and whether each one answers:

```
>>> Services of acme-env, zone lab.test, on 192.0.2.2:
  DNS              192.0.2.2:1053 (UDP and TCP)                  up    task acme-env:dig -- ca
  etcd             http://192.0.2.2:23790                        up    task acme-env:records
  ACME directory   https://192.0.2.2:8443/acme/acme/directory    up    task acme-env:check
  Root certificate /home/you/claude-github/labs/acme-env/.run/root_ca.crt
```

`task acme-env:check` puts `smoke.lab.test` in the zone, gets a certificate
for it with the ACME client lego, which listens on port 80 for the
challenge, and verifies the certificate against the root:

```
>>> Getting a certificate for smoke.lab.test over HTTP-01 ...
INFO  The server validated our request. domain=smoke.lab.test
INFO  Server responded with a certificate. domains=smoke.lab.test
X.509v3 TLS Certificate (ECDSA P-256) [Serial: 2642...6692]
  Subject:     smoke.lab.test
  Issuer:      Lab Internal CA Intermediate CA
  Provisioner: acme
>>> All good: step-ca issued smoke.lab.test over HTTP-01 through CoreDNS.
```

Rootless podman cannot bind port 80. Run the test rootful with
`PODMAN="sudo podman" task acme-env:check`, or let users bind low ports with
`sudo sysctl net.ipv4.ip_unprivileged_port_start=80`.

| Task | What it does |
|---|---|
| `task acme-env:up` | Start etcd, CoreDNS and step-ca; create the CA on the first run |
| `task acme-env:status` | List the services and whether each one answers |
| `task acme-env:check` | Run the end-to-end test |
| `task acme-env:root` | Copy the root certificate to `.run/root_ca.crt` and print its fingerprint |
| `task acme-env:trust` | Add the root to this computer's trust store; `untrust` removes it |
| `task acme-env:record-add -- NAME IP` | Add an A record, `NAME.lab.test` |
| `task acme-env:record-rm -- NAME` | Remove a name and everything under it |
| `task acme-env:records` | List every record in etcd |
| `task acme-env:dig -- NAME` | Resolve `NAME.lab.test` through the lab's CoreDNS |
| `task acme-env:logs -- ca` | Follow a container's log: `ca`, `dns` or `etcd` |
| `task acme-env:down` | Stop the containers, keep the CA and the records |
| `task acme-env:destroy` | Delete the containers, the CA and the records |
| `task acme-env:lint` | Run shellcheck on the scripts |

`task acme-env:down` and `task acme-env:up` keep the CA, its root and the
records: they live in the podman volumes `acme-env-ca` and `acme-env-etcd`.
`task acme-env:destroy` deletes them. The next `up` makes a new CA, whose
root you trust again.

## Use it

Get a certificate with any ACME client: point it at the ACME directory and
make it trust the root.

```sh
task acme-env:record-add -- myapp 192.0.2.2         # myapp.lab.test
sudo LEGO_CA_CERTIFICATES=labs/acme-env/.run/root_ca.crt \
	lego run --server https://192.0.2.2:8443/acme/acme/directory \
	--accept-tos --email me@lab.test --domains myapp.lab.test --http
```

Use your host address in place of `192.0.2.2`. lego answers the challenge on
port 80, hence `sudo`. lego reads the root from `LEGO_CA_CERTIFICATES`;
certbot from
`REQUESTS_CA_BUNDLE`; cert-manager from `caBundle` in its issuer, as
[talos-cluster](../talos-cluster) does. To resolve `lab.test` from your own
shell and browser, route the zone to `127.0.0.1:1053`;
[docs/access.md](../docs/access.md) has the steps for macOS and Linux.

## Settings

In [`../lab.env`](../lab.env), shared with talos-cluster. Each one is a
default, and the environment overrides it: `ZONE=dev.test task acme-env:up`.

| Variable | Default | Meaning |
|---|---|---|
| `ZONE` | `lab.test` | The DNS zone, and the names the CA issues for |
| `DNS_PORT` | `1053` | CoreDNS, UDP and TCP |
| `CA_PORT` | `8443` | step-ca |
| `ETCD_PORT` | `23790` | etcd's client port; `ETCD_PEER_PORT`, `23800`, is local only |
| `UPSTREAM_DNS` | `1.1.1.1 9.9.9.9` | Resolvers for names outside the zone |
| `HOST_IP` | the source address of the default route | The address the services are reached on |
| `PODMAN` | `podman` | The podman command; `sudo podman` runs everything rootful |

The CA's TLS certificate names `ca.lab.test`, `HOST_IP`, `localhost` and
`127.0.0.1`, as they were when the CA was made. After a change of `ZONE` or
`HOST_IP`, run `task acme-env:destroy` and `task acme-env:up`.

## How it is built

```
Taskfile.yml        the tasks; each one calls a script in bin/
bin/                one script per task, and functions.sh with the image versions
Corefile.tmpl       CoreDNS: the zone from etcd, ca.<zone> pinned, the rest forwarded
../lab.env          the settings shared with talos-cluster
../lab.sh           logging, the settings and the host address, for both labs
.run/               made by the tasks: the Corefile, the root, lego's files
```

| Document | Content |
|---|---|
| [../docs/architecture.md](../docs/architecture.md) | Diagram of both labs, their networks, ports and flows |
| [../docs/access.md](../docs/access.md) | Reach each service: DNS, etcd, the CA, the cluster |
| [../docs/troubleshooting.md](../docs/troubleshooting.md) | Where to look, known behaviour, why it is built this way, what is tested |
| [../../docs/updates.md](../../docs/updates.md) | How the pinned versions are kept current |

The design comes from the branch `claude/split-dns-acme-wildcard-Whz6b`,
which compared CoreDNS with dnsmasq and Unbound, and HTTP-01 with wildcard
certificates over DNS-01. [../docs/troubleshooting.md](../docs/troubleshooting.md)
keeps the reasons.
