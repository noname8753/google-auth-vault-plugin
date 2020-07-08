#!/bin/bash
echo "Building Plugin"

docker build ../ -f ../Dockerfile.build --tag vault-plugin-auth-google:build
docker container create --name plugin-extract-build vault-plugin-auth-google:build
mkdir -p ../bin/
docker container cp plugin-extract-build:/go/src/github.com/noname8753/vault-plugin-auth-google/cmd/vault-plugin-auth-google/vault-plugin-auth-google ../bin/vault-plugin-auth-google
docker container rm -f plugin-extract-build

echo "Plugin build complete"
