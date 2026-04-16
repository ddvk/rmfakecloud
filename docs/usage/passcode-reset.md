# Passcode (PIN) Reset

If you forget the passcode on your tablet, the reMarkable lockscreen offers a
**Forgot PIN** flow that asks the cloud to approve the reset. rmfakecloud
implements the same flow so you can recover a locked device without needing
the official cloud.

!!! note "Device support"
    This flow only exists on **reMarkable 1** and **reMarkable 2**. Later devices use a recovery code or a factory reset.

## How it works

1. After several incorrect passcode attempts, an option that says **swipe
   to unlock** will appear at the bottom of your screen. Swipe this and
   then press **Reset passcode**.
2. Follow the instructions on your device, and go to your rmfakecloud web
   UI on your computer and log in to your account.
3. A banner appears at the top of every page showing the pending reset
   request for your device — click **Approve**.
4. Your old passcode has been reset. Enter a new passcode on your reMarkable.

If the request wasn't initiated by you, click **Dismiss** instead — the
request is removed from rmfakecloud and the tablet will no longer be able to
complete the reset for that attempt.

Pending requests are kept in memory only and expire after 24 hours. If the
rmfakecloud container restarts before you approve, just tap **Forgot PIN**
again on the tablet to create a fresh request.
