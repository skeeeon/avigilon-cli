# Avigilon CLI & Exporter

An unofficial command-line interface and Prometheus exporter for the Avigilon Web Endpoint Service. This tool allows system administrators to manage resources via scripts, audit history, automate responses, and monitor VMS health via Prometheus/Grafana.

## Features

*   **Authentication**: Handles the complex Nonce/Key hashing and session management required by Avigilon WEP.
*   **Camera Management**: List cameras, view connection status, download JPEG snapshots, and **trigger manual recordings**.
*   **Alarm Management**: Monitor active alarms and perform actions (Acknowledge, Purge, Dismiss).
*   **Event Search**: Query historical events across all servers in a cluster (Motion, Login, Errors, etc.).
*   **Output Control**: Trigger digital outputs connected to cameras or I/O modules.
*   **Webhook Management**: Full CRUD support for event subscription webhooks.
*   **Prometheus Exporter**: A built-in daemon (installable as a system service) that exposes System Health, Camera Status, Recording Integrity, and Alarm counts.

## Prerequisites

To use this tool, you must have an User None and Key provided by Motorola, as well as a user with appropriate permissions configured in Avigilon.

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/skeeeon/avigilon-cli.git
cd avigilon-cli

# Build the binary
go build -o avigilon-cli main.go

# (Optional) Move to path
mv avigilon-cli /usr/local/bin/
```

## Quick Start: Authentication

The CLI maintains a session token in `~/.avigilon-cli.yaml`. You must log in once to initialize the configuration.

**Critical:** Ensure your system time is synced via NTP. The authentication hash relies on the system clock; time drift > 5 minutes will cause `403 Forbidden` errors.

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

# JSON output for scripting (jq compatible)
./avigilon-cli cameras list --json
```

Take a snapshot:
```bash
./avigilon-cli cameras snapshot --id "camera_id" --output "parking_lot.jpg"
```

Trigger/Stop Manual Recording:
```bash
# Start a 5-minute recording
./avigilon-cli cameras record --ids "cam_id_1,cam_id_2" --seconds 300

# Stop recording immediately
./avigilon-cli cameras record --ids "cam_id_1" --stop
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

### Historical Events
Search for events across all servers in the cluster.
```bash
# List user events from the last hour (default)
./avigilon-cli events list --topics "USER"

# List application events from the last 24 hours
./avigilon-cli events list --since 24h --topics "APPLICATION"

# Filter by multiple specific topics
./avigilon-cli events list --since 4h --topics "DEVICE_MOTION_START, USER_LOGIN"
```

### Digital Outputs
Trigger a digital output.
```bash
# Trigger all outputs on a specific camera
./avigilon-cli outputs trigger --id "camera_id" --camera

# Trigger a specific digital output entity
./avigilon-cli outputs trigger --id "output_entity_id"
```

### Webhooks
Manage event subscriptions.
```bash
# Create a webhook
./avigilon-cli webhooks create \
  --url "http://myserver.com/events" \
  --topics "MOTION,ALARM" \
  --heartbeat=true \
  --heartbeat-freq 60000

# List webhooks
./avigilon-cli webhooks list

# Delete a webhook
./avigilon-cli webhooks delete --id "webhook_id"
```

---

## Prometheus Exporter

The tool includes a built-in exporter mode. This runs as a long-lived process that scrapes the Avigilon API and exposes metrics for Prometheus.

**Note:** The exporter handles its own session management. It will automatically re-authenticate if the session expires.

### Running Interactively
```bash
./avigilon-cli exporter \
  --host "https://192.168.1.50/mt/api/rest/v1" \
  --username "administrator" \
  --password "myPassword" \
  --nonce "myUserNonce" \
  --key "myUserKey" \
  --port 9100
```

### Installing as a System Service
The tool can install itself as a service (systemd on Linux, Windows Service on Windows).

```bash
# Install (Must run as Root/Admin)
sudo ./avigilon-cli exporter --service install \
  --host "..." --username "..." --password "..." --nonce "..." --key "..."

# Start
sudo ./avigilon-cli exporter --service start

# Uninstall
sudo ./avigilon-cli exporter --service stop
sudo ./avigilon-cli exporter --service uninstall
```

### Exposed Metrics

Metrics are available at `http://localhost:9100/metrics`.

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `avigilon_up` | Gauge | None | 1 if the API scrape was successful, 0 otherwise. |
| `avigilon_system_health` | Gauge | None | 1.0 = GOOD, 0.5 = WARN, 0.0 = BAD. |
| `avigilon_servers_total` | Gauge | None | Number of servers detected in the cluster. |
| `avigilon_camera_up` | Gauge | `id`, `name`, `model`, `ip` | 1 if camera is CONNECTED, 0 otherwise. |
| `avigilon_camera_has_recorded_data` | Gauge | `id`, `name` | 1 if the camera has recorded footage on the timeline. |
| `avigilon_cameras_total` | Gauge | `state` | Aggregate count of cameras by state (e.g., CONNECTED, DISCONNECTED). |
| `avigilon_alarms_total` | Gauge | `state` | Aggregate count of alarms by state (e.g., ACTIVE, PURGED). |
| `avigilon_scrape_duration_seconds` | Gauge | None | Time taken to interact with the API. |

## Troubleshooting

*   **403 Forbidden on Login:** Check your system time. The authentication hash includes a timestamp. If your machine time differs from the ACC Server time by more than 5-10 minutes, the token is rejected.
*   **400 Bad Request on Webhooks:** Ensure your target URL is valid.
*   **Certificate Errors:** The client is configured to skip TLS verification (`InsecureSkipVerify: true`) by default to support on-prem servers with self-signed certificates.

## Disclaimer

This is an unofficial tool and is not affiliated with or endorsed by Motorola Solutions or Avigilon. Use at your own risk.
```
