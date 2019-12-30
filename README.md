# Go-HTTPGuard

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
  - username: "readonly"
    password: "readonly"
    permissions:
      get:
        - "/"
      head:
        - "/"
  - username: "laisky"
    permissions:
      get:
        - "/"
      head:
        - "/"
      post:
        - "/"
```

token:

```js
{
  "exp": 4701978061,
  "uid": "jiudian-ai"
}
```

Each request should has the Cookie named `token`:

![demo](https://s3.laisky.com/uploads/2018/06/jwt-demo.jpg)

You can try with readonly token:

```
eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ3MDE5NzgwNjEsInVpZCI6InJlYWRvbmx5In0.PRDnh3tpfxd2G4TDj29HW3QA5dZKq068AYaLSTpA2LQFuAZD-luXvorUfvIXGKY1ESkHVPbkvWEZaX7tTh2E8w
```

(You can generate HS512 token at [https://jwt.io/](https://jwt.io/))

![generate](https://s3.laisky.com/uploads/2019/12/go-httpguard.png)

