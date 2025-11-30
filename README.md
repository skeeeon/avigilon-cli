# Avigilon Control Center CLI & Exporter

A powerful command-line interface and Prometheus exporter for the Avigilon Web Endpoint Service. This tool allows system administrators to manage resources via scripts and monitor VMS health via Prometheus/Grafana.

## Features

*   **Authentication**: Handles the complex Nonce/Key hashing and session management required by Avigilon WEP.
*   **Camera Management**: List cameras, view connection status, IP addresses, and download JPEG snapshots.
*   **Alarm Management**: Monitor active alarms and perform actions (Acknowledge, Purge, Dismiss).
*   **Webhook Management**: Full CRUD support for event subscription webhooks.
*   **Infrastructure Discovery**: List Sites (Clusters) and Servers.
*   **Scripting Support**: All commands support `--json` output for piping into tools like `jq`.
*   **Prometheus Exporter**: A built-in daemon that exposes System Health, Camera Status, Recording Integrity, and Alarm counts.

## Prerequisites

To use this tool, you must have an User None and Key provided by Motorola, as well as a user with appropriate permissions configured in Avigilon.

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/avigilon-cli.git
cd avigilon-cli

# Build the binary
go build -o avigilon-cli main.go

# (Optional) Move to path
mv avigilon-cli /usr/local/bin/
```

## Quick Start: Authentication

The CLI maintains a session token in `~/.avigilon-cli.yaml`. You must log in once to initialize the configuration.

**Note:** Ensure your system time is synced via NTP. The authentication hash relies on the system clock; time drift > 5 minutes will cause `403 Forbidden` errors.

```bash
./avigilon-cli login \
  --host "https://192.168.1.50:8443/mt/api/rest/v1" \
  --username "administrator" \
  --password "myPassword" \
  --nonce "myUserNonce" \
  --key "myUserKey"
```

Once logged in, you can run other commands without providing credentials.

## CLI Usage

### Cameras
List all cameras, their models, IPs, and connection status.

```bash
# Human readable table
./avigilon-cli cameras list

# JSON output for scripting
./avigilon-cli cameras list --json
```

Take a snapshot from a specific camera:
```bash
./avigilon-cli cameras snapshot --id "camera_id_here" --output "parking_lot.jpg"
```

### Alarms
List active alarms:
```bash
./avigilon-cli alarms list
```

Acknowledge or Purge an alarm:
```bash
./avigilon-cli alarms update --id "alarm_id" --action "ACKNOWLEDGE" --note "Investigated by Security"
./avigilon-cli alarms update --id "alarm_id" --action "PURGE"
```

### Webhooks
Manage event subscriptions (e.g., sending motion events to an external URL).

```bash
# List webhooks
./avigilon-cli webhooks list

# Create a webhook
./avigilon-cli webhooks create \
  --url "http://myserver.com/events" \
  --topics "MOTION,ALARM" \
  --heartbeat=true \
  --token "my-secret-token"

# Delete a webhook
./avigilon-cli webhooks delete --id "webhook_id"
```

### Infrastructure
```bash
./avigilon-cli sites    # List clusters
./avigilon-cli servers  # List individual servers
```

---

## Prometheus Exporter

The tool includes a built-in exporter mode. This runs as a long-lived process that scrapes the Avigilon API and exposes metrics for Prometheus.

**Note:** The exporter handles its own session management. It will automatically re-authenticate if the session expires (usually every 1 hour).

### Running the Exporter

```bash
./avigilon-cli exporter \
  --host "https://192.168.1.50/mt/api/rest/v1" \
  --username "administrator" \
  --password "myPassword" \
  --nonce "myUserNonce" \
  --key "myUserKey" \
  --port 9100
```

### Exposed Metrics

The exporter exposes the following metrics at `http://localhost:9100/metrics`:

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `avigilon_up` | Gauge | None | 1 if the API scrape was successful, 0 otherwise. |
| `avigilon_system_health` | Gauge | None | 1.0 = GOOD, 0.5 = WARN, 0.0 = BAD. |
| `avigilon_camera_up` | Gauge | `id`, `name`, `model`, `ip` | 1 if camera is CONNECTED, 0 otherwise. |
| `avigilon_camera_has_recorded_data` | Gauge | `id`, `name` | 1 if the camera has recorded footage on the timeline. |
| `avigilon_cameras_total` | Gauge | `state` | Aggregate count of cameras by state (e.g., CONNECTED, DISCONNECTED). |
| `avigilon_alarms_total` | Gauge | `state` | Aggregate count of alarms by state (e.g., ACTIVE, PURGED). |
| `avigilon_servers_total` | Gauge | None | Number of servers detected in the cluster. |
| `avigilon_scrape_duration_seconds` | Gauge | None | Time taken to interact with the API. |

### Prometheus Configuration Example

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'avigilon'
    static_configs:
      - targets: ['localhost:9100']
    scrape_interval: 30s
```

## Troubleshooting

*   **403 Forbidden on Login:** Check your system time. The authentication hash includes a timestamp. If your machine time differs from the ACC Server time by more than 5-10 minutes, the token is rejected.
*   **400 Bad Request on Webhooks:** Ensure you are providing a non-empty `--token`.
*   **Certificate Errors:** The client is configured to skip TLS verification (`InsecureSkipVerify: true`) by default to support on-prem servers with self-signed certificates.

## Disclaimer

This is an unofficial tool and is not affiliated with or endorsed by Motorola Solutions or Avigilon. Use at your own risk.
