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

Talos k8s cluster (full rebuild from scratch):

```bash
cd tofu/talos
tofu apply -target=talos_machine_secrets.this   # create secrets first (OpenTofu provider bug workaround)
tofu apply                                       # VMs + config apply + bootstrap
tofu output -raw talosconfig > ~/repos/homelab/talos/talosconfig
tofu output -raw kubeconfig > ~/.kube/config

cd ../../ansible
ansible-playbook playbooks/talos-setup.yml       # Longhorn, ArgoCD, root app
```

After setup, apps are deployed via gitops: add manifests to `k8s/<app>/` and an ArgoCD Application to `k8s/argocd/`, commit and push.

Other infrastructure provisioning (per-host directories):

```bash
cd tofu/srv1   # or tofu/srv2
tofu plan      # review changes
tofu apply     # apply changes
```

Each host has its own `.env` with API credentials, loaded automatically by direnv.

## Architecture

**Monitoring pipeline:** Grafana Alloy runs on all Linux hosts, pushing node metrics to Prometheus via remote_write and journal logs to Loki. Prometheus also scrapes blackbox-exporter for HTTP/HTTPS/TCP probes. Grafana reads from Prometheus and Loki, rendering dashboards provisioned from JSON files.

**Hosts:** srv1/srv2/srv3 are Proxmox hypervisors. pve1/pve2 are service aliases for Proxmox web UIs through the nginx reverse proxy. OPNsense router uses a community Prometheus exporter (FreeBSD, can't run Alloy).

**Talos k8s cluster:** "dexter" on VLAN 30. VMs provisioned and configured by OpenTofu (`tofu/talos/`), cluster services (Longhorn, ArgoCD) installed by Ansible (`ansible/playbooks/talos-setup.yml`). Apps deployed via ArgoCD gitops — manifests in `k8s/<app>/`, ArgoCD Applications in `k8s/argocd/`. Root app auto-syncs everything in `k8s/argocd/`.

**k8s in-cluster monitoring:** Self-contained stack in `monitoring` namespace, independent from lab monitoring. Alloy (DaemonSet) → Prometheus + Loki ← Grafana. Dashboards are JSON in `k8s/monitoring/dashboards/`, wrapped as ConfigMaps by Kustomize, loaded by Grafana sidecar. All deployed via ArgoCD Helm Applications.

## Key Conventions

### Lab infrastructure (monitoring-ct)
- **Grafana dashboards** are hand-authored JSON in `grafana/dashboards/`. The provisioner auto-loads any `.json` file from `/var/lib/grafana/dashboards/`.
- **Alerting rules** are in `grafana/provisioning/alerting/homelab.yml` (Grafana unified alerting format, not Prometheus alertrules).
- **Blackbox modules** (`prometheus/blackbox.yml`): `http_2xx` for plain HTTP, `http_2xx_insecure` for HTTPS with self-signed certs (Proxmox UIs), `tcp_connect` for port checks.
- **Target files** use a `service` label for blackbox probes. Host metrics come from Alloy via remote_write with a `host` label.

### k8s platform (in-cluster)
- **ArgoCD Applications** live in `k8s/argocd/`. Root app auto-syncs this directory. Adding an app = add YAML here + manifests in `k8s/<app>/`, commit and push.
- **k8s Grafana dashboards** are JSON in `k8s/monitoring/dashboards/`. Kustomize generates labeled ConfigMaps; Grafana sidecar auto-loads them. Reference datasources by type only (`{"type": "prometheus"}`), not UID.
- **Helm apps** (Alloy, Prometheus, Loki, Grafana) are configured inline in their ArgoCD Application YAMLs via `valuesObject`.

### General
- **Sealed Secrets** encrypt k8s secrets for safe git storage. Controller runs in `kube-system`, decrypts SealedSecret resources into regular Secrets. Create with: `kubectl create secret ... --dry-run=client -o yaml | kubeseal --controller-name sealed-secrets --controller-namespace kube-system -o yaml`
- **Container images** are built with Podman on dev-ct (`sudo podman build`) and pushed to `ghcr.io/taylormeador/`. Clusters pull via `imagePullSecrets` referencing a sealed `ghcr-pull` secret.
- **Ansible vault** stores OPNsense API credentials. Always pass `--ask-vault-pass` when running playbooks.
- All hosts are on the `10.0.10.0/24` network. Talos cluster is on `10.0.30.0/24` (VLAN 30).

## Dexter (API Server + CLI)

Go gRPC API server (`dexter/cmd/dexterd/`) and CLI (`dexter/cmd/dexter/`). Proto definitions in `dexter/proto/dexter/v1/dexter.proto`, generated code in `dexter/internal/gen/`. Server logic in `dexter/internal/server/`.

Regenerate protobuf code:
```bash
cd dexter
protoc --proto_path=proto --go_out=internal/gen --go_opt=module=github.com/taylormeador/homelab/dexter/internal/gen --go-grpc_out=internal/gen --go-grpc_opt=module=github.com/taylormeador/homelab/dexter/internal/gen proto/dexter/v1/dexter.proto
```

Build CLI locally:
```bash
cd dexter
go build -o ~/.local/bin/dexter ./cmd/dexter/
```

Build and deploy server:
```bash
cd dexter
sudo podman build -t ghcr.io/taylormeador/dexterd:latest .
sudo podman login ghcr.io -u taylormeador --password-stdin
sudo podman push ghcr.io/taylormeador/dexterd:latest
kubectl rollout restart deployment dexterd -n dexter
```

k8s manifests in `k8s/dexter/`, ArgoCD app in `k8s/argocd/dexter-app.yaml`. Environment variables loaded via direnv (`.envrc` → `.env`) locally, k8s Secrets in cluster.

## Network and DNS

Internal DNS zone `lab.taylor-meador.com` is served by Unbound on OPNsense (not published to Cloudflare). Web services resolve to the nginx reverse-proxy CT (`10.0.10.120`); non-HTTP services resolve directly to their host. See `docs/dns-and-reverse-proxy.md` for setup procedures and gotchas.
