# Project Roadmap

Ideas and future work for the homelab platform. Pick up where it makes sense.

## Phase 1: Kubernetes + ArgoCD

Stand up a k8s cluster on Proxmox VMs (provisioned by tofu) with ArgoCD for GitOps deployments. This becomes the platform layer that new services deploy onto — separate from the existing Ansible-managed infrastructure that runs critical services.

- k3s or kubeadm cluster on Proxmox VMs, provisioned via tofu
- ArgoCD watches a GitOps repo, deploys apps from manifests/Helm charts
- Alloy DaemonSet for metrics + logs shipping to existing Prometheus/Loki
- Ingress controller + Cloudflare tunnel for external access
- This is the foundation everything else builds on

## Phase 2: Go CLI (`lab`)

Build a CLI that manages the full lifecycle of services, backed by Terraform modules and ArgoCD.

**Key feature: local vs cloud provisioning**
- `lab deploy --target local` provisions on Proxmox k8s cluster
- `lab deploy --target aws` provisions on EKS/EC2
- Both environments route to the same observability stack (Prometheus, Grafana, Loki)
- Same CLI, same dashboards, different infrastructure underneath
- The hard part is making the abstraction work cleanly across both — different Terraform modules, networking, log/metric shipping

**Lifecycle management (the impressive part)**

One command should handle:
- VM/CT or cloud resource provisioning
- App deployment (commit to GitOps repo, ArgoCD syncs)
- Prometheus target registration
- Grafana dashboard creation
- Log shipping setup (Loki)
- DNS/Cloudflare tunnel entry
- Portal registration
- Teardown cleans up all of the above

**Implementation notes:**
- Cobra for CLI framework
- Thin interface: tofu for infra, GitOps commits for app deployment
- Replaces the "I need a Go project" gap — bounded, practical, teaches Go in a real context

## Phase 3: Portal + Demo

### taylor-meador.com as Living Infrastructure Demo

One site serving as both public-facing landing page and gateway into infrastructure:

- **Landing page:** who you are, architecture overview, links to live tools
- **Grafana:** anonymous viewer mode for curated demo dashboards (public)
- **Portal:** behind Cloudflare Access -> provisioning UI, internal tools
- **Additional services** added as they're built

**Interview demo:** Deploy a full service stack live during an interview via the portal. Provision it, show it immediately appearing in monitoring, logging, and the portal. Tear it down, everything cleans up.

### Public architecture page
Diagram of homelab topology, services, design decisions. Easy to maintain, shows systems thinking.

## Service Ideas

### Game servers on demand
Spin up Minecraft/Valheim/etc servers via the portal, auto-shutdown after idle. Already started on Minecraft provisioning.

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
Self-hosted ChatGPT-style interface connected to the LLM gateway. Accessible through the portal.
