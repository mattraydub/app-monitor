# Application Monitor

A Go-based service monitoring tool that checks the health of web applications and sends email notifications when issues are detected.

## Features

- Monitors multiple web applications simultaneously
- Configurable check intervals
- Email notifications for service outages
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
export GOOS=freebsd
export GOARCH=amd64
go build -o monitor main.go
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
