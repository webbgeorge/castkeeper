---
sidebar_position: 4
---

# Database

CastKeeper uses a SQLite database to store data, such as podcast metadata and
users. The database is stored under the CastKeeper data directory, as specified
in the config file.

## Config

The data directory is defined in the config file:

```YAML
DataDirPath: /data
```

## Database migrations

CastKeeper automatically updates the database schema when the application
starts. When updating CastKeeper, it is recommended that database backups
are taken first.

## Database backups

It is highly recommended that regular backups are taken of the CastKeeper
database.
