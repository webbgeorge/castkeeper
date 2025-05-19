---
sidebar_position: 4
---

# Database support

CastKeeper uses a database to store data, such as podcast metadata and users.
It supports both SQLite and PostgreSQL.

## SQLite

SQLite is the recommended choice for most users running a single node or
local deployment of CastKeeper.

```YAML
Database:
  Driver: sqlite
  DSN: /app/data/castkeeper.db
```

## PostgreSQL

PostgreSQL is recommended for users who are familiar with managing a
database server, and who wish to run CastKeeper as a multi-node cluster,
which isn't supported with SQLite.

```YAML
Database:
  Driver: postgres
  DSN: host=localhost user=localdev password=localdev dbname=castkeeper port=5432 sslmode=disable
```

## Database migrations

CastKeeper automatically updates the database schema when the application
starts. When updating CastKeeper, it is recommended that database backups
are taken first.

## Database backups

It is highly recommended that regular backups are taken of the CastKeeper
database.
