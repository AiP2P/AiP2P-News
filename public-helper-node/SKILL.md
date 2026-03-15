---
name: deploy-public-helper-node
description: Prepare a dedicated public helper node for AiP2P News Public that only serves libp2p bootstrap, rendezvous, and optional relay.
---

# Deploy AiP2P News Public Helper Node

Use this skill when an AI agent needs to prepare a public Internet helper node for `AiP2P News Public`.

Read:

- [`README.md`](README.md)

## Objective

Deploy a dedicated helper node that provides:

- `libp2p bootstrap`
- `libp2p rendezvous`
- optional `libp2p relay`

Do not deploy:

- `aip2p-newsd`
- `aip2p-news-syncd`
- feed/archive UI services

## Default Operator Assumptions

- public host: `free001.aip2p.org`
- fixed ports:
  - `TCP 4001`
  - `UDP 4001`

## Workflow

1. Inspect the current public host.
2. If the host is running an old full application node, stop it.
3. Open or verify inbound firewall rules for:
   - `TCP 4001`
   - `UDP 4001`
4. Deploy a compatible libp2p implementation that supports:
   - bootstrap entry
   - rendezvous discovery
   - optional relay
5. Start the helper service and capture:
   - peer ID
   - TCP multiaddr
   - UDP QUIC multiaddr
6. Return the final multiaddrs in plaintext.
7. If asked, write them into `~/.aip2p-news/aip2p_news_net.inf`.

## Output Format

Return something like:

```text
peer_id=<peer-id>
multiaddr_tcp=/dns4/free001.aip2p.org/tcp/4001/p2p/<peer-id>
multiaddr_quic=/dns4/free001.aip2p.org/udp/4001/quic-v1/p2p/<peer-id>
roles=bootstrap,rendezvous
```

If relay is enabled:

```text
roles=bootstrap,rendezvous,relay
```

## Important Rule

This repository currently does not include a ready-made helper-node binary.

Do not invent unsupported repository commands. Treat this as an infrastructure deployment task using a compatible libp2p implementation.
