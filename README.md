# Portfolio DevOps Platform

Your portfolio website, served by a small Go app, deployed and shipped the way
a real production service is — containerized, running on Kubernetes, packaged
with Helm, built by GitHub Actions, and deployed by Argo CD.

## What each piece does, in one line

| Piece | What it does |
|---|---|
| **Go app** (`cmd/server`, `internal/handlers`) | Serves the portfolio's HTML/CSS/images and answers `/healthz` so Kubernetes knows it's alive. |
| **Dockerfile** | Compiles the Go code in one throwaway image, then copies *only the binary* into a tiny, shell-less "distroless" image — nothing extra ships to production. |
| **k8s/*.yaml** | Raw Kubernetes objects: `Deployment` keeps N pods running, `Service` gives them one stable address, `Ingress` lets the outside world reach it. |
| **helm/portfolio** | The same Kubernetes objects, but templated — so `values.yaml` (or `values-prod.yaml`) changes the replica count, image tag, or hostname without editing YAML by hand. |
| **.github/workflows/ci.yml** | On every push: run tests → build the binary → build the Docker image → scan it for vulnerabilities → push it tagged with the CI run number → update the Helm chart's image tag in Git. |
| **argocd/application.yaml** | Tells Argo CD "watch this Git repo's `helm/portfolio` folder and always make the cluster match it" — this is what makes deployment automatic. |

## The flow, end to end

```
you push code
     │
     ▼
GitHub Actions runs tests, builds the Docker image, pushes it, tags it with the run number
     │
     ▼
GitHub Actions commits the new image tag into helm/portfolio/values.yaml
     │
     ▼
Argo CD notices Git changed and syncs the cluster to match
     │
     ▼
Kubernetes rolls out new pods, old ones drain, site is updated — no one ran kubectl
```

## Running it locally (no Kubernetes needed)

```bash
make run
# then open http://localhost:8080
```

## Deploying to a local Kubernetes cluster (Kind)

```bash
kind create cluster
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/
```

Or, using Helm directly instead of raw manifests:

```bash
helm install portfolio ./helm/portfolio
```

## Wiring up GitOps with Argo CD

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl apply -f argocd/application.yaml
```

From that point on, every commit to `helm/portfolio` deploys itself.
