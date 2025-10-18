random
=====

A tiny Go web service that returns random strings via HTML and JSON. Runs locally, in Docker, and on AWS Lambda (API Gateway) using Gin with aws-lambda-go-api-proxy.

Features
--------
- Endpoints:
  - GET `/` — HTML page that shows two random strings and updates as you change lengths.
  - GET `/json` — JSON payload with two random strings.
- Query params on both endpoints:
  - `p` — printable string length (1–99)
  - `a` — alphanumeric string length (1–99)
- Defaults: each length randomly chosen between 12 and 30 if not provided.
- Version info embedded at build time and printed on start (version, buildTime, commitHash).

API
---
- JSON endpoint: `GET /json?p=33&a=22`
  - Response shape:
    {
      "printable": { "length": 33, "string": "..." },
      "alphanumeric": { "length": 22, "string": "..." }
    }
  - Notes:
    - Printable strings are alphanumeric with a few random substitutions from: `!#$%*+-=?@^_`.
    - Lengths are clamped to [1, 99].

Examples
--------
- JSON (quote the URL if your shell might treat & specially):
  curl -fsS "http://localhost:8080/json?p=33&a=22"
- HTML:
  open http://localhost:8080/

Run locally (Go)
----------------
Requirements: Go 1.24+
- go run ./...
- or build: go build -o random ./... && ./random
The service listens on port 8080.

Docker
------
Build and run locally using the provided multi-stage Dockerfile:
- docker build -t random:local -f docker/Dockerfile .
- docker run --rm -p 8080:8080 --name random random:local

docker-compose example (provided):
- docker compose -f docker/docker-compose.yml up --build

Pull from GHCR
---------------
Images are published on tags matching v* via GitHub Actions.
- docker login ghcr.io -u YOUR_GH_USER -p YOUR_GITHUB_TOKEN
- docker pull ghcr.io/OWNER/REPO:latest
- docker run --rm -p 8080:8080 ghcr.io/OWNER/REPO:latest

AWS Lambda deployment
---------------------
Runs on AWS Lambda behind API Gateway. The app auto-detects Lambda via AWS_LAMBDA_FUNCTION_NAME and uses the Lambda adapter.

Workflow: .github/workflows/deploy-lambda.yml uses aws-actions/aws-lambda-deploy@v1.

Required repo secrets (or configure OIDC role):
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_REGION
- LAMBDA_FUNCTION_NAME (target Lambda function name)

Behavior:
- Builds a Linux bootstrap binary for the custom runtime.
- Prepares lambda-artifacts/ and deploys with handler=bootstrap, runtime=provided.al2, publish=true.
- Trigger: push to branch deploy/lambda.

CI
--
Workflows under .github/workflows/:
- test.yml (push/PR to main): go fmt, go vet, golangci-lint, go test
- build.yml (tags v*): multi-arch build and push to GHCR; runtime verification by starting the image and checking / and /json

Development
-----------
- go fmt ./...
- go vet ./...
- go test -v ./...

Notes
-----
- Port: 8080
- Length limits: 1–99
- HTML page updates strings as you change inputs
- Quote URLs with & in query strings in shells
