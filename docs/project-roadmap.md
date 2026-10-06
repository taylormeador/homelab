# Project Roadmap

Ideas and future work for the homelab platform. Pick up where it makes sense.

## Phase 1: Kubernetes (Talos) + ArgoCD

Stand up a Talos k8s cluster on Proxmox VMs with ArgoCD for GitOps deployments. Set up manually first to learn the pain points, then codify with tofu later.

- Talos Linux VMs on Proxmox (own VLAN, e.g. VLAN 30)
- Longhorn for persistent storage (dedicated virtual disks per node)
- ArgoCD watches GitOps repos, deploys apps from manifests/Helm charts
- Alloy DaemonSet for metrics + logs to existing Prometheus/Loki (firewall rule from k8s VLAN to monitoring-ct)
- Ingress controller + Cloudflare tunnel for external access
- OPNsense firewall rules for inter-VLAN routing to k8s
- Once stable, rebuild with tofu for the production setup

## Phase 2: Go CLI + API Server

### Architecture

```
User's machine              Cluster
┌──────────┐               ┌────────────────────────┐
│  CLI     │──── gRPC ────>│  API server            │
│  (thin   │               │  (auth + authorization)│
│  client) │               │    │                   │
└──────────┘               │    ├─> git push        │
                           │    ├─> ArgoCD CR       │
                           │    └─> k8s API         │
                           │                        │
                           │  All credentials stay  │
                           │  in the cluster        │
                           └────────────────────────┘
```

**Three layers, strict separation:**
- **`pkg/platform/`** — core library: all platform logic (scaffolding, git ops, ArgoCD management, teardown)
- **API server** — runs in k8s, wraps the library, handles auth + authorization, holds all infrastructure credentials. Users never touch git/ArgoCD/kubectl directly.
- **CLI** — thin gRPC client. Knows how to call the API and display results. Contains zero platform logic.

**Why gRPC:** protobuf definitions are the API contract — server and client are generated from the same `.proto` files so they can't drift. Server-streaming gives real-time progress in the CLI ("creating repo... pushing... syncing... live"). It's what the k8s/ArgoCD ecosystem uses internally.

### CLI capabilities

One command handles the full lifecycle:
- `mycli create app --name foo --template go-api` — scaffold, push, ArgoCD sync
- `mycli status foo` — deployment state, endpoints, health
- `mycli logs foo` — stream logs from Loki
- `mycli destroy foo` — tear down everything (app, repo, ArgoCD CR, DNS, monitoring)

Each operation goes through the API server, which enforces what the user is allowed to do.

### Local vs cloud provisioning (stretch goal)
- `--target local` provisions on Proxmox k8s cluster
- `--target aws` provisions on EKS/EC2
- Same CLI, same observability, different infrastructure underneath

## Phase 3: Demo + taylor-meador.com

### Web terminal for demos
Browser-based terminal (ttyd/xterm.js) with the CLI pre-installed. Interviewers can run CLI commands and watch the platform work in real time — no install needed. Way more impressive than a dashboard with buttons.

### taylor-meador.com
- **Web terminal** — live platform demo
- **Grafana** — anonymous viewer mode for curated dashboards
- **Stock app** — real application running on the platform
- **Architecture docs** — topology, design decisions, shows systems thinking

### Interview demo flow
Open the web terminal, run `mycli create app --name demo --template go-api`, watch it provision, show it in Grafana monitoring and Loki logs, then `mycli destroy demo` — everything cleans up.

## Service Ideas

### Game servers on demand
Spin up Minecraft/Valheim/etc servers via the CLI, auto-shutdown after idle.

### VPN / tunnel service
WireGuard-based access for friends or as a personal travel VPN.

## AI Infrastructure

### LLM Gateway (Ollama + LiteLLM)
- Ollama serves local models (Llama 8B, Mistral 7B, Phi-3 — fit in 12GB VRAM on RTX 5070)
- LiteLLM sits in front as unified OpenAI-compatible API proxy
- Routes between local models and cloud APIs (Anthropic, OpenAI)
- Single endpoint for all consumers
- Add Langfuse for observability, usage analytics, prompt tracing
- Routing logic: simple tasks -> local model, complex tasks -> Claude

### Open WebUI
Self-hosted ChatGPT-style interface connected to the LLM gateway. Accessible through the CLI/platform.
