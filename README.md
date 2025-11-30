# Avigilon CLI & Exporter

An unofficial command-line interface and Prometheus exporter for the Avigilon Web Endpoint Service. This tool allows system administrators to manage resources via scripts, audit history, automate responses, and monitor VMS health via Prometheus/Grafana.

## Features

*   **Authentication**: Handles the complex Nonce/Key hashing and session management required by Avigilon WEP.
*   **Camera Management**: List cameras, view connection status, download JPEG snapshots, and **trigger manual recordings**.
*   **Alarm Management**: Monitor active alarms and perform actions (Acknowledge, Purge, Dismiss).
*   **Event Search**: Query historical events across all servers in a cluster (Motion, Login, Errors, etc.).
*   **Output Control**: Trigger digital outputs connected to cameras or I/O modules.
*   **Webhook Management**: Full CRUD support for event subscription webhooks.
*   **Prometheus Exporter**: A built-in daemon that exposes System Health, Camera Status, Recording Integrity, and Alarm counts.
    *  Supports secure configuration via Windows Registry or Environment Variables to keep credentials out of process lists.

## Prerequisites

To use this tool, you must have:
1.  **Integration Credentials:** A `User Nonce` and `User Key` provided by Motorola.
2.  **User Account:** A dedicated Avigilon user with appropriate permissions.
3.  **Time Sync:** **Critical.** The machine running this CLI must be time-synced (NTP) with the Avigilon Server. Time drift > 5 minutes will cause `403 Forbidden` errors.

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/skeeeon/avigilon-cli.git
cd avigilon-cli

# Build the binary
go build -o avigilon-cli main.go

# (Optional) Move to path (Linux) or C:\Windows\System32 (Windows)
mv avigilon-cli /usr/local/bin/
```

## Quick Start: CLI Usage

The CLI maintains a session token in `~/.avigilon-cli.yaml`. You must log in once to initialize the configuration.

### Interactive Login

You can pass credentials via flags or let your shell handle environment variables.

```bash
# Option 1: Flags
./avigilon-cli login \
  --host "https://192.168.1.50/mt/api/rest/v1" \
  --username "administrator" \
  --password "myPassword" \
  --nonce "myUserNonce" \
  --key "myUserKey"

# Option 2: Environment Variables (Bash/PowerShell)
# Bash: export AVIGILON_PASSWORD="myPassword"
# PS:   $env:AVIGILON_PASSWORD="myPassword"
./avigilon-cli login --host "..." --username "admin" --nonce "..." --key "..."
```

### Common Commands

**Cameras**
```bash
# List all cameras (JSON format for scripting)
./avigilon-cli cameras list --json

# Take a snapshot
./avigilon-cli cameras snapshot --id "camera-id-123" --output "parking.jpg"

# Trigger a 5-minute manual recording
./avigilon-cli cameras record --ids "camera-id-123" --seconds 300
```

**Alarms & Events**
```bash
# List active alarms
./avigilon-cli alarms list

# Search for motion events in the last 4 hours
./avigilon-cli events list --since 4h --topics "DEVICE_MOTION_START"
```

---

## Prometheus Exporter

The tool runs as a long-lived service to scrape metrics for Prometheus. It handles session auto-renewal automatically.

### Option 1: Interactive Run (Testing)
Useful for verifying connectivity before installing as a service.
```bash
./avigilon-cli exporter \
  --host "https://192.168.1.50/mt/api/rest/v1" \
  --username "administrator" \
  --password "myPassword" \
  --nonce "myUserNonce" \
  --key "myUserKey" \
  --port 9100
```

### Option 2: Windows Service (Secure Installation)
**Recommended.** This method prevents your password from appearing in the Windows Service properties or process list.

1.  **Install the Service** (Run as Administrator):
    *Do not pass credentials here.*
    ```powershell
    .\avigilon-cli.exe exporter --service install --host "https://192.168.1.50/mt/api/rest/v1"
    ```

2.  **Configure Credentials via Registry**:
    *   Open `regedit`.
    *   Navigate to: `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\avigilon-exporter`
    *   Right-click the `avigilon-exporter` folder -> **New** -> **Multi-String Value**.
    *   Name it: `Environment`
    *   Double-click to edit and add your credentials (one per line):
        ```text
        AVIGILON_HOST=https://192.168.1.50/mt/api/rest/v1
        AVIGILON_USERNAME=administrator
        AVIGILON_PASSWORD=YourSecurePassword
        AVIGILON_NONCE=YourUserNonce
        AVIGILON_KEY=YourUserKey
        ```

3.  **Start the Service**:
    ```powershell
    .\avigilon-cli.exe exporter --service start
    ```

### Option 3: Linux Systemd Service
```bash
# Install
sudo ./avigilon-cli exporter --service install

# Edit the service file to add Environment variables securely
sudo systemctl edit avigilon-exporter
# Add:
# [Service]
# Environment="AVIGILON_HOST=..."
# Environment="AVIGILON_PASSWORD=..."

# Start
sudo ./avigilon-cli exporter --service start
```

### Exposed Metrics
Metrics are available at `http://localhost:9100/metrics`.

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `avigilon_up` | Gauge | None | 1 if API scrape succeeded. |
| `avigilon_system_health` | Gauge | None | 1.0 (GOOD), 0.5 (WARN), 0.0 (BAD). |
| `avigilon_camera_up` | Gauge | `id`, `name`, `ip` | 1 if Connected, 0 if Disconnected. |
| `avigilon_camera_has_recorded_data` | Gauge | `id`, `name` | 1 if recording exists on timeline. |
| `avigilon_alarms_total` | Gauge | `state` | Count of alarms by state (ACTIVE, PURGED). |

## Troubleshooting

*   **Service fails to start:** Check the Windows Event Viewer or syslog. If you installed using the "Secure" method, ensure you created the `Environment` registry key correctly as a **Multi-String Value** (REG_MULTI_SZ).
*   **403 Forbidden:** Check system time. The authentication hash is time-sensitive.
*   **TLS Errors:** The client defaults to `InsecureSkipVerify: true` to support self-signed certificates common on VMS appliances.

## Disclaimer

This is an unofficial tool and is not affiliated with or endorsed by Motorola Solutions or Avigilon. Use at your own risk.
