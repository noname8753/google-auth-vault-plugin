# HashiCorp Vault plugin for Google Auth.

A HashiCorp Vault plugin for Google Auth.

This fork supports direct mapping of policies to users. If domain or
user to allowed_ it will map the policy=default automatically.
("default" policy is built-in to vault)

directory service/google group usage has been removed

## Setup

The setup guide assumes some familiarity with Vault and Vault's plugin
ecosystem. You must have a Vault server already running, unsealed, and
authenticated.

1. Compile the plugin from source.

2. Move the compiled plugin into Vault's configured `plugin_directory`:

   ```sh
   $ mv vault-plugin-auth-google /etc/vault/plugins/vault-plugin-auth-google
   ```

3. Calculate the SHA256 of the plugin and register it in Vault's plugin catalog.
If you are downloading the pre-compiled binary, it is highly recommended that
you use the published checksums to verify integrity.

   ```sh
   $ export SHA256=$(shasum -a 256 "/etc/vault/plugins/vault-plugin-auth-google" | cut -d' ' -f1)
   $ vault write sys/plugins/catalog/auth/vault-plugin-auth-google \
       sha_256="${SHA256}" \
       command="vault-plugin-auth-google"
   ```

4. Mount the auth method:

   #(Had to enable in both paths at the moment)
   ```sh
   $ vault auth enable \
       -path="google" \
       -plugin-name="vault-plugin-auth-google" plugin
   ```
   ```sh
   $ vault auth enable \
       -path="auth/google" \
       -plugin-name="vault-plugin-auth-google" plugin   
   ``` 
   ```sh
   $ vault read auth/google/config

    Key                              Value
    ---                              -----
    allowed_domains                  []
    allowed_groups                   <nil>
    allowed_users                    []
    cli_client_id                    <>
    cli_client_secret                <redacted>
    cli_max_ttl                      0s
    cli_ttl                          0s
    web_client_id                    <>
    web_client_secret                <redacted>
    web_max_ttl                      0s
    web_redirect_url                 n/a
    web_ttl                          0s
  ```
  ```sh
  vault path-help auth/google

  ## DESCRIPTION

  The Google credential provider allows you to authenticate with Google.

  Documentation can be found at https://github.com/noname8753/vault-plugin-auth-google.

  ## PATHS

  The following paths are supported by this backend. To view help for
  any of the paths below, use the help command with any route matching
  the path pattern. Note that depending on the policy of your auth token,
  you may or may not be able to access certain paths.

    ^cli_code_url$


    ^config$


    ^login$


    ^users/(?P<name>.+)$
        Map username/email to policy.
        vault write auth/google/users/someuser@someemail.com policies=default

    ^users/?$
        This endpoint allows you to create, read, update, and delete configuration

    ^web_code_url$

  ```

5. Create an OAuth client ID in [the Google Cloud Console](https://console.cloud.google.com/apis/credentials), of type "Other".

6. Configure the auth method:

   ```sh
   $ vault write auth/google/config \
       web_client_id=<GOOGLE_CLIENT_ID> \
       web_client_secret=<GOOGLE_CLIENT_SECRET>
   ```

   ```sh
   $ vault write auth/google/config \
       cli_client_id=<GOOGLE_CLIENT_ID> \
       cli_client_secret=<GOOGLE_CLIENT_SECRET>
   ```


7. Allow domains/users -> You do not have to set allowed_domains AND add a user from said domain to allowed_users
  ```sh
  vault write auth/google/config allowed_domains="someotherdomain.com"
  vault write auth/google/config allowed_users="someuser@someemail.com"

  ```

8. Set allowed callback (if using web redirect/compliing the webui with the patches) - Optional
  ```sh
  vault write auth/google/config web_redirect_url="http://127.0.0.1:8200"
  ```

8. Write a policy

  ```sh
  vault policy write admin admin-policy.hcl
  ```

9. Attach said policy to user

  ```sh
  vault write auth/google/users/someuser@ssomeemail.com policies=admin

  or

  vault write auth/google/users/someuser@ssomeemail.com policies=admin,default
  ```



10. Login using Google credentials (NB we use `open` to navigate to the Google Auth URL to get the code).
   (No token is returned if user has no attached policy)

   ```sh
   $ open $(vault read -field=url auth/google/cli_code_url)
   $ vault write auth/google/login code=$GOOGLE_CODE
   ```


## Notes

* If running this inside a docker container or similar, you need to ensure the plugin has the IPC_CAP as well as vault.

  e.g.
  ```sh
  $ sudo setcap cap_ipc_lock=+ep /etc/vault/plugins/google-auth-vault-plugin
  ```

* When building remember your target platform.

  e.g. on MacOS targeting Linux:
  ```sh
  GOOS=linux make
  ```
* You may need to set [api_addr](https://www.vaultproject.io/docs/configuration/index.html#api_addr)

  This can be set at the top level for a standalone setup, or in a ha_storage stanza.

## License

This code is licensed under the MPLv2 license.