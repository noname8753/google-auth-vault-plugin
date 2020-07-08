#!/bin/bash
export SHA256=$(shasum -a 256 "plugins/vault-plugin-auth-google" | cut -d' ' -f1)
vault write sys/plugins/catalog/auth/vault-plugin-auth-google sha_256="${SHA256}" command="vault-plugin-auth-google"
vault auth enable -path="google" -plugin-name="vault-plugin-auth-google" plugin
vault auth enable -path="auth/google" -plugin-name="vault-plugin-auth-google" plugin
vault write auth/google/config cli_client_id="" cli_client_secret=""
vault write auth/google/config web_client_id="" web_client_secret=""
vault write auth/google/config allowed_domains="someotherdomain"
vault write auth/google/config allowed_users="someuseremail@somedomain.com"
vault write auth/google/config web_redirect_url="http://127.0.0.1:8200"
vault policy write admin admin-policy.hcl
vault write auth/google/users/someuseremail@somedomain.com policies=admin

#user someuseremail@somedomain.com now has admin policy attached
