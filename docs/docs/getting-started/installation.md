---
sidebar_position: 1
---

# Installation

The CastKeeper server can be installed in multiple ways:

- [Docker image](#docker-image)
- [Build from source](#build-from-source)

## Docker image

The docker image can be found at `ghcr.io/webbgeorge/castkeeper`.

```shell
docker pull ghcr.io/webbgeorge/castkeeper
```

### Docker Compose

Below is a simple Docker Compose setup, using local file storage
in a Docker volume `castkeeper_data`.

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
DataPath: /data

WebServer:
  Port: 8080

ObjectStorage:
  Driver: local
```

```shell
# Start the docker compose example
docker compose up -d

# Create an admin user (required on first run only, password set via interactive prompt)
docker compose exec castkeeper /castkeeper users create --username myuser --access-level 3

# View the logs
docker compose logs castkeeper -f
```

### Helm

Coming soon

## Build From Source

Note that CastKeeper currently can only be built from source on Linux and
MacOS.

1. Follow the developer instructions in the [README](https://github.com/webbgeorge/castkeeper).
2. Place the CastKeeper binary in a location in your system's `$PATH`.
3. Install CastKeeper as a system service, e.g. using `systemctl`.
4. [Configure CastKeeper](/getting-started/configuration).
