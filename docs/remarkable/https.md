Now you can run TLS directly on rmfakecloud:
```
TLS_CERT=somecert
TLS_KEY=somecertkey
```

Set both variables. If only one of them is set, rmfakecloud refuses to start
instead of silently serving plain HTTP.
