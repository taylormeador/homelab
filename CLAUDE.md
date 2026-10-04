# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Config and documentation repo for a Proxmox-based homelab. Config files are deployed by running `bootstrap.sh` on the monitoring host; operational docs live in `docs/`. There is no build system or test suite.

## Deploying

```bash
./bootstrap.sh   # copies all configs to system paths, restarts prometheus + blackbox-exporter + grafana
```

This must run as root (or with sudo) on the monitoring container (monitoring-ct). It installs files with `install(1)` so permissions are set explicitly.

## Architecture

**Monitoring pipeline:** Prometheus scrapes node-exporter (host metrics) and blackbox-exporter (HTTP/HTTPS/TCP probes) using file-based service discovery. Targets are defined in `prometheus/targets/*.yml`. Grafana reads from Prometheus and renders dashboards provisioned from JSON files.

**Backup observability:** The `common/bin/backup-run` wrapper runs any backup command and writes `.prom` textfiles to the node-exporter directory (`/var/lib/prometheus/node-exporter/`). Prometheus picks these up as `backup_last_success_timestamp_seconds` etc., which feeds both the dashboard's "Backup age" panel and the stale-backup alert.

**Host-specific scripts:** `hosts/<hostname>/` contains scripts and cron entries for individual machines (e.g., the Minecraft restic backup).

## Key Conventions

- **Grafana dashboards** are hand-authored JSON in `grafana/dashboards/`. The provisioner auto-loads any `.json` file from `/var/lib/grafana/dashboards/`.
- **Alerting rules** are in `grafana/provisioning/alerting/homelab.yml` (Grafana unified alerting format, not Prometheus alertrules).
- **Blackbox modules** (`prometheus/blackbox.yml`): `http_2xx` for plain HTTP, `http_2xx_insecure` for HTTPS with self-signed certs (Proxmox UIs), `tcp_connect` for port checks.
- **Target files** use a `service` label for probes and a `host` label for node-exporter targets. These labels drive dashboard legend and alert annotations.
- All hosts are on the `10.0.10.0/24` network.

## Network and DNS

Internal DNS zone `lab.taylor-meador.com` is served by Unbound on OPNsense (not published to Cloudflare). Web services resolve to the nginx reverse-proxy CT (`10.0.10.120`); non-HTTP services resolve directly to their host. See `docs/dns-and-reverse-proxy.md` for setup procedures and gotchas.
