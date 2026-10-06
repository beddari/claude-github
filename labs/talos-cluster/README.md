# talos-cluster

A small [Talos](https://www.talos.dev) Kubernetes cluster in Docker that
consumes [acme-env](../acme-env). Its Kubernetes CA is signed by acme-env's
root. One Ingress shows the whole chain: external-dns publishes its name in
acme-env's DNS, and cert-manager gets its certificate from acme-env's CA. A
second one is served by a wildcard certificate, ordered over DNS-01.

## What you get

- Talos v1.14 with Kubernetes 1.37: one control plane and one worker, as
  Docker containers on the network `acme-lab`, `10.5.0.0/24`.
- A Kubernetes CA that acme-env signs with its root: the API server's
  certificate and the client certificates of the cluster chain to the lab's
  root. The CAs of the Talos API and of etcd stay Talos' own.
- Traefik as the ingress controller, on the worker's ports 80 and 443.
- external-dns, which writes the names of Ingresses into acme-env's etcd.
- cert-manager with a `ClusterIssuer` named `step-ca`, which orders
  certificates from acme-env over ACME: HTTP-01 for names, DNS-01 for the
  wildcard `*.lab.test`.
- whoami, a demo app at `https://whoami.lab.test`, with a DNS record and a
  certificate made the way any other app's would be.
- hello, at `https://hello.lab.test`, served by the wildcard certificate.

| Service | Address | Port | Login |
|---|---|---|---|
| Kubernetes API | `https://127.0.0.1`, a port Docker picks | random | `.run/kubeconfig` |
| Talos API | `10.5.0.2` | 50000 | `.run/talosconfig` |
| Ingress, Traefik | `10.5.0.3`, the worker | 80, 443 | none |
| whoami | `https://whoami.lab.test` | 443 | none |
| hello, the wildcard | `https://hello.lab.test` | 443 | none |

## Run it

You need Docker, talosctl, kubectl, helm, jq, curl and `dig`. On macOS and
Linux, `task tools` at the root of the repository installs them from the
Brewfile; Docker Engine comes from your distribution, or from colima on
macOS. acme-env must be up. The two nodes take 2 CPUs and 2 GiB of memory
each. On Linux the kernel module `br_netfilter` must be loaded, or a pod
cannot reach a service whose pod is on its own node; `task talos-cluster:up`
warns when it is not:

```sh
sudo modprobe br_netfilter
```

```sh
task acme-env:up            # the DNS and CA this cluster consumes
task talos-cluster:up       # about 3 minutes; ends with verify
export KUBECONFIG=$PWD/labs/talos-cluster/.run/kubeconfig
```

`task talos-cluster:up` creates the cluster, installs the add-ons, the
issuer and whoami, and ends with the end-to-end test. These lines are from a
run on a GitHub runner, where `up` took 3 minutes 10 seconds:

```
>>> A pod on acme-lab-controlplane-1 resolves ca.lab.test
>>> A pod on acme-lab-worker-1 resolves ca.lab.test
certificate.cert-manager.io/whoami-tls condition met
>>> whoami.lab.test -> 10.5.0.3, published by external-dns
curl: (60) SSL certificate problem: self-signed certificate
Hostname: whoami-848b9bdb5-mc4zq
Host: whoami.lab.test
issuer=O = Lab Internal CA, CN = Lab Internal CA Intermediate CA
notAfter=Oct  7 12:48:51 2026 GMT
>>> All good: whoami.lab.test has a record from external-dns and a step-ca cert.
```

The `curl` error is Traefik still serving its own certificate; `verify`
retries until it serves the new one, here 2 seconds later. From the Ingress
to an issued certificate takes about 20 seconds.

| Task | What it does |
|---|---|
| `task talos-cluster:up` | Run every step below, then `verify` |
| `task talos-cluster:verify` | The end-to-end test: the API server's chain to the lab's root, DNS from a pod on every node, both certificates, DNS records, HTTPS trusted by the root |
| `task talos-cluster:status` | Show the nodes, the lab's workloads and the certificates |
| `task talos-cluster:debug` | Print Helm, pods, events, logs and the cluster's DNS, for when `up` or `verify` fails |
| `task talos-cluster:down` | Remove the demo and its DNS records, then destroy the cluster and its CA |
| `task talos-cluster:lint` | Run shellcheck on the scripts |

The steps of `up` are tasks of their own, to run one again:

| Task | What it does |
|---|---|
| `task talos-cluster:create` | Have acme-env sign the cluster's CA, then create the Talos cluster in Docker with it; nothing if it exists |
| `task talos-cluster:kubeconfig` | Write `.run/kubeconfig` and wait for the nodes |
| `task talos-cluster:dns-forward` | Make the cluster's CoreDNS forward `lab.test` to acme-env |
| `task talos-cluster:addons` | Install Traefik, external-dns and cert-manager |
| `task talos-cluster:issuer` | Create the `ClusterIssuer` step-ca and the TSIG secret of its DNS-01 solver |
| `task talos-cluster:demo` | Deploy whoami with an Ingress, the wildcard certificate and hello |

## Use it

Any Ingress in the zone gets a record and a certificate. Give it a host in
`lab.test`, the class `traefik`, a TLS secret and the issuer:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  annotations:
    cert-manager.io/cluster-issuer: step-ca
spec:
  ingressClassName: traefik
  tls:
    - hosts: [myapp.lab.test]
      secretName: myapp-tls
  rules:
    - host: myapp.lab.test
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service: {name: myapp, port: {number: 80}}
```

[`manifests/whoami.yaml.tmpl`](manifests/whoami.yaml.tmpl) is a whole app
in this shape. [../docs/access.md](../docs/access.md) shows how to reach the
cluster with kubectl and talosctl.

For many names behind one certificate, use the wildcard. The demo orders it
once, in the namespace `whoami`, and any Ingress in that namespace can
serve it without the annotation:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wildcard
spec:
  secretName: wildcard-tls
  dnsNames: ["*.lab.test"]
  issuerRef: {kind: ClusterIssuer, name: step-ca}
```

The issuer solves `*.lab.test` over DNS-01 and every other name over
HTTP-01. A Secret is visible to its own namespace only: another namespace
orders its own copy of the wildcard, or uses per-name certificates.
[`manifests/wildcard.yaml.tmpl`](manifests/wildcard.yaml.tmpl) is the
demo.

To trust the cluster's API with the lab's root alone, a client needs the
cluster's CA as an intermediate: the API server sends only its own
certificate.

```sh
openssl verify -CAfile ../acme-env/.run/root_ca.crt -untrusted .run/ca/ca.crt \
	.run/apiserver.crt
```

`task talos-cluster:verify` saves `.run/apiserver.crt` and runs this check.

## Settings

From the environment, with [`../lab.env`](../lab.env) for what the two labs
share:

| Variable | Default | Meaning |
|---|---|---|
| `CLUSTER` | `acme-lab` | Name of the cluster, its Docker network and containers |
| `SUBNET` | `10.5.0.0/24` | The Docker network; the control plane is `.2`, the first worker `.3` |
| `WORKERS` | `1` | Number of workers |
| `DOCKER` | `docker` | The Docker command |
| `TALOS_ARGS` | empty | More arguments for `talosctl cluster create docker`, such as `--config-patch @file.yaml` |

```sh
WORKERS=2 task talos-cluster:up
```

Traefik's Ingress address is the first worker, `.3`. A host without IPv6
needs the hidden `--disable-ipv6` in `TALOS_ARGS`.

## How it is built

```
Taskfile.yml                 the tasks; each one calls a script in bin/
bin/                         one script per task, and functions.sh with the chart versions
values/                      Helm values: traefik.yaml, external-dns.yaml, cert-manager.yaml
manifests/*.yaml.tmpl        the ClusterIssuer, whoami and the wildcard; bin/render fills in the zone and the root
.run/                        made by the tasks: kubeconfig, talosconfig, cluster state, manifests
.run/ca/                     the cluster's CA, signed by acme-env, and the Talos config patches that set it
```

| Document | Content |
|---|---|
| [../docs/architecture.md](../docs/architecture.md) | Diagram of both labs, their networks, ports and flows |
| [../docs/access.md](../docs/access.md) | Reach each service: DNS, etcd, the CA, the cluster |
| [../docs/troubleshooting.md](../docs/troubleshooting.md) | Where to look, known behaviour, why it is built this way, what is tested |
| [../../docs/updates.md](../../docs/updates.md) | How the pinned versions are kept current |
