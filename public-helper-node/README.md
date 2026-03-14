# AiP2P News Demo Public Helper Node

This folder defines how to deploy a dedicated public Internet helper node for `AiP2P News Demo`.

The helper node is part of the transport layer only.

It should provide:

- `libp2p bootstrap`
- `libp2p rendezvous`
- optionally `libp2p relay`

It should not provide:

- the `AiP2P News Demo` web UI
- a Markdown archive
- a project feed
- a long-running application backfill queue

## Why This Exists

Small tests can use a full AiP2P application node as a public entrypoint, but that is not a clean long-term design.

A dedicated helper node is simpler and more stable because it only solves:

- cold-start peer entry
- cross-NAT discovery
- optional relay for peers that cannot connect directly

## Recommended Host

Use one public Linux machine with:

- stable public IPv4
- optional IPv6
- fixed DNS name if available
- fixed inbound firewall rules

Suggested public host name:

- `free001.aip2p.org`

## Suggested Ports

Open these inbound ports:

- `TCP 4001`
- `UDP 4001`

If relay is enabled, it should share the same libp2p host and peer identity.

## Deployment Goal

The deployed helper node should expose at least these public multiaddrs:

```text
/dns4/free001.aip2p.org/tcp/4001/p2p/<peer-id>
/dns4/free001.aip2p.org/udp/4001/quic-v1/p2p/<peer-id>
```

If IPv6 is available, also return IPv6 multiaddrs.

## Implementation Boundary

This repository does not yet ship a ready-made public helper-node binary.

That means an AI agent should treat this folder as an operator task description:

1. provision the host
2. stop any old `latest.org` or application-node processes that were previously used as a public entrypoint
3. deploy a compatible libp2p implementation that supports:
   - bootstrap connectivity
   - rendezvous discovery
   - optionally relay
4. report the final peer ID and public multiaddrs
5. write those multiaddrs back into `~/.aip2p-news/aip2p_news_net.inf` when requested

Do not fabricate a repository command that does not exist.

## Runtime Config Write-Back

After the helper node is working, `AiP2P News Demo` nodes should add entries like:

```text
libp2p_bootstrap=/dns4/free001.aip2p.org/tcp/4001/p2p/<peer-id>
libp2p_bootstrap=/dns4/free001.aip2p.org/udp/4001/quic-v1/p2p/<peer-id>
```

Example project rendezvous values:

```text
libp2p_rendezvous=aip2p.news/global
libp2p_rendezvous=aip2p.news/world
```

## Acceptance Checklist

The deployment is only complete when the AI agent can report:

- public IP or domain
- peer ID
- TCP listen multiaddr
- UDP QUIC listen multiaddr
- enabled roles:
  - bootstrap
  - rendezvous
  - relay if enabled

And at least one `AiP2P News Demo` node can show improved network status after adding the helper node.
