# acme-env

Split DNS and an internal certificate authority (CA) for local clusters. Four
containers in podman on your computer: CoreDNS answers for a lab zone,
`lab.test`, step-ca issues certificates for names in it over ACME, and BIND
takes the DNS-01 challenges that a wildcard certificate needs. A
cluster, such as [talos-cluster](../talos-cluster), publishes its names in
the zone and gets its certificates from the CA.

## What you get

- A DNS zone, `lab.test`, whose records anyone can add: external-dns in a
  cluster, or you with `task acme-env:record-add`. Names outside the zone are
  forwarded to public resolvers.
- An ACME CA, the "Lab Internal CA", that issues certificates for names in
  the zone over HTTP-01, and for the wildcard `*.lab.test` over DNS-01.
- A root certificate to trust on your computer and in clusters.
- A CA for each cluster, signed by the root: `task acme-env:cluster-ca`.
  [talos-cluster](../talos-cluster) makes its Kubernetes CA this way.

| Service | Address | Port | Login |
|---|---|---|---|
| DNS, CoreDNS | your host address | 1053, UDP and TCP | none |
| Record store, etcd | your host address | 23790 | none |
| DNS-01 updates, BIND | your host address | 1054, UDP and TCP | TSIG key `lab-dns01`, secret in `.run/tsig.secret` |
| ACME directory, step-ca | `https://ca.lab.test:8443/acme/acme/directory` | 8443 | none; clients trust the root |

The host address is the source address of your default route, or `HOST_IP`
in [`../lab.env`](../lab.env). Containers and cluster nodes reach the
services there.

## Run it

You need podman, Task and, for the tasks that test DNS, `dig`. On macOS and
Linux, `task tools` at the root of the repository installs them from the
Brewfile. The ports 1053, 1054, 8443, 23790 and 23800 must be free, and port
80 for `task acme-env:check`.

```sh
task acme-env:up        # about 5 seconds; the first run pulls 5 images
task acme-env:check     # the end-to-end test: about 6 seconds
task acme-env:trust     # trust the root on this computer (asks for sudo)
```

`task acme-env:up` ends with the services and whether each one answers:

```
>>> Services of acme-env, zone lab.test, on 192.0.2.2:
  DNS              192.0.2.2:1053 (UDP and TCP)                  up    task acme-env:dig -- ca
  DNS-01 updates   192.0.2.2:1054, TSIG key lab-dns01            up    task acme-env:check
  etcd             http://192.0.2.2:23790                        up    task acme-env:records
  ACME directory   https://192.0.2.2:8443/acme/acme/directory    up    task acme-env:check
  Root certificate /home/you/claude-github/labs/acme-env/.run/root_ca.crt
```

`task acme-env:check` gets two certificates with the ACME client lego, and
verifies each one against the root:

- `smoke.lab.test` over HTTP-01: it puts the name in the zone, and lego
  answers the challenge on port 80.
- `*.lab.test` over DNS-01: lego puts its token in BIND with an update
  signed with the TSIG key, and step-ca finds it through CoreDNS.

```
>>> Getting a certificate for smoke.lab.test over HTTP-01 ...
INFO  The server validated our request. domain=smoke.lab.test
INFO  Server responded with a certificate. domains=smoke.lab.test
X.509v3 TLS Certificate (ECDSA P-256) [Serial: 5877...0252]
  Subject:     smoke.lab.test
  Issuer:      Lab Internal CA Intermediate CA
  Provisioner: acme
>>> Getting a certificate for *.lab.test over DNS-01 ...
INFO  The server validated our request. domain=*.lab.test
INFO  Server responded with a certificate. domains=*.lab.test
X.509v3 TLS Certificate (ECDSA P-256) [Serial: 1851...5765]
  Subject:     *.lab.test
  Issuer:      Lab Internal CA Intermediate CA
  Provisioner: acme
>>> All good: step-ca issued smoke.lab.test over HTTP-01, *.lab.test over DNS-01.
```

