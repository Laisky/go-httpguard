# Go-HTTPGuard


Install & Run:

```sh
go install github.com/Laisky/go-httpguard/v2/cmd/gohttpguard@latest

gohttpguard --help
```

Run with source code:

```sh
go run -race ./cmd/gohttpguard -c ./sample/config.yml --debug
```

## Go compatibility

The current security dependency set requires Go 1.26.0 or newer. Use the
patched Go 1.26.9 or Go 1.27.2 compiler when building or testing. This raises
the previous Go 1.18 module floor; downstream consumers using older compilers
must upgrade their toolchain before adopting this version. The floor follows
[golang.org/x/net v0.60.0](https://proxy.golang.org/golang.org/x/net/@v/v0.60.0.mod),
which selects the fixed crypto and text dependency versions. Security metadata
covers the dependency graph and does not establish that every advisory is
reachable through this application.

The Docker build uses patched Go 1.26.9 on Bookworm and explicitly enables
external static linking. The runtime now uses supported Debian Bookworm:
the previous Bullseye security-package URLs returned 404 during a fresh build.
The existing runtime package, user and startup commands are preserved.
