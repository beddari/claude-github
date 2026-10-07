# TPM-rooted machine identity: core concepts illustrated

A visual companion to *tpm-machine-identity-brief.md*. Each diagram shows one idea.

## 1. Who owns what

The provider owns the hardware and its identity. The organisation owns the machine identity, because only the organisation's CA can certify it.

```mermaid
flowchart TB
    INV["Provider inventory"]
    EK["EK = server identity"]
    AK["AK"]
    MK["Machine key = machine identity"]
    VER["Org verifier"]
    OCA["Org root CA"]
    INV -->|records| EK
    EK -->|vouches for| AK
    AK -->|certifies| MK
    OCA -->|delegates to| VER
    VER -->|checks evidence from| AK
    VER -->|issues cert to| MK
    classDef provider fill:#e8eef7,stroke:#4a6fa5,color:#1a1a1a
    classDef tpm fill:#f3f3f3,stroke:#555,color:#1a1a1a
    classDef org fill:#e9f5ec,stroke:#3d8b55,color:#1a1a1a
    class INV provider
    class EK,AK,MK tpm
    class VER,OCA org
```

Blue is the provider, grey is inside the server's TPM, and green is the organisation.

## 2. The key chain inside the TPM

Three keys, three jobs. The EK says "genuine TPM", the AK says "this is what booted", and the machine key is the identity that gets used.

```mermaid
flowchart TB
    EK["Endorsement Key - EK<br/>fixed per chip, vendor cert<br/>survives TPM clear"]
    AK["Attestation Key - AK<br/>signs PCR quotes<br/>destroyed by TPM clear"]
    MK["Machine key - P-256<br/>signs TLS handshakes<br/>destroyed by TPM clear"]
    EK -->|"credential activation proves AK is on this chip"| AK
    AK -->|"TPM2_Certify proves key is TPM-resident"| MK
    classDef tpm fill:#f3f3f3,stroke:#555,color:#1a1a1a
    class EK,AK,MK tpm
```

## 3. Enrollment: the organisation vouches

The provider only relays. Every piece of evidence is signed by the TPM or encrypted to it, so the relay cannot tamper undetected.

```mermaid
%%{init: {"themeVariables": {"actorBkg": "#ffffff", "actorBorder": "#555555", "actorTextColor": "#1a1a1a", "noteBkgColor": "#f3f3f3", "noteBorderColor": "#555555", "actorLineColor": "#555555", "signalColor": "#333333", "signalTextColor": "#1a1a1a", "primaryBorderColor": "#555555", "sequenceNumberColor": "#ffffff"}, "themeCSS": "rect.rect { stroke: #999999 !important; }"}}%%
sequenceDiagram
    autonumber
    box rgb(243,243,243) Machine
        participant A as Agent on machine
    end
    box rgb(232,238,247) Provider
        participant P as Provider relay
    end
    box rgb(233,245,236) Organisation
        participant V as Org verifier
    end
    A->>P: EK cert, AK public, machine key + certify
    P->>V: forward evidence
    V->>V: check EK chain or provider inventory
    V->>P: challenge encrypted to EK, plus nonce
    P->>A: forward challenge
    A->>A: TPM decrypts challenge, AK quotes PCRs over nonce
    A->>P: secret, quote, event log
    P->>V: forward
    V->>V: check secret, quote and golden measurements
    V->>P: X.509 machine cert signed by org CA
    P->>A: deliver cert
```

## 4. Runtime: connecting to NATS

The TLS handshake proves the agent holds the TPM key. The callout then checks that the certificate chains to the right organisation before granting a short-lived NATS identity.

