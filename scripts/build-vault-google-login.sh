#!/bin/bash
#change to know if you want other vault builtin auth methods to display in ui
disableOtherAuths="yes"
echo "Building Vault"

docker build ../ -f ../Dockerfile.vault --tag vault-google:build --build-arg disableOtherAuths="$disableOtherAuths"
docker container create --name vault-google-extract-build vault-google:build
mkdir -p ../bin/

docker container cp vault-google-extract-build:/go/src/github.com/hashicorp/vault/bin/vault ../bin/vault-google-auth-ui
docker container rm -f vault-google-extract-build

echo "Done building Vault"

