# Application Monitor

A Go-based service monitoring tool that checks the health of web applications and sends email and webhook notifications when issues are detected.

## Features

- Monitors multiple web applications simultaneously
- Configurable check intervals
- Email notifications for service outages
- Webhook notifications with JSON payloads
- HMAC-SHA256 signature support for webhook security
- Alert after two consecutive failures
- Recovery notifications when service is restored
- JSON-based configuration

## Configuration

Create a `config.json` file with the following structure:
```json
{
  "check_interval": "30s",
  "email": {
    "smtp_host": "smtp.example.com",
    "smtp_port": "587",
    "username": "your-email@example.com",
    "password": "your-password",
    "from_email": "monitor@example.com",
    "to_email": "alerts@example.com"
  },
  "webhook": {
    "enabled": true,
    "url": "https://your-webhook-endpoint.com/alerts",
    "secret": "optional-hmac-secret"
  },
  "applications": [
    {
      "name": "Example App",
      "url": "https://example.com/health",
      "enabled": true,
      "expected_code": 200
    }
  ]
}
```

### Webhook Configuration

The webhook feature allows sending HTTP POST notifications to external systems when applications fail or recover.

**Configuration Options:**
- `enabled`: Set to `true` to enable webhook notifications
- `url`: The HTTP endpoint that will receive webhook notifications
- `secret`: (Optional) HMAC secret for payload signature verification

**Webhook Events:**
- `application_down`: Sent after 2 consecutive failures
- `application_recovery`: Sent when application recovers

**Webhook Payload:**
```json
{
  "event": "application_down",
  "application": "Example App",
  "url": "https://example.com/health",
  "timestamp": 1627843200,
  "status_code": 500,
  "expected_code": 200,
  "error": "connection timeout",
  "failure_count": 2
}
```

**Security:**
When a `secret` is configured, webhooks include an `X-AppMonitor-Signature` header with HMAC-SHA256 signature:
```
X-AppMonitor-Signature: sha256=abc123...
```

**Supported Integrations:**
- Zapier
- Discord webhooks
- Slack webhooks
- Custom HTTP endpoints
- Monitoring systems (PagerDuty, etc.)

## Building

```bash
go build -o monitor main.go
```

## Running

```bash
./monitor config.json
```

## Building/Running for FreeBSD environment

```bash
GOOS=freebsd GOARCH=amd64 go build -o monitor main.go
```

### Deploy to remote server

```bash
scp monitor host:/usr/local/sbin/app-monitor
ssh host "mkdir -p /usr/local/etc/app-monitor"
scp config.json host:/usr/local/etc/app-monitor/config.json
```

### Run with rc.d (recommended)

Copy the included rc.d script and enable the service:

```bash
sudo cp init/freebsd/app_monitor /usr/local/etc/rc.d/app_monitor
sudo chmod +x /usr/local/etc/rc.d/app_monitor
```

Add to `/etc/rc.conf`:

```
app_monitor_enable="YES"
```

Optionally override defaults in `/etc/rc.conf`:

```
app_monitor_config="/usr/local/etc/app-monitor/config.json"
app_monitor_logfile="/var/log/app-monitor.log"
```

Manage the service:

```bash
sudo service app_monitor start
sudo service app_monitor stop
sudo service app_monitor restart
sudo service app_monitor status
```

### Run as daemon (manual)

```bash
daemon -p /var/run/app-monitor.pid -o /var/log/app-monitor.log /usr/local/sbin/app-monitor /usr/local/etc/app-monitor/config.json
```

## Running on Devuan Linux (sysvinit)

Build for Linux:

```bash
go build -o monitor main.go
```

Deploy the binary and config:

```bash
sudo cp monitor /usr/local/sbin/app-monitor
sudo chmod +x /usr/local/sbin/app-monitor
sudo mkdir -p /etc/app-monitor
sudo cp config.json /etc/app-monitor/config.json
```

Install the init script:

```bash
sudo cp init/devuan/app-monitor /etc/init.d/app-monitor
sudo chmod +x /etc/init.d/app-monitor
```

Enable the service to start at boot:

```bash
sudo update-rc.d app-monitor defaults
```

Manage the service:

```bash
sudo service app-monitor start
sudo service app-monitor stop
sudo service app-monitor restart
sudo service app-monitor status
```

Logs are written to `/var/log/app-monitor.log`.

To disable autostart:

```bash
sudo update-rc.d app-monitor disable
```
