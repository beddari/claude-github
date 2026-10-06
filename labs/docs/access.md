# Access

How to reach each service of the two labs. The commands run from the root
of the repository. `HOST_IP` is the address acme-env prints with
`task acme-env:status`.

## The CA's root certificate

Every client of the CA needs its root. `task acme-env:up` copies it to
`labs/acme-env/.run/root_ca.crt` and prints its fingerprint:

```sh
task acme-env:root
```

To trust it on this computer, for browsers and most tools:

```sh
task acme-env:trust        # step certificate install; asks for sudo
task acme-env:untrust      # removes it again
```

On Linux this adds the root to the system's trust store; on macOS, to the
keychain. A tool that keeps its own list needs the file: curl takes
`--cacert`, lego `LEGO_CA_CERTIFICATES`, Python's requests
`REQUESTS_CA_BUNDLE`, Node `NODE_EXTRA_CA_CERTS`.

A new CA, after `task acme-env:destroy`, has a new root. Run `untrust`
before `destroy`, and `trust` after the next `up`.

## The ACME directory

```
https://ca.lab.test:8443/acme/acme/directory
https://HOST_IP:8443/acme/acme/directory
```

The CA's TLS certificate names both. Use the first where `lab.test`
resolves, as in the cluster; the second everywhere else. Any ACME client
works. The CA validates HTTP-01 and TLS-ALPN-01 by resolving the name
through acme-env's DNS, so the name must be in the zone first: external-dns
adds it for an Ingress, `task acme-env:record-add -- NAME IP` by hand.

The step CLI talks to the CA too:

```sh
step ca health --ca-url https://HOST_IP:8443 \
	--root labs/acme-env/.run/root_ca.crt
```

## DNS

```sh
task acme-env:dig -- whoami                 # dig @127.0.0.1 -p 1053 whoami.lab.test
dig @HOST_IP -p 1053 +tcp whoami.lab.test   # the same, over TCP
```

To resolve `lab.test` from your own shell and browser, send only that zone
to the lab. Other names keep going where they went.

**macOS**: one file per zone in `/etc/resolver`:

```sh
sudo mkdir -p /etc/resolver
printf 'nameserver 127.0.0.1\nport 1053\n' | sudo tee /etc/resolver/lab.test
```

**Linux with systemd-resolved**: a drop-in that routes the zone:

```sh
sudo mkdir -p /etc/systemd/resolved.conf.d
printf '[Resolve]\nDNS=127.0.0.1:1053\nDomains=~lab.test\n' |
	sudo tee /etc/systemd/resolved.conf.d/lab.conf
sudo systemctl restart systemd-resolved
```

The `~` makes `lab.test` a routing domain only, so other lookups are not
affected.

## Records, in etcd

```sh
task acme-env:records                       # every key under /skydns
task acme-env:record-add -- myapp 10.5.0.3  # myapp.lab.test
task acme-env:record-rm -- myapp
```

A record is a key and a small JSON value, the format of CoreDNS's etcd
plugin: `myapp.lab.test` is `/skydns/test/lab/myapp`. What external-dns
wrote for whoami, on a GitHub runner:

```
/skydns/test/lab/a-whoami/19a3bf2a
{"text":"\"heritage=external-dns,external-dns/owner=acme-lab,external-dns/resource=ingress/whoami/whoami\"","targetstrip":1}
/skydns/test/lab/whoami/3494e992
{"host":"10.5.0.3","targetstrip":1}
```

The first is a TXT record that says which cluster owns the name. external-dns
only changes or deletes names that its own cluster owns, so records added by
hand are left alone.

## DNS-01 updates, in BIND

BIND listens on port 1054 and holds one zone, `_acme-challenge.lab.test`.
It takes an update only when it is signed with the TSIG key `lab-dns01`,
whose secret is in `labs/acme-env/.run/tsig.secret`. With BIND's
`nsupdate`:

```sh
printf 'key hmac-sha256:lab-dns01 %s\nserver 127.0.0.1 1054\nzone _acme-challenge.lab.test\nupdate add _acme-challenge.lab.test 60 TXT "hello"\nsend\n' \
	"$(cat labs/acme-env/.run/tsig.secret)" | nsupdate
dig @127.0.0.1 -p 1053 +short TXT _acme-challenge.lab.test   # "hello", through CoreDNS
```

The zone lives in memory and starts empty on every `task acme-env:up`.

## The cluster, with kubectl

```sh
export KUBECONFIG=$PWD/labs/talos-cluster/.run/kubeconfig
kubectl get nodes
kubectl get ingress,certificate -A
```

`task talos-cluster:status` shows the same and prints the `export` line. The
API is on a random port of `127.0.0.1`, which Docker forwards to the control
plane.

The kubeconfig trusts the cluster's CA, `labs/talos-cluster/.run/ca/ca.crt`,
which acme-env's root signed. Its client certificate comes from the same CA.
To check the chain:

```sh
openssl x509 -in labs/talos-cluster/.run/ca/ca.crt -noout -subject -issuer
```

## The nodes, with talosctl

```sh
export TALOSCONFIG=$PWD/labs/talos-cluster/.run/talosconfig
talosctl -n 10.5.0.2 dashboard
talosctl -n 10.5.0.3 logs kubelet
```

Talos has no shell. Every question to a node goes through its API on port
50000.

## whoami, over HTTPS

```sh
curl --cacert labs/acme-env/.run/root_ca.crt \
	--resolve whoami.lab.test:443:10.5.0.3 https://whoami.lab.test/
```

With the root trusted and the zone routed as above, `https://whoami.lab.test`
works in a browser as well.
