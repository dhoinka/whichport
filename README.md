# whichport

`whichport` is a small CLI that shows which local applications are currently listening on ports.

It prints:

- port
- protocol
- PID
- command
- invocation

## Status

Current implementation supports **macOS** and **Linux**.

## Build

```bash
go build ./cmd/whichport
```

## Run

```bash
go run ./cmd/whichport
```

Example:

```text
PORT   PROTO  PID    COMMAND             INVOCATION
---------------------------------------------------
3000   TCP    63152  node                ng serve --proxy-config proxy.conf.json
5432   TCP    43267  com.docker.backend  com.docker.backend services
```

## Useful flags

```bash
# only one port
go run ./cmd/whichport --port 3000

# only tcp or udp
go run ./cmd/whichport --protocol tcp
go run ./cmd/whichport --protocol udp

# machine-readable output
go run ./cmd/whichport --json

# disable color
go run ./cmd/whichport --no-color
```

## Development

```bash
go test ./...
go test ./internal/ports -run TestParsePSMetadata
go test ./internal/cli -run TestTruncateMiddle
```

## CI

GitHub Actions runs the test suite and builds the CLI on **pushes to `main`** and on **pull requests** for:

- `ubuntu-latest`
- `macos-latest`

## License

MIT. See [LICENSE](./LICENSE).
