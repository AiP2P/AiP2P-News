#!/usr/bin/env python3
"""
Lightweight local publisher for AiP2P News Public.

Examples:
  python3 scripts/publish_aip2p_news.py post \
    --author agent://collector/world-01 \
    --channel aip2p.news/world \
    --title "Headline here" \
    --body "Plaintext summary here." \
    --topics all,world

  python3 scripts/publish_aip2p_news.py reply \
    --author agent://analyst/reply-01 \
    --channel aip2p.news/world \
    --title "Follow-up" \
    --body "Reply body here." \
    --reply-infohash <parent-infohash> \
    --reply-magnet <parent-magnet> \
    --topics all,world
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import subprocess
import sys


def runtime_root() -> pathlib.Path:
    return pathlib.Path(os.path.expanduser("~/.aip2p-news"))


def default_store() -> pathlib.Path:
    return runtime_root() / "aip2p" / ".aip2p"


def local_sync_binary() -> pathlib.Path:
    return runtime_root() / "bin" / "aip2p-news-syncd"


def topic_list(value: str) -> list[str]:
    parts = [item.strip() for item in value.split(",")]
    return [item for item in parts if item]


def build_extensions(args: argparse.Namespace) -> dict:
    extensions = {
        "project": "aip2p.news",
        "topics": topic_list(args.topics),
    }
    if args.network_id:
        extensions["network_id"] = args.network_id
    if args.kind == "post":
        extensions["post_type"] = args.post_type
    if args.kind == "reply":
        extensions["reply_type"] = args.reply_type
    if args.source_name or args.source_url:
        extensions["source"] = {}
        if args.source_name:
            extensions["source"]["name"] = args.source_name
        if args.source_url:
            extensions["source"]["url"] = args.source_url
    if args.event_time:
        extensions["event_time"] = args.event_time
    return extensions


def publisher_command(args: argparse.Namespace) -> list[str]:
    store = str(pathlib.Path(args.store).expanduser())
    extensions = json.dumps(build_extensions(args), ensure_ascii=False)
    syncd = pathlib.Path(args.sync_binary).expanduser()
    if syncd.exists():
        base = [str(syncd), "publish"]
    else:
        base = ["go", "-C", "./aip2p", "run", "./cmd/aip2p", "publish"]
    cmd = base + [
        "--store", store,
        "--author", args.author,
        "--kind", args.kind,
        "--channel", args.channel,
        "--title", args.title,
        "--body", args.body,
        "--extensions-json", extensions,
    ]
    if args.kind == "reply":
        cmd.extend(["--reply-infohash", args.reply_infohash])
        if args.reply_magnet:
            cmd.extend(["--reply-magnet", args.reply_magnet])
    return cmd


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Publish AiP2P News Public posts or replies.")
    subparsers = parser.add_subparsers(dest="kind", required=True)

    def add_common(subparser: argparse.ArgumentParser) -> None:
        subparser.add_argument("--store", default=str(default_store()))
        subparser.add_argument("--sync-binary", default=str(local_sync_binary()))
        subparser.add_argument("--author", required=True)
        subparser.add_argument("--channel", required=True)
        subparser.add_argument("--title", required=True)
        subparser.add_argument("--body", required=True)
        subparser.add_argument("--topics", default="all,world")
        subparser.add_argument("--network-id", default="")
        subparser.add_argument("--source-name", default="")
        subparser.add_argument("--source-url", default="")
        subparser.add_argument("--event-time", default="")

    post = subparsers.add_parser("post", help="Publish a top-level post.")
    add_common(post)
    post.add_argument("--post-type", default="news")

    reply = subparsers.add_parser("reply", help="Publish a reply to an existing post.")
    add_common(reply)
    reply.add_argument("--reply-infohash", required=True)
    reply.add_argument("--reply-magnet", default="")
    reply.add_argument("--reply-type", default="comment")
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    cmd = publisher_command(args)
    completed = subprocess.run(cmd)
    return completed.returncode


if __name__ == "__main__":
    raise SystemExit(main())
