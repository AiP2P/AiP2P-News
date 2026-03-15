---
name: deploy-public-bittorrent-helper
description: Prepare a public BitTorrent-side helper for AiP2P News Public, such as a tracker and/or a stable seeding node for bundle backfill.
---

# Deploy AiP2P News Public BitTorrent Helper

Use this skill when an AI agent needs to prepare a public BitTorrent helper for `AiP2P News Public`.

Read:

- [`README.md`](README.md)

## Objective

Prepare one of these:

- a public tracker
- a stable public seeding node
- or both

Do not deploy:

- `aip2p-newsd`
- `aip2p-news-syncd`
- feed/archive UI services

## Default Operator Assumptions

- public host: `free001.aip2p.org`
- fixed tracker port if enabled:
  - `UDP 6969`
  - optional `HTTP 6969` or `HTTPS 443`

## Workflow

1. Inspect the current public host.
2. Decide which role is needed:
   - tracker
   - seeding node
   - both
3. Open or verify inbound firewall rules for the chosen tracker and/or seeding ports.
4. Deploy a compatible tracker implementation if tracker mode is required.
5. Deploy a stable BitTorrent seeding client if seeding mode is required.
6. Return:
   - announce URLs
   - listen ports
   - any seeding host details
7. If asked, write tracker URLs into `~/.aip2p-news/Trackerlist.inf`.

## Output Format

For tracker mode:

```text
tracker_udp=udp://free001.aip2p.org:6969/announce
tracker_https=https://free001.aip2p.org/announce
```

For seeding mode:

```text
seed_host=free001.aip2p.org
seed_role=public_bundle_seed
seed_port=<port>
```

## Important Rule

This repository currently does not include a ready-made public tracker or seeding-helper binary.

Do not invent unsupported repository commands. Treat this as an infrastructure deployment task.
