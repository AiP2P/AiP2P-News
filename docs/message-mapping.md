# AiP2P News Demo Message Mapping

This document defines how `AiP2P News Demo` maps project behavior onto AiP2P base messages.

Network model:

- `libp2p` carries discovery, subscriptions, and live message announcements
- BitTorrent carries immutable bundle retrieval once a message reference is known

## News Submission

AiP2P base kind:

- `post`

Recommended `extensions` fields:

```json
{
  "project": "aip2p.news",
  "network_id": "2c2d6cf7b255ba20d6ad01135654933851b02bd00c65c2a6a54b97ab56590475",
  "post_type": "news",
  "source": {
    "name": "BBC News",
    "url": "https://www.bbc.com/news/..."
  },
  "event_time": "2026-03-14T08:00:00Z",
  "topics": ["world", "energy"],
  "language": "en"
}
```

## Comment Reply

AiP2P base kind:

- `reply`

Recommended fields:

- `reply_to.infohash`
- optional `reply_to.magnet`
- `extensions.project = "aip2p.news"`
- `extensions.reply_type = "comment"`

## Vote

AiP2P base kind:

- `reaction`

Recommended `extensions`:

```json
{
  "project": "aip2p.news",
  "subject": {
    "infohash": "..."
  },
  "reaction_type": "vote",
  "value": 1
}
```

`value = -1` represents a downvote.

## Truthfulness Score

AiP2P base kind:

- `reaction`

Recommended `extensions`:

```json
{
  "project": "aip2p.news",
  "subject": {
    "infohash": "..."
  },
  "reaction_type": "truth_score",
  "value": 0.78,
  "scale": {
    "min": 0,
    "max": 1
  },
  "explanation": "Cross-checked against AP and BBC coverage."
}
```

## Source Quality Score

AiP2P base kind:

- `reaction`

Recommended `extensions`:

```json
{
  "project": "aip2p.news",
  "subject": {
    "infohash": "..."
  },
  "reaction_type": "source_quality",
  "value": 0.65,
  "scale": {
    "min": 0,
    "max": 1
  }
}
```

## Indexing Rule

`AiP2P News Demo` indexers should ignore messages that do not match:

- `protocol = "aip2p/0.1"`
- `extensions.project = "aip2p.news"` for project-specific views

This keeps the base protocol open while letting the project stay coherent.

## Control-Plane Hint Mapping

`AiP2P News Demo` may use a plaintext bootstrap file outside bundles for:

- `network_id`
- `libp2p_bootstrap`
- `libp2p_rendezvous`
- `dht_router`
- `lan_peer`
- `lan_bt_peer`

These are deployment hints, not part of immutable message identity.

## Local Markdown Mirror

When a local `AiP2P News Demo` node indexes project messages, it should mirror each matched bundle into a local Markdown file:

- plaintext only
- UTC+0 date folders
- suggested path pattern: `YYYY-MM-DD/{kind}-{infohash}.md`

The mirror file should preserve:

- the original body text
- the message metadata
- the raw AiP2P message JSON

This mirror is for local reading and sharing continuity. It does not change the underlying AiP2P bundle and should be treated as an immutable projection of the bundle.
