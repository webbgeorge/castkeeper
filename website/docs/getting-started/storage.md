---
sidebar_position: 5
---

# Object storage support

CastKeeper stores audio files and images for podcasts. It can store these
files locally, or using a cloud storage provider.

## Local file storage

The local file storage driver simply stores all podcasts in a given directory
on the file system where CastKeeper is running.

```YAML
ObjectStorage:
  Driver: local
  LocalBasePath: /path/to/storage/dir
```

## Amazon S3

The Amazon S3 driver uses the AWS API to store all podcasts on Amazon S3. Note
that only podcast audio files and images associated with the podcast are stored
in S3 - the database (e.g. if using SQLite) is still stored on the file system.

```YAML
ObjectStorage:
  Driver: awss3
  S3Bucket: my-castkeeper-bucket
  S3Prefix: my-prefix/ # optional
```

AWS config such as credentials, profile, region, etc, should be configured
using AWS environment variables, e.g. `AWS_PROFILE` or `AWS_ACCESS_KEY_ID`. See
[https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html)
for more information.

NOTE that the Amazon S3 driver currently does not support streaming when
downloading podcast files from CastKeeper, which may have a small performance
impact when listening to podcasts. This will be resolved in a future version
of CastKeeper.

## S3-compatible storage

Other S3-compatible storage services can be used instead of Amazon S3, e.g.
Backblaze B2, DigitalOcean Spaces or MinIO. This uses the same driver, `awss3`,
however requires some different configuration: e.g. using the AWS environment
variables such as `AWS_ENDPOINT_URL` as per the services requirements.

Note that an additional config key is provided for compatibility with some
storage providers:

```YAML
ObjectStorage:
  Driver: awss3
  S3Bucket: eg-digital-ocean-space-name
  S3Prefix: prefix/ # optional
  S3ForcePathStyle: true # may be required by some providers
```

## Backups

It it recommended that CastKeeper object data is backed up frequently.
