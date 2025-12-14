---
sidebar_position: 1
---

# Managing Podcasts

## Adding podcasts

Podcasts can be added in 2 ways:

- By search:
  - CastKeeper uses the iTunes API to search for podcasts.
  - Podcasts are added from search results by clicking "Add podcast"
- By directly adding a feed URL:
  - For feeds which aren't available on the iTunes API (e.g. private or
    subscriber-only feeds)
  - Choose the "Add Feed URL" button on the Add Podcast page, and then enter
    the Feed URL and credentials (optional).

When a podcast is added to CastKeeper, all previous episodes will be downloaded
and any new episodes are automatically downloaded as they are released.

## Deleting podcasts

Coming soon.

## Retrying failed downloads

If an episode download fails, CastKeeper automatically retries the download up
to 5 times. If the download continues to fail, it will show as `failed` in the
view podcast page with a "retry" link - which can be used to requeue the
download.