Rootless podman cannot bind port 80. Run the test rootful with
`PODMAN="sudo podman" task acme-env:check`, or let users bind low ports with
`sudo sysctl net.ipv4.ip_unprivileged_port_start=80`.

| Task | What it does |
|---|---|
| `task acme-env:up` | Start etcd, BIND, CoreDNS and step-ca; create the CA on the first run |
| `task acme-env:status` | List the services and whether each one answers |
| `task acme-env:check` | Run the end-to-end test: HTTP-01 and the DNS-01 wildcard |
| `task acme-env:cluster-ca -- NAME DIR` | Sign a CA for a cluster with the root: `DIR/ca.crt` and `DIR/ca.key` |
| `task acme-env:root` | Copy the root certificate to `.run/root_ca.crt` and print its fingerprint |
| `task acme-env:trust` | Add the root to this computer's trust store; `untrust` removes it |
| `task acme-env:record-add -- NAME IP` | Add an A record, `NAME.lab.test` |
| `task acme-env:record-rm -- NAME` | Remove a name and everything under it |
| `task acme-env:records` | List every record in etcd |
| `task acme-env:dig -- NAME` | Resolve `NAME.lab.test` through the lab's CoreDNS |
| `task acme-env:logs -- ca` | Follow a container's log: `ca`, `dns`, `bind` or `etcd` |
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

A wildcard needs DNS-01. Give the ACME client BIND's address and the TSIG
key; lego calls it the `dnsupdate` provider, cert-manager `rfc2136`:

```sh
sudo LEGO_CA_CERTIFICATES=labs/acme-env/.run/root_ca.crt \
	DNSUPDATE_NAMESERVER=127.0.0.1:1054 DNSUPDATE_TSIG_KEY=lab-dns01 \
	DNSUPDATE_TSIG_SECRET="$(cat labs/acme-env/.run/tsig.secret)" \
	DNSUPDATE_TSIG_ALGORITHM=hmac-sha256. \
	lego run --server https://192.0.2.2:8443/acme/acme/directory \
	--accept-tos --email me@lab.test --domains '*.lab.test' \
	--dns dnsupdate --dns.resolvers 127.0.0.1:1053 \
	--dns.propagation.disable-ans
```

`--dns.propagation.disable-ans` makes lego check for its token through
CoreDNS only. Without it, lego asks the zone's name servers on port 53,
where BIND does not listen.

BIND holds the zone `_acme-challenge.lab.test` alone, which is where the
token for `lab.test` and `*.lab.test` goes. Any other name over DNS-01,
such as `myapp.lab.test`, has its token at `_acme-challenge.myapp.lab.test`,
which CoreDNS reads from etcd; use HTTP-01 for those.

A CA for a cluster:

```sh
task acme-env:cluster-ca -- mycluster /path/to/dir   # dir/ca.crt, dir/ca.key
```

Its certificate is signed by the root, so whatever it issues chains to the
root. It may sign certificates, not further CAs (`pathlen:0`), and is valid
for five years. A second run keeps the files in `DIR`.

## Settings

In [`../lab.env`](../lab.env), shared with talos-cluster. Each one is a
default, and the environment overrides it: `ZONE=dev.test task acme-env:up`.

| Variable | Default | Meaning |
|---|---|---|
| `ZONE` | `lab.test` | The DNS zone, and the names the CA issues for |
| `DNS_PORT` | `1053` | CoreDNS, UDP and TCP |
| `CA_PORT` | `8443` | step-ca |
| `ETCD_PORT` | `23790` | etcd's client port; `ETCD_PEER_PORT`, `23800`, is local only |
| `DNS_UPDATE_PORT` | `1054` | BIND, which takes DNS-01 updates, UDP and TCP |
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
Corefile.tmpl       CoreDNS: the zone from etcd, ca.<zone> pinned, _acme-challenge.<zone> to BIND
bind/               BIND's named.conf and its one zone, as templates
../lab.env          the settings shared with talos-cluster
../lab.sh           logging, the settings and the host address, for both labs
.run/               made by the tasks: the Corefile, BIND's files, the TSIG secret, the root, lego's files
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
