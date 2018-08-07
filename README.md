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
  "username": "laisky",
  "expires_at": "2118-01-01"
}
```

Each request should has the Cookie named `token`:

![demo](https://s3.laisky.com/uploads/2018/06/jwt-demo.jpg)

You can try with readonly token:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmVzX2F0IjoiMjI4Ni0xMS0yMFQxNzo0Njo0MFoiLCJ1c2VybmFtZSI6InJlYWRvbmx5In0.CSb3uVJ-8mXOM0uo1SUkAalQpwmnAU6QvWta6FF3LXo
```

(You can generate HS256 token at [https://jwt.io/](https://jwt.io/))

![generate](https://s3.laisky.com/uploads/2018/06/jwt-generate.jpg)

