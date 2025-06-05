---
sidebar_position: 2
---

# Configuration

CastKeeper is configured using a YAML config file (`castkeeper.yml`),
environment variables, or a mixture of both.

## Config files

Config files are loaded by the CastKeeper server when it starts - either
from a path specified when starting the server, or from the default CastKeeper
config path. Only the first config file found is loaded, following this order:

1. The path specified as an argument when running the CastKeeper server, e.g.
   `castkeeper serve --config /path/to/my/castkeeper.yml`
2. `./castkeeper.yml` - in the current working directory where the CastKeeper
   server was started.
3. `/etc/castkeeper/castkeeper.yml`

### Example config file

```YAML
EnvName: prod
LogLevel: warn
BaseURL: https://castkeeper.example.com

WebServer:
  Port: 8080
  CSRFSecretKey: testValueDoNotUseInProd

Database:
  Driver: sqlite
  DSN: /app/data/castkeeper.db

ObjectStorage:
  Driver: awss3
  S3Bucket: my-castkeeper-bucket
  S3Prefix: podcasts/
```

## Environment variables

Every CastKeeper config option can alternatively be specified as an
environment variable.

The names of environment variables follow a convention:

- the config file option `EnvName` would become `CASTKEEPER_ENVNAME`.
- the nested config file option `WebServer.Port` would become `CASTKEEPER_WEBSERVER_PORT`

See the reference table below for the full list of environment variables.

## Config reference

| Config option                  | Environment variable                      | Description |
| ------------------------------ | ----------------------------------------- | ----------- |
| LogLevel                       | CASTKEEPER_LOGLEVEL                       | Log verbosity. Allowed values: `debug`,`info`, `warn`, `error`. Default value: `info`. |
| EnvName                        | CASTKEEPER_ENVNAME                        | The name of the environment, used in logs. Default value: `unknown`. |
| BaseURL                        | CASTKEEPER_BASEURL                        | The URL that CastKeeper is hosted at, e.g. `https://ck.example.com`. Required. |
| WebServer.Port                 | CASTKEEPER_WEBSERVER_PORT                 | The port the web server should listen to. Default value: `8080`. |
| WebServer.CSRFSecretKey        | CASTKEEPER_WEBSERVER_CSRFSECRETKEY        | Unique secret key used to create CSRF tokens. It is very important that this value remains secret. To avoid this key being stored in plain text in the config file, it is recommended that it is provided as an environment variable. Required. |
| WebServer.CSRFSecureCookie     | CASTKEEPER_WEBSERVER_CSRFSECURECOOKIE     | Boolean value - if true, the CSRF cookie has the secure attribute. Should be true in production if CastKeeper is on HTTPS. Default value: `true`. |
| Database.Driver                | CASTKEEPER_DATABASE_DRIVER                | Database driver to use. Allowed values: `sqlite`, `postgres`. Required. |
| Database.DSN                   | CASTKEEPER_DATABASE_DSN                   | Connection string for the database. Uses [GORM DSN format](https://gorm.io/docs/connecting_to_the_database.html). E.g. `filepath.sql` for SQLite, or `host=localhost user=localdev password=localdev dbname=castkeeper port=5432 sslmode=disable` for PostgreSQL. Required. |
| ObjectStorage.Driver           | CASTKEEPER_OBJECTSTORAGE_DRIVER           | The object storage provider to use. Allowed values: `local`, `awss3`. Required. |
| ObjectStorage.LocalBasePath    | CASTKEEPER_OBJECTSTORAGE_LOCALBASEPATH    | The path to the directory to store files in when using the `local` provider. Required when `Driver` is `local`. |
| ObjectStorage.S3Bucket         | CASTKEEPER_OBJECTSTORAGE_S3BUCKET         | The S3 bucket to use for file storage when using the `awss3` provider. Required when `Driver` is `awss3`. |
| ObjectStorage.S3Prefix         | CASTKEEPER_OBJECTSTORAGE_S3PREFIX         | Optional prefix for files when using the `awss3` provider. |
| ObjectStorage.S3ForcePathStyle | CASTKEEPER_OBJECTSTORAGE_S3FORCEPATHSTYLE | Boolean value. Usually false, may need to be set to true for some S3 compatible storage services. Default value: false. |
