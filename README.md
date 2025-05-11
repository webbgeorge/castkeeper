# CastKeeper

CastKeeper is a free application for archiving podcasts. It is designed to be
easy to self-host, using either Docker or a self-contained executable. It
supports a variety of file storage options.

CastKeeper was built primarily to do 2 things I wanted:

1. Store copies of podcasts I listen to, safe from being taken offline or
   being modified (e.g. due to being abandoned or censored), allowing me to
   re-listen to them in future.
2. Transcribe and index the text content of those podcasts, allowing me to
   search for when certain topics were discussed (coming very soon in a
   future release).

## Getting started

- [Getting started](https://castkeeper.org/docs/intro)
  - [Installation](https://castkeeper.org/docs/getting-started/installation)
  - [Configuration](https://castkeeper.org/docs/getting-started/configuration)

## Documentation

The CastKeeper documentation is available at [castkeeper.org](https://castkeeper.org).

## Contributing

Contributions are welcome. If you wish to contribute more than bug-fixes,
please start a discussion in GitHub issues first to ensure the feature is
desired - as I want to make sure your time isn't wasted.

### Developer guide

#### Prerequisites

- Go 1.24
- Node.js 22
- Docker and Docker Compose (or equivalent)

```shell
# Install tools and dependencies
make install
```

#### Running tests

```shell
# Run unit tests + static analysis checks
make test

# Run unit tests only
go test ./... -short -count=1

# Run E2E tests
docker compose up -d
make test_e2e
```

#### Running local development server (default configuration)

To run locally with the default configuration, using SQLite and local object
storage:

```shell
# Create a user to log in with (first time only)
make create_user

# Run server
make run

# or, watch and rebuild on file changes
make watch
```

Visit the web UI at: <http://localhost:8080>

#### Running local development server (alt configuration)

To run locally with the alt configuration, using PostgreSQL and S3 (localstack):

```shell
# Start Docker Compose
docker compose up -d

# Create a user to log in with (first time only)
make create_user_postgres

# Run server
make run_postgres_s3
```

Visit the web UI at: <http://localhost:8081>

## License

CastKeeper
Copyright (C) 2025  George Webb

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
