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
scp monitor host:/srv/app-monitor/monitor
scp config.json host:/srv/app-monitor/config.json
```

### Run as daemon 

```bash
daemon -p /var/run/app-monitor.pid -o /var/log/app-monitor.log /srv/app-monitor/monitor /srv/app-monitor/config.json
```
