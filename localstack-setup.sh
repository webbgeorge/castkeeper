#!/bin/bash

# this script runs in localstack container on startup to setup S3 for testing

export AWS_ACCESS_KEY_ID=000000000000 AWS_SECRET_ACCESS_KEY=000000000000

awslocal s3 mb s3://castkeeper
