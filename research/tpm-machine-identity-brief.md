# TPM-rooted machine identity with org-owned trust

Technical brief: building blocks, lifecycle and references. All claims are grounded in the references at the end.

## Goal

- The **provider** owns the *server*: the physical hardware, identified by its TPM Endorsement Key (EK).
- The **organisation** owns the *machine*: its use of that server.
- The machine's identity is a TPM-resident key certified by the **organisation's own CA**, after the organisation's verifier has checked TPM attestation evidence.
- The provider never holds org private keys, so it cannot forge a machine identity.
- Tenant messaging access is derived from that identity through NATS auth callout.

## Building blocks

### B1. Endorsement Key (EK) = server identity

- A manufacturer-provisioned key per TPM, optionally with a manufacturer-signed certificate. For privacy it is not used to sign; it attests that an AK lives on the same TPM. [1]
- EK certificates live at the standard NV indices from the TCG EK Credential Profile: `0x01c00002` (RSA-2048) and `0x01c0000a` (ECC P-256). [2]
- **Rule:** prefer servers whose TPMs ship EK certificates. Otherwise the provider records the EK public key at racking, in an inventory signed by the provider CA. A verifier may trust an EK by chaining its cert to a known manufacturer and/or by querying an inventory system. [1]
- The EK survives tenure changes. `TPM2_Clear` regenerates only the storage primary seed and flushes owner and endorsement objects. [3] The endorsement primary seed changes only through the separate `ChangeEPS` command, [4] so the EK re-derives identically from its template.

### B2. Attestation Key (AK)

- A restricted signing key created inside the TPM. `AttestationKeyECC` in go-tpm-tools creates it in the Owner hierarchy, [2] so `TPM2_Clear` destroys it.
- **Credential activation:** the verifier encrypts a challenge to the EK, and the client proves possession of the AK by decrypting it. The verifier then trusts the AK and may issue it a certificate. [1]
- **Quotes:** the TPM hashes the selected PCRs together with a nonce and signs them with the AK. Replaying the event log against the quoted PCRs verifies the boot chain. [2]

### B3. Machine key = machine identity

- A non-exportable P-256 signing key in the TPM, with attributes `FlagFixedTPM | FlagFixedParent | FlagSensitiveDataOrigin`, certified by the AK via `TPM2_Certify`.
- go-tpm-tools exposes it as a `crypto.Signer` through `Key.GetSigner()`. [2] Go's `tls.Certificate.PrivateKey` accepts any `crypto.Signer`, so the key is directly the mTLS client key.

### B4. Org trust anchor = a private X.509 CA per organisation

- A self-signed P-256 root, held offline by the organisation (HSM or YubiKey), with optional online intermediates for its verifier.
- The provider stores only the public root, registered at onboarding.
- X.509 is chosen for two reasons: NATS mTLS consumes it natively (B6), and name constraints and validity periods come for free.

### B5. Org verifier

A small Go service operated by the organisation. The provider relays the evidence but cannot alter it undetected.

It checks, in order:

1. the EK cert chain against vendor roots, or the EK against the provider's signed inventory;
2. credential activation against that EK;
3. the quote and event log against signed golden measurements of the OS image;
4. the `TPM2_Certify` of the machine key by the AK.

If all four pass, it issues a short-lived X.509 client certificate to the machine-key public key:

- SAN URI `urn:org:<org>:machine:<id>`;
- an extension carrying the EK hash and the agent's NKey public key;
- EKU `clientAuth`.

Renewal repeats the four checks (re-attestation).

### B6. Runtime use with NATS

- **Server TLS:** the NATS server runs with `verify: true`. This is required, because client certificates reach the auth callout only when the server verifies them itself. [5]
- **CA bundle:** the server's `ca_file` is the bundle of all registered org roots (public data). Onboarding an org appends its root and reloads the server config, with no restart. [6]
- **Callout checks:** the callout (callout.go) reads `client_tls.verified_chains`, where the client's peer cert is `VerifiedChains[0]`. [7] It enforces three things:
  - the chain's root equals the anchor registered for the org owning the target account;
  - the cert's NKey extension equals the connecting NKey;
  - the machine is active in the control plane.
- **Result:** the callout mints a ~15-minute User JWT in that org's NATS account. Possession of the TPM key is proven by the TLS handshake itself.
- **Hardening:** run the callout in its own account. Use `xkey` to encrypt callout requests, since requests carry client credentials readable by any subscriber. [8]

### B7. At-rest protection (NKey seed, agent state)

- **Talos ≥1.10:**
  - A `UserVolumeConfig` with a `tpm` LUKS2 key, [9][10] mounted at `/var/mnt/<name>` [11] and bind-mounted into the agent's extension service.
  - The TPM releases the key only against the signed UKI PCR policy plus PCR 7, which covers SecureBoot state and enrolled keys. [12]
- **IncusOS:** native. Disk keys are bound to PCRs 4, 7, 11 and 15. [13]
- **TPM device access from the Talos extension:** bind-mount `/dev/tpmrm0` in the extension's OCI `mounts`. This is the same mechanism the official Tailscale extension uses for `/dev/net/tun`. [14]

### B8. Optional Zitadel machine credential

