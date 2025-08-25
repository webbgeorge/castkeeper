---
sidebar_position: 6
---

# User Management

When setting up CastKeeper, you'll first need to create an user so you can log
in. This initial user needs to be created using the CastKeeper CLI. It is
recommended that an **admin** user is created, so any future user management
can be done via the CastKeeper web UI - see [Managing Users](/usage/managing-users)
for more information.

## Creating your first user

The initial user needs to be created using the CastKeeper CLI.

```shell
castkeeper users create --username bob --access-level 3
```

This command needs to be run in the same environment where `castkeeper serve`
is run - i.e. the same system which is running the CastKeeper server.

It takes a few flags:

- `--config <path_to_config>`. OPTIONAL. The same config file used for the
  CastKeeper server. By default uses the standard config file paths
  (see [Configuration](/getting-started/configuration)).
- `--username <username>`. REQUIRED. The username to create.
- `--access-level <level>`. REQUIRED. The access level to assign the
  user - 3 is admin.
- `--password <password>`. OPTIONAL. The user's password. If not provided
  it will provide an interactive prompt so the password can be entered securely.

E.g. with specified config file:

```shell
castkeeper users create --config /my/config.yml --username user1 --access-level 3
# As --password is not provided, the command will prompt the user for it 
# securely.
```

### Creating a new user inside a Docker container

To create a user for an instance of CastKeeper running inside a Docker
container, the above command needs to be run inside of that container. This can
be done in several ways, depending on your Docker setup.

```shell
# Docker (container is called `mycontainer`, cli is located at `/castkeeper`)
docker exec -d mycontainer /castkeeper users create --username myuser --access-level 3

# Docker compose (service is called `castkeeper`, cli is located at `/castkeeper`)
docker compose exec castkeeper /castkeeper users create --username myuser --access-level 3

# Kubernetes (pod is called `mypod`, cli is located at `/castkeeper`)
kubectl exec mypod -- /castkeeper users create --username myuser --access-level 3
```

## Modifying, deleting and adding additional users

Users can be managed using either the web UI, or using the CastKeeper CLI. This
is documented on the [Managing Users](/usage/managing-users) page.
