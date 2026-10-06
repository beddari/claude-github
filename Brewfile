# Host tools for this repo, on macOS and Linux (Linuxbrew): `task tools`
# or `brew bundle`. Container runtimes are noted below; brew only installs
# their CLIs.

# Build + task runner
brew "go"
brew "go-task"

# Lab tooling (labs/acme-env, labs/talos-cluster)
brew "podman"          # runs acme-env (step-ca, CoreDNS, etcd)
brew "talosctl"        # Talos cluster in Docker
brew "kubernetes-cli"  # kubectl
brew "helm"            # cert-manager, external-dns, Traefik charts
brew "step"            # trust/inspect the lab CA
brew "bind"            # dig
brew "jq"

# The Talos docker provisioner needs a Docker Engine API.
#   Linux: install Docker Engine from your distro (brew's "docker" is only the CLI).
#   macOS: colima provides a Docker VM (`colima start`), or use Docker Desktop/OrbStack.
if OS.mac?
  brew "docker"
  brew "colima"
end
