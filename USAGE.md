# CHIMERA — Usage Guide

CHIMERA is the Hyperloop UPV board emulator. It loads the ADJ (vehicle definition), spins up one runtime per board, and exposes a control interface to inspect/modify measurements.

## Build

```sh
go build -o CHIMERA .
```

## Configuration (`config.json`)

| Field                  | Description                                                                                         |
| ---------------------- | --------------------------------------------------------------------------------------------------- |
| `adj-branch`           | ADJ git branch to clone when no local path is given.                                                |
| `adj-path`             | Optional local path to an existing ADJ repository. Empty → clone from GitHub.                       |
| `network.interface`    | Network interface to bind board IPs to.                                                             |
| `network.ip`           | IP/CIDR assigned to the host on that interface.                                                     |
| `network.control-port` | TCP port the daemon listens on for remote TUI clients.                                              |
| `initial-period-ms`    | Default packet emission period (ms) per plate runtime.                                              |

### ADJ resolution order

1. If `adj-path` is set and exists → use it.
2. Else clone `adj-branch` from `https://github.com/Hyperloop-UPV/adj.git`.
3. If `adj-path` is set but missing, **or** the clone fails (e.g. no network) → fall back to `./adj` next to the executable.

The branch and commit hash of the resolved ADJ are printed at startup of every mode that loads ADJ.

## Modes

CHIMERA has three modes, selected via `-mode` or as the first positional argument:

```sh
./CHIMERA -mode <daemon|tui|remote> [-config path/to/config.json]
./CHIMERA <daemon|tui|remote>
```

Default mode is `daemon`. Default config is `./config.json`.

### `daemon`

Headless mode. Loads ADJ, configures the network, starts every plate runtime, and exposes a TCP control server on `network.control-port`. Connect to it with the `remote` mode from another terminal.

```sh
sudo ./CHIMERA daemon
```

Stops cleanly on `Ctrl+C` / `SIGTERM`.

### `tui`

Interactive mode. Same setup as `daemon`, but instead of a TCP server it runs the TUI prompt directly in the current terminal. The banner shows the loaded ADJ branch and commit.

```sh
sudo ./CHIMERA tui
```

TUI commands:

- `help` / `h` — show command help
- `list [board] [packets|measurements]` — list registered boards, packets, or measurements
- `set <board> <measurementId> <value>` — override a measurement value
- `test TCP-abrupt <board>` — run a built-in test
- `quit` / `exit` / `bye` — clean up boards and exit

### `remote`

Thin TUI client. Does **not** load ADJ or touch the network — it just connects to a running `daemon` over TCP and forwards the same commands.

```sh
./CHIMERA remote
```

The `network.control-port` in the client's config must match the daemon's.

## Privileges

`daemon` and `tui` configure network interfaces and bind board IPs, so they typically need to run as root (or with `CAP_NET_ADMIN`). `remote` does not.

## Typical workflows

**Local development, single terminal:**
```sh
sudo ./CHIMERA tui
```

**Headless run with a separate control terminal:**
```sh
# terminal 1
sudo ./CHIMERA daemon

# terminal 2
./CHIMERA remote
```

**Offline / no internet:** keep an `adj` directory next to the executable. CHIMERA will fall back to it automatically when the clone fails.
