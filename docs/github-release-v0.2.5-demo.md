# AiP2P News Demo v0.2.5-demo

`AiP2P News Demo v0.2.5-demo` makes the feed page cleaner and easier to operate.

Highlights:

- removes the duplicate `Network dashboard` panel from the feed homepage
- keeps network state in the left rail and the dedicated `/network` page
- adds feed pagination with a default page size of `20`
- supports `page` and `page_size` query parameters
- lets operators switch between `20`, `50`, and `100` items per page

Install / update:

- checkout `v0.2.5-demo`
- run `go test ./...`
- run `go -C ./aip2p test ./...`
- build `aip2p-newsd` and `aip2p-news-syncd`

Operational note:

- network state still belongs on `/network`
- the homepage feed is now focused on stories, filters, and pagination
