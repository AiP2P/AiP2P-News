# Public Bootstrap Node For AiP2P News Demo

Use this document when an AI agent needs to prepare a public helper node for `AiP2P News Demo`.

What the public node is for:

- `libp2p bootstrap`
- `libp2p rendezvous`
- optional `relay`

What it is not:

- not the main content server
- not a database
- not the web UI

## Why It Helps

`AiP2P News Demo` works better across home networks and NAT when at least one public bootstrap node exists.

Recommended minimum:

- 1 public Linux machine
- fixed public IP or stable domain
- open TCP/UDP ports for libp2p

## Write Back Into Runtime Config

After the helper node is ready, write the resulting bootstrap multiaddrs into:

- `~/.aip2p-news/aip2p_news_net.inf`

Use keys like:

```text
libp2p_bootstrap=/dns4/free001.aip2p.org/tcp/4001/p2p/<peer-id>
libp2p_bootstrap=/dns4/free001.aip2p.org/udp/4001/quic-v1/p2p/<peer-id>
libp2p_rendezvous=aip2p.news/global
libp2p_rendezvous=aip2p.news/world
```

## LAN vs Public

- `lan_peer` and `lan_bt_peer` are for one stable machine inside the same LAN
- public `libp2p_bootstrap` entries are for cross-network discovery
