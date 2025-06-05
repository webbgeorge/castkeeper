---
sidebar_position: 6
---

# User Management

CastKeeper uses username and password authentication to let users log in.

## Creating users

Currently, there is no GUI implemented to manage users. This is on the roadmap
to be implemented soon - however there is a command line interface to allow
user creation.

```shell
castkeeper user create
```

This command needs to be run in the same environment where `castkeeper serve`
is run - i.e. the same system which is running the CastKeeper server.

It takes a few flags:

- `--config <path_to_config>`. OPTIONAL. The same config file used for the
  CastKeeper server. By default uses the standard config file paths
  (see [Configuration](/getting-started/configuration)).
- `--username <username>`. OPTIONAL. The username to create. If not provided
  it will provide an interactive prompt.
- `--password <password>`. OPTIONAL. The user's password. If not provided
  it will provide an interactive prompt.

E.g. with optional flags:

```shell
castkeeper user create --config /my/config.yml --username user1
# As --password is not provided, the command will prompt the user for it 
# securely.
```

## Modifying and deleting users

No facility is currently provided for modifying or deleting users. This is on
the roadmap for the near future, until then users can be deleted manually from
the database and then recreated.

## User permissions

There is currently no user permission system implemented. This is on the
roadmap for the near future.

## Alternative authentication methods

CastKeeper currently does not support other authentication methods, however
there are plans to support auth proxies (such as CloudFlare Access, TailScale
serve or oauth2-proxy) in the near future.
