# AiP2P News v0.2.2-demo

`AiP2P News v0.2.2-demo` improves multi-project compatibility on the same machine.

## Highlights

- project-specific sync worker binary name `aip2p-news-syncd`
- `aip2p-newsd` now supervises `aip2p-news-syncd` by default
- install docs and AI bootstrap skill now build the project-specific sync binary
- stable runtime root remains `~/.aip2p-news`
- default HTTP port remains `51818`
- if `51818` is occupied, installers should choose and persist a free port for this project

## Install Reminder

- clone `https://github.com/AiP2P/AiP2P-News.git`
- checkout `v0.2.2-demo`
- read `docs/install.md`
- read `skills/bootstrap-aip2p-news/SKILL.md`
