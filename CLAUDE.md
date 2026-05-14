# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build (native)
go build -o monitor main.go

# Build for FreeBSD
GOOS=freebsd GOARCH=amd64 go build -o monitor main.go

# Run
./monitor config.json

# Run tests
go test ./...

# Run a single test
go test -run TestName ./...
```

## Architecture

The entire service lives in `main.go` — no packages, no subdirectories. It is a single `Monitor` struct with an HTTP client, a JSON config, and an in-memory `alertTracker` map.

**Alert state machine** (`alertTracker map[string]int`, keyed by app name):
- `0` — healthy / no active alert
- `1` — one failure observed, no notification sent yet
- `2` — second consecutive failure; triggers email + webhook (immediately bumped to `3`)
- `3+` — alert already sent; suppresses further notifications until recovery resets to `0`

Recovery (`sendRecoveryNotice`) only fires if the counter is ≥ 3, meaning an alert was actually sent. This prevents spurious recovery emails for transient single-check failures.

**Check loop**: `runChecks` fires all application checks concurrently via goroutines + `sync.WaitGroup` on every ticker tick. The HTTP client has a fixed 10-second timeout.

**Notifications**: both `sendAlert` and `sendRecoveryNotice` send email (SMTP plain auth) and optionally a webhook (HTTP POST JSON). Webhook payloads can be signed with HMAC-SHA256 via the `X-AppMonitor-Signature` header when a secret is configured.

## Service scripts

`init/freebsd/app_monitor` — rc.d script; install to `/usr/local/etc/rc.d/`, enable with `app_monitor_enable="YES"` in `/etc/rc.conf`.

`init/devuan/app-monitor` — sysvinit LSB script; install to `/etc/init.d/`, enable with `update-rc.d app-monitor defaults`.

FreeBSD paths: binary at `/usr/local/sbin/app-monitor`, config at `/usr/local/etc/app-monitor/config.json`. Linux paths: binary at `/usr/local/sbin/app-monitor`, config at `/etc/app-monitor/config.json`.