```mermaid
%%{init: {"themeVariables": {"actorBkg": "#ffffff", "actorBorder": "#555555", "actorTextColor": "#1a1a1a", "noteBkgColor": "#f3f3f3", "noteBorderColor": "#555555", "actorLineColor": "#555555", "signalColor": "#333333", "signalTextColor": "#1a1a1a", "primaryBorderColor": "#555555", "sequenceNumberColor": "#ffffff"}, "themeCSS": "rect.rect { stroke: #999999 !important; }"}}%%
sequenceDiagram
    autonumber
    box rgb(243,243,243) Machine
        participant A as Agent
    end
    box rgb(232,238,247) Provider
        participant N as NATS server
        participant C as Auth callout
    end
    A->>N: mTLS handshake, signed by TPM machine key
    N->>N: verify chain against bundle of org roots
    A->>N: CONNECT with NKey nonce signature
    N->>C: auth request with verified cert chain
    C->>C: root equals org anchor, NKey matches cert, machine active
    C->>N: User JWT, about 15 min, in the org account
    N->>A: connected, scoped to own subjects
```

## 5. Server lifecycle: authority moves between anchors

The same mechanism runs under two different anchors. Only the TPM clear at the end of a tenure separates one organisation's machine from the next.

```mermaid
stateDiagram-v2
    [*] --> Racked: EK recorded
    Racked --> ProviderHeld: provider certifies
    ProviderHeld --> OrgHeld: handover
    OrgHeld --> OrgHeld: anchor rotated
    OrgHeld --> Cleaning: machine deleted
    Cleaning --> ProviderHeld: TPM cleared
    ProviderHeld --> Retired: retired
    Retired --> [*]
    classDef provider fill:#e8eef7,stroke:#4a6fa5,color:#1a1a1a
    classDef org fill:#e9f5ec,stroke:#3d8b55,color:#1a1a1a
    classDef neutral fill:#f3f3f3,stroke:#555,color:#1a1a1a
    class Racked,ProviderHeld,Cleaning provider
    class OrgHeld org
    class Retired neutral
```

Blue states are when the provider holds authority over the server, and green is when the organisation does.

- **EK recorded:** the server's EK goes into the provider inventory at racking.
- **Handover:** the org verifier attests the server and certifies a new machine key.
- **Anchor rotated:** the org re-issues the cert over the same TPM key.
- **TPM cleared:** the AK and machine key are destroyed; the EK is kept.

## 6. Physical vs virtual TPM

The protocol is the same. Only the answer to "who vouches that the TPM is genuine" changes.

```mermaid
flowchart TB
    VCA["TPM vendor CA"]
    HEK["Hardware EK"]
    HMK["Host machine key"]
    GEK["vTPM EK"]
    GMK["Guest machine key"]
    ORG["Org root CA"]
    VCA -->|certifies| HEK
    HEK -->|attests| HMK
    HMK -->|"signs, via Incus platform_cert"| GEK
    GEK -->|attests| GMK
    ORG -->|certifies| HMK
    ORG -->|certifies| GMK
    classDef phys fill:#e8eef7,stroke:#4a6fa5,color:#1a1a1a
    classDef virt fill:#fdf1e3,stroke:#c07a2c,color:#1a1a1a
    classDef org fill:#e9f5ec,stroke:#3d8b55,color:#1a1a1a
    class VCA,HEK,HMK phys
    class GEK,GMK virt
    class ORG org
```

Blue is the physical server, orange is the VM, and green is the organisation.

**Trust summary:**

- **Physical TPM:** you trust the TPM vendor and the hardware.
- **vTPM on an org-held server:** the host is already org-certified, so no new trusted party is added.
- **vTPM on a provider-held host:** the provider host is trusted, unless the guest runs as a confidential VM (SEV-SNP or TDX).

## 7. What each layer protects

| Layer | Protects against | Does not protect against |
|---|---|---|
| EK and vendor cert | Fake or emulated TPMs | Provider choosing which genuine server to use |
| AK quote | Tampered firmware, bootloader or OS | Runtime compromise after a good boot |
| Org-issued machine cert | Provider forging a machine identity | Physical attack with BMC access |
| TPM-sealed disk | Disk cloning, offline extraction | Root access on a running, correctly booted node |
| Short-lived NATS JWT | Long-lived credential theft | Abuse within the 15-minute window |
