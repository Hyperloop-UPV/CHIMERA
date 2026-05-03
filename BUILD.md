# Building CHIMERA

CHIMERA ships as **two separate binaries**, both built from the same shared
`pkg/` tree:

| Binary | Path | Target | Privileges |
|--------|------|--------|------------|
| `chimera-raspi` | `cmd/chimera-raspi` | Raspberry Pi (production). Configures the host network interface, creates dummy interfaces for every board, supports `daemon` / `tui` / `remote` modes. | Requires `sudo` (touches kernel network state). |
| `chimera-sender` | `cmd/chimera-sender` | Desktop test rig. Local TUI only. Skips host-IP setup. When board IPs are loopback (`127.x.x.x`) it skips dummy interfaces too — `lo` already serves that range. | Runs as a normal user when board IPs are loopback and TCP ports are ≥ 1024. |

There is no project-wide build tool (no Makefile, no Taskfile). Everything is
plain `go` commands.

## Prerequisites

- Go 1.24 or newer (`go version`).
- Linux (uses `ip`, dummy interfaces, `/proc/sys` knobs).
- For `chimera-raspi`: `sudo` and the `iproute2` package.

## Build

```sh
# Build chimera-raspi
go build -o chimera-raspi ./cmd/chimera-raspi

# Build chimera-sender
go build -o chimera-sender ./cmd/chimera-sender
```

If you omit `-o`, the resulting binary name matches the `cmd/` directory name
(`chimera-raspi`, `chimera-sender`) and is placed in the current working
directory.

To cross-compile from a desktop for the Raspberry Pi (ARM64):

```sh
GOOS=linux GOARCH=arm64 go build -o chimera-raspi ./cmd/chimera-raspi
```

## Run without building

For quick iteration you can skip the explicit build step:

```sh
go run ./cmd/chimera-sender -config config.json
sudo go run ./cmd/chimera-raspi -mode tui -config config.json
```

## Configuration

Both binaries read the same `config.json` (path overridable with `-config`).
See `USAGE.md` for the schema and ADJ resolution rules.

## Common flags

Shared by both binaries:

| Flag | Default | Description |
|------|---------|-------------|
| `-config <path>` | `config.json` | Path to the configuration file. |
| `-verbose` | `false` | Log every shell command executed by `pkg/utils` and its output. |

Only on `chimera-raspi`:

| Flag | Default | Description |
|------|---------|-------------|
| `-mode <daemon\|tui\|remote>` | `daemon` | Run mode. The same selection can be passed as a positional argument. |

## Tests

```sh
go test ./pkg/...
```

## Cleaning up

The binaries are produced in the working directory and are git-ignored
(`chimera-raspi`, `chimera-sender`). Remove them by hand:

```sh
rm -f chimera-raspi chimera-sender
```

If a `chimera-raspi` run is killed before cleanup, residual dummy interfaces
can be removed with:

```sh
sudo ./removeDummyInterfaces.sh
```
