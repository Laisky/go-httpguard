# GO-HTTPGuard

Prepare:

```sh
git clone git@gitlab.com:Laisky/configs.git /opt/config
```

Run:

```sh
go run entrypoints/main.go --config=/etc/go-httpguard/settings --debug --dry
```

Settings.yml:

```
addr: "0.0.0.0:24215"
audit: "http://xxx/httpguard/logs"
backend: "http://xxx"
secret: "123456"
users:
  - username: "laisky"
    permissions:
      get:
        - "/"
      post:
        - "/"
      head:
        - "/"
```

Each request should has the Cookie named `token`.
