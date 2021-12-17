The following environment variables are available:

- `JWT_SECRET_KEY` needed for the whole auth thing to work, set something long  
- `STORAGE_URL` controls whether file upload/download goes through the local proxy or directly. the address of rmfakecloud **as visible from the tablet**, especially if the host is behind a reverse proxy or in a container (default: http://hostname:port)  
- `PORT` port number (default: 3000)  
- `DATADIR` to set data/files directory (default: data in current dir)  
- `LOGLEVEL` default to **info** (set to **debug** for more logging or **warn**, **error** for less)
- `RM_HTTPS_COOKIE=1` UI, send auth cookies only via https

For handwriting recognition support you need an application key and hmac from myscript.com:

- `RMAPI_HWR_APPLICATIONKEY` application key obtained from myscript
- `RMAPI_HWR_HMAC` hmac obtained from myscript

For sending emails:

- `RM_SMTP_SERVER` the smtp server address
- `RM_SMTP_USERNAME` username/email for login
- `RM_SMTP_PASSWORD` plaintext password (application password should work)
- `RM_SMTP_FROM` custom `FROM` header for the mails
