# safplatform

Minimal Go HTTP service that echoes back every request it receives. Any path and any method return a JSON payload with the request headers, query params, body and path.

## Stack

| Item | Value |
| --- | --- |
| Language | Go 1.27 |
| Router | `github.com/go-chi/chi/v5` |
| Runtime image | `gcr.io/distroless/static-debian12:nonroot` |
| Delivery | GitHub Actions + ArgoCD (GitOps) |

## Design decisions

- **chi** instead of the standard `net/http` mux — keeps routing and middleware code simpler and more readable.
- **ArgoCD** for delivery — adopts the market-standard GitOps flow.

## API

`ANY /*` → `200 application/json`

```json
{
  "Headers": { "Accept": ["*/*"] },
  "Params": { "q": ["1"] },
  "Body": "",
  "Path": "/anything"
}
```

- `Body` is returned as raw JSON when the payload is valid JSON, otherwise as a string.
- `PORT` env var overrides the default port `8080`.
- Graceful shutdown on `SIGINT`/`SIGTERM` (10s grace).

## Local development

```bash
go run .          # start on :8080
scripts/test.sh   # go vet + go test
```

Recommended: open the repo in VS Code and reopen in the devcontainer.

## Folder architecture

| Path | Purpose | Improvement points |
| --- | --- | --- |
| `.devcontainer` | Setup to build the app locally without installing dependencies on the host, and to guarantee IntelliSense in VS Code. | — |
| `.github` | CI/CD workflow, calling reusable workflows from [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp). | — |
| `.idp` | Helm chart applied into Kubernetes by ArgoCD, depending on the shared `web` chart published from [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp). | — |
| `scripts` | Local test script (`go vet` + `go test`), not used by CI/CD anymore. | Remove once test also runs through a reusable workflow. |
| `Dockerfile` | Production workload image — used by the build job in the CI/CD pipeline. | — |
| `main.go` | Main application file containing the code. | — |

## CI/CD

Triggered on push ([.github/workflows/cicd.yaml](.github/workflows/cicd.yaml)), delegating to reusable workflows from [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp):

1. **build** — [`build.yaml`](https://github.com/jasondavindev/open-idp/blob/main/.github/workflows/build.yaml) builds and pushes `<registry>/safplatform:<commit-sha>`.
2. **deploy** — [`deploy.yaml`](https://github.com/jasondavindev/open-idp/blob/main/.github/workflows/deploy.yaml) bumps `global.image.tag` in [.idp/safplatform/values.yaml](.idp/safplatform/values.yaml) to that SHA and commits it back to this repository; ArgoCD syncs the new tag and rolls out a new version of the Deployment in Kubernetes.

Required secrets: `CONTAINER_REGISTRY`, `REPO_USER`, `REPO_PASSWORD`.

Improvement points:

- Instead of committing to this repository, the deploy job should commit to [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp) — the central catalog of all apps installed in the cluster.

## Screenshots

ArgoCD syncing the tag committed by the deploy job and rolling out the new ReplicaSet:

![ArgoCD application synced and rolling out](docs/images/argocd.png)

Request to the app running in the cluster, behind Traefik:

![curl request to the app returning the echoed JSON](docs/images/app.png)
