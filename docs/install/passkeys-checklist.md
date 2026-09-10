# Passkeys (WebAuthn) — manual checklist

Requires `RMFAKECLOUD_WEBAUTHN=true` and HTTPS origin matching `RMFAKECLOUD_WEBAUTHN_RPID` / `ORIGINS` (or derived from https `STORAGE_URL`).

1. Open the web UI over HTTPS (not `http://acorn:3000` / raw IP).
2. Sign in with email/password.
3. Profile → **Passkeys** → Add passkey (optional label). Complete browser/OS prompt.
4. Sign out. Login page should show **Sign in with passkey**.
5. Use passkey to sign in; land on Documents; cookie/JWT works as usual.
6. Confirm password login still works.
7. Profile → remove passkey; optional: set `RMFAKECLOUD_WEBAUTHN=false` and restart → status/`Passkeys` UI hidden, login button gone.
