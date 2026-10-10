# ArgoCD Setup

ArgoCD is installed manually since it manages everything else in the cluster.

## Install

```bash
kubectl create namespace argocd

kubectl apply -n argocd --server-side --force-conflicts \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.2.0/manifests/install.yaml
```

## Configure

Disable internal TLS (nginx handles HTTP proxying):

```bash
kubectl -n argocd patch configmap argocd-cmd-params-cm \
  --type merge -p '{"data": {"server.insecure": "true"}}'
kubectl -n argocd rollout restart deployment argocd-server
```

Expose via NodePort:

```bash
kubectl patch svc argocd-server -n argocd \
  -p '{"spec": {"type": "NodePort", "ports": [{"port": 443, "targetPort": 8080, "nodePort": 30443, "protocol": "TCP", "name": "https"}]}}'
```

DNS and nginx entries are in the Ansible playbooks (`dns.yml`, `nginx.yml`).

## Access

- **Web UI:** http://argocd.lab.taylor-meador.com
- **CLI:** `argocd login argocd.lab.taylor-meador.com:80 --grpc-web --plaintext`
- **Initial admin password:** `kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d`

## Docs

- https://argo-cd.readthedocs.io/en/stable/getting_started/
- https://argo-cd.readthedocs.io/en/stable/operator-manual/ingress/
