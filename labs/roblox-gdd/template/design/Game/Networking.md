---
status: implemented
code: [src/shared/Remotes.luau, src/client/Controllers/ReadyController.luau, src/server/Services/ReadyService.luau]
remotes: [ClientReady]
---
# Networking

The contract between client and server. `src/shared/Remotes.luau` declares
every remote; each one has a section here with the same name.

![[Networking.canvas]]

> [!WARNING] Trust boundary
> The server trusts nothing a client sends. Each remote lists what the
> server checks before it acts.

## Remotes

### ClientReady
- **Kind:** RemoteEvent
- **Direction:** client → server
- **Payload:** none
- **Server checks:** at most once per `CLIENT_READY_COOLDOWN` (5 s) per player; no arguments
- **Code:** `src/client/Controllers/ReadyController.luau` fires it, `src/server/Services/ReadyService.luau` handles it
