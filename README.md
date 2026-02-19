# LocalSend Go

A CLI/TUI implementation of the [LocalSend protocol](https://github.com/localsend/protocol) for local network file transfer. Compatible with official LocalSend apps on all platforms.

## Install

### From source

```bash
git clone https://github.com/meowrain/localsend-go.git
cd localsend-go
make build
```

Binaries are output to `./bin/` for all supported platforms (Linux, macOS, Windows on amd64/arm64/riscv64).

### Using `go install`

```bash
go install github.com/meowrain/localsend-go@latest
```

### Arch Linux

```bash
yay -S localsend-go
```

## Usage

Running without arguments starts an interactive TUI where you can choose to send, receive, or start the web server.

```
Usage: localsend-go [options] <command> [arguments]

Commands:
  send <file_path>    Send a file to a device on the local network
  receive             Wait for incoming files from other devices
  web                 Start the web file server with QR code
  help                Display help information

Options:
  --help              Display help information
  --port=<number>     Specify server port (default: 53317)
  --config=<path>     Specify config file path (default: ./localsend.yaml)
```

### Examples

```bash
# Send a file (interactive device selection)
localsend-go send photo.jpg

# Send a file using an absolute path
localsend-go send /path/to/file.zip

# Receive files from other devices
localsend-go receive

# Start the web file server on a custom port
localsend-go --port=8080 web

# Use a custom config file
localsend-go --config=/etc/localsend-go/localsend.yaml receive
```

## Configuration

By default, localsend-go looks for `./localsend.yaml` in the working directory. If not found, it falls back to the embedded default configuration. You can specify a custom path with `--config`.

Example `localsend.yaml`:

```yaml
# Device name displayed to other devices on the network.
# If empty, a random name will be generated (e.g. "Happy Phoenix").
device_name: ""

# Directory where received files will be saved.
save_dir: "./uploads"

functions:
  # Enable the HTTP file server (web mode).
  http_file_server: true
  # Enable the LocalSend protocol server (send/receive mode).
  local_send_server: true
```

## Running as a systemd service

A systemd unit file is included for running localsend-go in receive mode as a background service.

### Setup

```bash
# Copy the binary
sudo cp bin/localsend-go-linux-amd64 /usr/local/bin/localsend-go

# Copy the service file
sudo cp localsend-go.service /etc/systemd/system/

# (Optional) Place a config file — otherwise embedded defaults are used
sudo mkdir -p /var/lib/localsend-go
sudo cp localsend.yaml /var/lib/localsend-go/

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable --now localsend-go
```

Received files are saved to `/var/lib/localsend-go/uploads/` by default (configurable via `save_dir` in the config file). The service's `WorkingDirectory` is `/var/lib/localsend-go`, so the default `./localsend.yaml` lookup works if you place a config file there.

### Managing the service

```bash
sudo systemctl status localsend-go    # Check status
sudo systemctl stop localsend-go      # Stop
sudo systemctl restart localsend-go   # Restart
journalctl -u localsend-go -f         # Follow logs
```

## Version

- v1.3.0 - Current version
- [v1.1.0](doc/version1.1.0/) - Historical version

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=meowrain/localsend-go&type=Date)](https://www.star-history.com/#meowrain/localsend-go&Date)
