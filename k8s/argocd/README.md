# ArgoCD

Installed by `ansible/playbooks/talos-setup.yml`. All apps are managed via the root app pattern.

## How it works

The setup playbook installs ArgoCD and applies `root-app.yaml`, which watches `k8s/argocd/` with auto-sync and prune. Any Application or ApplicationSet YAML added to this directory is automatically picked up.

## Adding a new app

1. Put manifests in `k8s/<app-name>/`
2. Create `k8s/argocd/<app-name>-app.yaml` (ArgoCD Application pointing at that directory)
3. Commit and push — ArgoCD syncs automatically

## Access

- **Web UI:** http://argocd.lab.taylor-meador.com
- **CLI:** `argocd login argocd.lab.taylor-meador.com:80 --grpc-web --plaintext`
- **Admin password:** `kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d`

## Config details

- Insecure mode (nginx handles TLS termination)
- NodePort 30443
- DNS and nginx entries managed by Ansible (`dns.yml`, `nginx.yml`)
