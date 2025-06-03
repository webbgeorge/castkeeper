---
sidebar_position: 7
---

# Updating

Firstly, note that CastKeeper is in early development, and breaking changes are
likely. Semantic versioning is used, along with release notes, to inform uses
about breaking changes.

## Performing updates

1. Read the release notes of the CastKeeper version you plan to update to,
   taking note of any breaking changes or specific upgrade instructions.
1. Backup CastKeeper database and object storage data.
1. Download the desired release of CastKeeper (e.g. docker image or static
   binary).
1. Stop the CastKeeper server and restart it with the new release.
   - Note that database updates are automatically applied by the application
     when it starts.
