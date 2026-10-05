# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Config and documentation repo for a Proxmox-based homelab. Infrastructure is provisioned with OpenTofu and configured with Ansible. Operational docs live in `docs/`.

## Deploying

All configuration is deployed via Ansible playbooks, run from the `ansible/` directory:

```bash
cd ansible
ansible-playbook playbooks/<playbook>.yml --ask-become-pass --ask-vault-pass
```

Key playbooks:
- `logging.yml` — Loki + monitoring role on monitoring-ct, Alloy on all nodes
- `dns.yml` — Unbound DNS overrides on OPNsense
- `nginx.yml` — Reverse proxy configs on nginx-ct
- `backup-opnsense.yml` — OPNsense config backup

Infrastructure provisioning:

```bash
cd tofu
tofu plan    # review changes
tofu apply   # apply changes
```

## Architecture

**Monitoring pipeline:** Grafana Alloy runs on all Linux hosts, pushing node metrics to Prometheus via remote_write and journal logs to Loki. Prometheus also scrapes blackbox-exporter for HTTP/HTTPS/TCP probes. Grafana reads from Prometheus and Loki, rendering dashboards provisioned from JSON files.

**Hosts:** srv1/srv2 are Proxmox hypervisors (srv1 currently offline). pve1/pve2 are service aliases for Proxmox web UIs through the nginx reverse proxy. OPNsense router uses a community Prometheus exporter (FreeBSD, can't run Alloy).

## Key Conventions

- **Grafana dashboards** are hand-authored JSON in `grafana/dashboards/`. The provisioner auto-loads any `.json` file from `/var/lib/grafana/dashboards/`.
- **Alerting rules** are in `grafana/provisioning/alerting/homelab.yml` (Grafana unified alerting format, not Prometheus alertrules).
- **Blackbox modules** (`prometheus/blackbox.yml`): `http_2xx` for plain HTTP, `http_2xx_insecure` for HTTPS with self-signed certs (Proxmox UIs), `tcp_connect` for port checks.
- **Target files** use a `service` label for blackbox probes. Host metrics come from Alloy via remote_write with a `host` label.
- **Ansible vault** stores OPNsense API credentials. Always pass `--ask-vault-pass` when running playbooks.
- All hosts are on the `10.0.10.0/24` network.

## Network and DNS

Internal DNS zone `lab.taylor-meador.com` is served by Unbound on OPNsense (not published to Cloudflare). Web services resolve to the nginx reverse-proxy CT (`10.0.10.120`); non-HTTP services resolve directly to their host. See `docs/dns-and-reverse-proxy.md` for setup procedures and gotchas.
