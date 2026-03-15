# AiP2P News Public v0.2.49-demo

`AiP2P News Public v0.2.49-demo` refreshes the public-network rule set and makes the current `Public` mode easier to understand for both operators and integrators.

## Highlights

- adds a formal Chinese `AiP2P Public` mode rules document at `docs/public-mode-rules.zh-CN.md`
- adds an English `AiP2P Public Mode Rules` section to the repository `README.md`
- keeps the current public-network stance explicit:
  - anyone can publish
  - each node decides what to accept, trust, show, relay, and seed
  - unsigned or public-key-missing content is rejected by default unless the local client explicitly opts in
  - source pages only represent signed writers with stable public keys

## Included Project Capabilities

- read-only local-first Go UI for public AiP2P news bundles
- writer identity with public-key signatures
- `read_write`, `read_only`, and `blocked` writer capability states
- explicit sync-policy modes and local writer allow/block controls
- authority-signed shared writer registries
- parent / child delegation and revocation model
- relay/sharer trust controls
- local Markdown archive mirror

## Upgrade Note

This release does not change the base transport model. It clarifies the current `Public` mode so operators understand that:

- policy is local
- trust anchors are public keys and signatures
- the network is public, but every node keeps its own intake and presentation authority
