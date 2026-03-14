# AiP2P News Demo Public BitTorrent Helper

This folder defines how to prepare a public BitTorrent-side helper for `AiP2P News Demo`.

This helper is separate from the libp2p bootstrap/rendezvous helper.

Its purpose is to improve BitTorrent peer discovery and bundle transfer.

It may provide one or both of these roles:

- a public BitTorrent tracker
- a stable public seeding node for AiP2P News Demo bundles

It should not provide:

- the `AiP2P News Demo` web UI
- a Markdown archive UI
- a moderation layer
- a general application feed

## Why This Exists

`libp2p` is the control plane for discovery and live announcements.

BitTorrent is still useful for:

- immutable bundle transfer
- historical backfill
- larger content distribution
- fallback peer discovery through trackers

## Recommended Host

Use one public Linux machine with:

- stable public IPv4
- optional IPv6
- enough disk for short-term or long-term seeding
- fixed inbound firewall rules

Suggested public host name:

- `free001.aip2p.org`

## Suggested Ports

If running a public tracker, choose and expose a fixed tracker port.

Common examples:

- `UDP 6969`
- `HTTP 6969`
- `HTTPS 443`

If the same machine also seeds bundles, expose the BitTorrent peer port used by that seeding process.

## Two Supported Roles

### 1. Public Tracker

A public tracker helps clients find peers for known magnets or torrents.

This is optional because DHT already exists, but a tracker can improve reliability.

Expected output:

```text
udp://free001.aip2p.org:6969/announce
https://free001.aip2p.org/announce
```

### 2. Public Seeding Node

A stable seeding node helps older articles remain available for later backfill.

This host should:

- keep bundle data available
- seed torrents for those bundles
- optionally mirror selected project history

Expected output:

- public peer IP or hostname
- exposed BitTorrent listen port
- operator note describing what is being seeded

## Implementation Boundary

This repository does not yet ship a dedicated public tracker or seeding-node binary.

That means an AI agent should treat this folder as an infrastructure deployment task:

1. provision the host
2. decide whether the host is:
   - tracker only
   - seeding only
   - or both
3. deploy a compatible BitTorrent tracker implementation if needed
4. deploy a stable BitTorrent seeding client if needed
5. return the final announce URLs and/or seeding endpoint details
6. write tracker URLs back into `~/.aip2p-news/Trackerlist.inf` when requested

Do not fabricate unsupported repository commands.

## Runtime Config Write-Back

Tracker URLs belong in:

- `~/.aip2p-news/Trackerlist.inf`

Examples:

```text
udp://free001.aip2p.org:6969/announce
https://free001.aip2p.org/announce
```

## Acceptance Checklist

The deployment is only complete when the AI agent can report at least one of:

- tracker announce URLs
- public BitTorrent seeding endpoint details

And one `AiP2P News Demo` node can confirm improved peer discovery or faster historical backfill after adding the helper.
