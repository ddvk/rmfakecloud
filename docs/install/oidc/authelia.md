# OIDC with Authelia

This guide shows how to configure [Authelia](https://www.authelia.com/) as the identity provider for rmfakecloud.

See the [OIDC configuration reference](../configuration.md#oidc-authentication) for the full list of available environment variables.

## Authelia client configuration

Add a client entry to the `identity_providers.oidc.clients` section of your Authelia configuration. Generate a hashed secret with:

```bash
authelia crypto hash generate argon2 --random --random.length 64 --random.charset alphanumeric
```

This prints a plaintext secret and its hash. Use the hash in Authelia's config and the plaintext value in `OIDC_CLIENT_SECRET`.

Authelia 4.38+ does not include the `groups` claim in the ID token by default, even when the `groups` scope is requested. You must define a `claims_policy` that explicitly lists `groups` in `id_token`, and reference it from the client. Add the following to the `identity_providers.oidc` section of your Authelia configuration (not inside `clients:`):

```yaml
identity_providers:
  oidc:
    claims_policies:
      with_groups:
        id_token:
          - email
          - email_verified
          - groups
          - preferred_username
          - name
```

Then add the client entry to `identity_providers.oidc.clients`:

```yaml
identity_providers:
  oidc:
    clients:
      - client_id: 'rmfakecloud'
        client_name: 'rmfakecloud'
        client_secret: '$argon2id$v=19$...'  # hashed secret from above
        public: false
        authorization_policy: 'one_factor'  # or 'two_factor' for stricter security
        consent_mode: implicit
        claims_policy: 'with_groups'
        redirect_uris:
          - 'https://your-domain.com/ui/api/oidc/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

The `claims_policy: 'with_groups'` is what causes the `groups` claim to appear in the ID token that rmfakecloud reads. Without it, the groups claim is only available at the userinfo endpoint and `OIDC_ADMIN_CLAIM=groups` will not work.

## Admin group

Create a group named `rmfakecloud-admins` in your Authelia user database and add the users who should have admin access.

## rmfakecloud environment variables

```env
OIDC_PROVIDER_URL=https://auth.example.com
OIDC_CLIENT_ID=rmfakecloud
OIDC_CLIENT_SECRET=<plaintext secret from above>
OIDC_REDIRECT_URL=https://your-domain.com/ui/api/oidc/callback
RM_HTTPS_COOKIE=true  # required: OIDC flow cookies carry the Secure flag
OIDC_ADMIN_CLAIM=groups
OIDC_ADMIN_CLAIM_VALUE=rmfakecloud-admins
OIDC_EXTRA_SCOPES=groups
OIDC_DISPLAY_NAME=Login with Authelia
```

Replace `auth.example.com` with your Authelia hostname and `your-domain.com` with the hostname of your rmfakecloud instance.
