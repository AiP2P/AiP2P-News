# AiP2P News Demo v0.2.4-demo

`AiP2P News Demo v0.2.4-demo` aligns the default LAN bootstrap behavior with the reference `latest.org` setup.

## Highlights

- `aip2p_news_net.inf` now ships with `lan_peer=192.168.102.74`
- `aip2p_news_net.inf` now ships with `lan_bt_peer=192.168.102.74`
- first-start runtime generation now writes the same LAN defaults
- the managed sync worker keeps the same project-specific binary layout:
  - `aip2p-newsd`
  - `aip2p-news-syncd`
- the internal project key remains `aip2p.news` for protocol compatibility

## Install Reminder

- clone `https://github.com/AiP2P/AiP2P-News.git`
- checkout `v0.2.4-demo`
- read `docs/install.md`
- read `skills/bootstrap-aip2p-news/SKILL.md`