- B6 makes a stored Zitadel credential unnecessary for node-to-NATS authentication.
- If the agent still needs Zitadel APIs, create a TPM-resident **RSA-2048** key and register its public key with `POST /v2/users/{id}/keys`. Zitadel accepts externally generated keys but documents them as RSA only. [15][16]

## Lifecycle

| Event | Effect |
|---|---|
| Handover | B5 runs under the org anchor. Before handover and after deletion, the same mechanism runs under the provider's anchor. |
| Org anchor rotation | The org re-issues certs over the same TPM machine keys and swaps its root in the CA bundle. Nothing changes on the machine. |
| Tenure end | `TPM2_Clear` destroys the AK and machine key. The EK, and so the server identity, persists. |

## Physical vs software TPM (VMs)

The protocol B2–B6 is identical with a vTPM. Only B1's root of trust changes.

| | Physical TPM | swtpm vTPM |
|---|---|---|
| EK cert issuer | TPM vendor CA | Whoever provisions the vTPM: `swtpm_setup --create-ek-cert` simulates TPM manufacturing with a local CA (`swtpm-localca`) [17] |
| Platform support | Server hardware | Incus `tpm` devices use swtpm (TPM 2.0) [18]; Incus 7.1 adds `instances.tpm.platform_cert` / `platform_key` so Incus signs each vTPM's EK [19] |
| Trusted party | Vendor plus hardware | The host: vTPM state is trivial to inspect and change on the host [20] |
| Protects against | Disk cloning, hardware swap, tampered boot | Disk cloning and tampered guest boot; **not** a malicious host |

**Generalisation:**

- Set the Incus `platform_cert` to a CA whose key is held by the org-certified host machine (B3). The chain then becomes org root → host machine cert → guest EK → guest machine cert.
- For org-held servers, this adds no new trusted party.
- For VMs on provider-held hosts, the provider is in the TCB unless the guest runs as SEV-SNP or TDX. go-tpm-tools already collects those attestation reports alongside the TPM quote. [2]
- Reject IncusOS hosts that ever booted with swtpm: IncusOS permanently records this and reports it in `system_state_is_trusted`. [20]

## Libraries

| Library | Use |
|---|---|
| `github.com/google/go-attestation` | EK/AK, credential activation, quotes, event logs [1] |
| `github.com/google/go-tpm-tools` | `client`: keys, `GetSigner`, `Quote`, `Attest`; `server`: `VerifyAttestation` [2] |
| `github.com/google/go-tpm` | Raw TPM 2.0 commands |
| `github.com/nats-io/jwt/v2` | `AuthorizationRequest.TLS` [7] |
| `github.com/synadia-io/callout.go` | Auth callout service [8] |
| swtpm | `swtpm_setup`, `swtpm-localca` [17] |

**Specifications:**

- TCG EK Credential Profile.
- TPM 2.0 Library Specification.
- RFC 9334 (RATS architecture): attester = agent, verifier = org verifier, relying party = callout.

## References

1. go-attestation README: https://github.com/google/go-attestation
2. go-tpm-tools `client` package: https://pkg.go.dev/github.com/google/go-tpm-tools/client
3. TPM 2.0 reference implementation, `TPM2_Clear`: https://android.googlesource.com/platform/external/tpm2/+/716a46a/Clear.c
4. tpm2-tools (`tpm2_changeeps`, `tpm2_clear`): https://mankier.com/package/tpm2-tools
5. nats-server discussion #5119, client certs to callout require `verify`: https://github.com/nats-io/nats-server/discussions/5119
6. NATS security, config reload without restart: https://docs.nats.io/learn/security/where-next
7. nats-io/jwt `authorization_claims.go` (`ClientTLS`): https://github.com/nats-io/jwt/blob/main/v2/authorization_claims.go
8. NATS auth callout: https://docs.nats.io/learn/security/auth-callout
9. Talos 1.10 what's new (UserVolumeConfig): https://docs.siderolabs.com/talos/v1.10/getting-started/what's-new-in-talos
10. Talos disk encryption: https://docs.siderolabs.com/talos/v1.11/configure-your-talos-cluster/storage-and-disk-management/disk-encryption
11. Talos UserVolumeConfig reference: https://www.talos.dev/v1.10/reference/configuration/block/uservolumeconfig
12. Talos SecureBoot and TPM disk encryption: https://docs.siderolabs.com/talos/v1.11/platform-specific-installations/bare-metal-platforms/secureboot
13. IncusOS security: https://linuxcontainers.org/incus-os/docs/main/reference/security/
14. siderolabs Tailscale extension service: https://github.com/siderolabs/extensions/blob/main/network/tailscale/tailscale.yaml
15. Zitadel AddKey: https://zitadel.com/docs/reference/api/user/zitadel.user.v2.UserService.AddKey
16. Zitadel private key JWT: https://zitadel.com/docs/guides/integrate/service-accounts/private-key-jwt
17. swtpm certificates: https://github.com/stefanberger/swtpm/wiki/Certificates-created-by-swtpm_setup
18. Incus `tpm` device: https://linuxcontainers.org/incus/docs/main/reference/devices_tpm/
19. Incus TPM platform cert (Incus 7.1): https://discuss.linuxcontainers.org/t/issues-with-tpm-platform-cert-setup/26806
20. IncusOS without a TPM: https://linuxcontainers.org/incus-os/docs/main/reference/installing-without-tpm/
