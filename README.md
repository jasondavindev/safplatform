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
scripts/test.sh   # go vet + go test — same check the CI test job runs
```

Recommended: open the repo in VS Code and reopen in the devcontainer.

## Folder architecture

| Path | Purpose | Improvement points |
| --- | --- | --- |
| `.devcontainer` | Setup to build the app locally without installing dependencies on the host, and to guarantee IntelliSense in VS Code. | — |
| `.github` | CI/CD workflow, calling reusable workflows from [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp). | — |
| `.idp` | Helm chart depending on the shared `web` chart published from [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp). Copied by the deploy job into the `open-idp-apps` catalog repo, which ArgoCD actually watches. | — |
| `scripts` | Local test script (`go vet` + `go test`), mirroring the CI test job. | — |
| `Dockerfile` | Production workload image — used by the build job in the CI/CD pipeline. | — |
| `main.go` | Main application file containing the code. | — |

## CI/CD

Triggered on push ([.github/workflows/cicd.yaml](.github/workflows/cicd.yaml)):

1. **test** — runs `scripts/test.sh` (`go vet` + `go test`) inside `golang:1.27`.
2. **build** (on `main` only) — reusable [`build.yaml`](https://github.com/jasondavindev/open-idp/blob/main/.github/workflows/build.yaml) from [jasondavindev/open-idp](https://github.com/jasondavindev/open-idp) builds and pushes `<registry>/safplatform:<commit-sha>`.
3. **deploy** — reusable [`deploy.yaml`](https://github.com/jasondavindev/open-idp/blob/main/.github/workflows/deploy.yaml) from the same repo checks out the centralized [jasondavindev/open-idp-apps](https://github.com/jasondavindev/open-idp-apps) repo, copies `.idp/*` into `apps/safplatform/`, bumps `global.image.tag` to that SHA, and pushes the commit there; ArgoCD watches `open-idp-apps` and rolls out a new version of the Deployment in Kubernetes.

Required secrets: `CONTAINER_REGISTRY`, `REPO_USER`, `REPO_PASSWORD` (build), `PAT_WRITE_TOKEN` (deploy — write access to `open-idp-apps`).

## Screenshots

ArgoCD syncing the tag committed by the deploy job and rolling out the new ReplicaSet:

![ArgoCD application synced and rolling out](docs/images/argocd.png)

Request to the app running in the cluster, behind Traefik:

![curl request to the app returning the echoed JSON](docs/images/app.png)
