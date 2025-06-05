---
sidebar_position: 1
---

# Installation

The CastKeeper server can be installed in multiple ways:

- [Docker image](#docker-image)
- [Static binary](#static-binary)

## Docker image

The docker image can be found at `ghcr.io/webbgeorge/castkeeper`.

```shell
docker pull ghcr.io/webbgeorge/castkeeper
```

### Docker Compose

Below is a simple Docker Compose setup, using SQLite and
local file storage in a Docker volume `castkeeper_data`.

See [configuration](/getting-started/configuration#docker-compose) for
how to set up the `castkeeper.yml` file.

Also note if using this approach, you should back up the `castkeeper_data`
volume.

```yml
# docker-compose.yml
services:
  castkeeper:
    image: ghcr.io/webbgeorge/castkeeper:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./castkeeper.yml:/etc/castkeeper/castkeeper.yml
      - castkeeper_data:/data

volumes:
  castkeeper_data:
```

```YAML
# castkeeper.yml
# See the configuration page for more information about configuring CastKeeper
EnvName: prod
LogLevel: warn
BaseURL: http://localhost:8080

WebServer:
  Port: 8080
  CSRFSecretKey: <REPLACE_THIS_VALUE>

Database:
  Driver: sqlite
  DSN: /data/castkeeper.db

ObjectStorage:
  Driver: local
  LocalBasePath: /data/objects
```

### Helm

Coming soon

## Static binary

Note that CastKeeper static binaries currently only run on Linux and MacOS.

1. Obtain the binary from [releases on GitHub](https://github.com/webbgeorge/castkeeper/releases)
2. Place the CastKeeper binary in a location in your system's `$PATH`.
3. Install CastKeeper as a system service, e.g. using `systemctl`.
4. [Configure CastKeeper](/getting-started/configuration).
