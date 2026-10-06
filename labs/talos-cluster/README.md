# talos-cluster

A small [Talos](https://www.talos.dev) Kubernetes cluster in Docker that
consumes [acme-env](../acme-env). One Ingress shows the whole chain:
external-dns publishes its name in acme-env's DNS, and cert-manager gets its
certificate from acme-env's CA.

## What you get

- Talos v1.14 with Kubernetes 1.37: one control plane and one worker, as
  Docker containers on the network `acme-lab`, `10.5.0.0/24`.
- Traefik as the ingress controller, on the worker's ports 80 and 443.
- external-dns, which writes the names of Ingresses into acme-env's etcd.
- cert-manager with a `ClusterIssuer` named `step-ca`, which orders
  certificates from acme-env over ACME HTTP-01.
- whoami, a demo app at `https://whoami.lab.test`, with a DNS record and a
  certificate made the way any other app's would be.

| Service | Address | Port | Login |
|---|---|---|---|
| Kubernetes API | `https://127.0.0.1`, a port Docker picks | random | `.run/kubeconfig` |
| Talos API | `10.5.0.2` | 50000 | `.run/talosconfig` |
| Ingress, Traefik | `10.5.0.3`, the worker | 80, 443 | none |
| whoami | `https://whoami.lab.test` | 443 | none |

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
task talos-cluster:up       # about 5 minutes; ends with verify
export KUBECONFIG=$PWD/labs/talos-cluster/.run/kubeconfig
```

`task talos-cluster:up` creates the cluster, installs the add-ons, the
issuer and whoami, and ends with the end-to-end test. These lines are from a
run on a GitHub runner:

```
certificate.cert-manager.io/whoami-tls condition met
dns: whoami.lab.test -> 10.5.0.3 (published by external-dns)
Hostname: whoami-848b9bdb5-4gkjq
Host: whoami.lab.test
issuer=O = Lab Internal CA, CN = Lab Internal CA Intermediate CA
notAfter=Oct  7 12:05:33 2026 GMT
```

From the Ingress to an issued certificate takes about 20 seconds.

| Task | What it does |
|---|---|
| `task talos-cluster:up` | Run every step below, then `verify` |
| `task talos-cluster:verify` | The end-to-end test: DNS from a pod on every node, certificate Ready, DNS record, HTTPS trusted by the lab's root |
| `task talos-cluster:status` | Show the nodes, the lab's workloads and the certificates |
| `task talos-cluster:debug` | Print Helm, pods, events, logs and the cluster's DNS, for when `up` or `verify` fails |
| `task talos-cluster:down` | Remove whoami and its DNS records, then destroy the cluster |
| `task talos-cluster:lint` | Run shellcheck on the scripts |

The steps of `up` are tasks of their own, to run one again:

| Task | What it does |
|---|---|
| `task talos-cluster:create` | Create the Talos cluster in Docker; nothing if it exists |
| `task talos-cluster:kubeconfig` | Write `.run/kubeconfig` and wait for the nodes |
| `task talos-cluster:dns-forward` | Make the cluster's CoreDNS forward `lab.test` to acme-env |
| `task talos-cluster:addons` | Install Traefik, external-dns and cert-manager |
| `task talos-cluster:issuer` | Create the `ClusterIssuer` step-ca |
| `task talos-cluster:demo` | Deploy whoami with an Ingress |

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
manifests/*.yaml.tmpl        the ClusterIssuer and whoami; bin/render fills in the zone and the root
.run/                        made by the tasks: kubeconfig, talosconfig, cluster state, manifests
```

| Document | Content |
|---|---|
| [../docs/architecture.md](../docs/architecture.md) | Diagram of both labs, their networks, ports and flows |
| [../docs/access.md](../docs/access.md) | Reach each service: DNS, etcd, the CA, the cluster |
| [../docs/troubleshooting.md](../docs/troubleshooting.md) | Where to look, known behaviour, why it is built this way, what is tested |
| [../../docs/updates.md](../../docs/updates.md) | How the pinned versions are kept current |
