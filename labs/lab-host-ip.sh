#!/bin/sh
# Print the host IP the labs advertise: $HOST_IP from lab.env if set,
# otherwise the source address of the default route.
. "$(dirname "$0")/lab.env"
if [ -n "$HOST_IP" ]; then echo "$HOST_IP"; exit 0; fi
if command -v ip >/dev/null 2>&1; then
  ip -4 route get 1.1.1.1 2>/dev/null | awk '{for (i = 1; i < NF; i++) if ($i == "src") { print $(i + 1); exit }}'
elif hostname -I >/dev/null 2>&1; then
  hostname -I | awk '{print $1}'
else
  # macOS: podman runs in a VM; set HOST_IP in lab.env to an address the VM can reach.
  ipconfig getifaddr en0
fi
