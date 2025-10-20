# Random

[![Test](https://github.com/johnwmail/random/workflows/Test/badge.svg)](https://github.com/johnwmail/random/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/johnwmail/random)](https://goreportcard.com/report/github.com/johnwmail/random)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](go.mod)

A tiny Go web service that generates secure random strings via HTML and JSON endpoints. The service runs locally, in Docker, or on AWS Lambda via the aws-lambda-go-api-proxy adapters.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Quick Start](#quick-start)
- [Deployment](#deployment)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Development](#development)
- [Build Metadata](#build-metadata)
- [Links](#links)

<a id="overview"></a>
## Overview

`random` ships a minimal UI at `/` and a JSON API at `/json`. Both endpoints generate two strings on every request:

- **Printable String**: Alphanumeric with 1–3 substitutions from `!#$%*+-=?@^_`
- **Alphanumeric String**: Letters and digits only

Query parameters let callers control the length of each string while the server clamps values to the safe range of 1–99 characters.

<a id="features"></a>
## ✨ Features

- 🚀 **Live Web UI** – Interactive page updates strings instantly as you tweak lengths
- 🎯 **JSON API** – Simple `GET /json` endpoint for programmatic clients
- 📏 **Length Clamping** – Prevents invalid values and enforces 1–99 character range
- 🔄 **Cache Busting** – Build metadata injected into static assets for fresh browser loads
- ☁️ **Lambda Ready** – Auto-detects `AWS_LAMBDA_FUNCTION_NAME` and runs behind API Gateway with zero code changes
- 🔬 **Tested** – Unit tests cover string generation, HTML rendering, and Lambda adapters

<a id="quick-start"></a>
## 🚀 Quick Start

### Run with Go

Requires Go 1.24+.

```bash
git clone https://github.com/johnwmail/random.git
cd random
go run ./...
# or build a binary
go build -o random .
./random
```

Visit http://localhost:8080 for the UI or call http://localhost:8080/json?p=20&a=25 for JSON.

### Quick API Examples

```bash
# Printable=33, Alphanumeric=22
curl -fsS "http://localhost:8080/json?p=33&a=22"

# Open the UI (macOS)
open http://localhost:8080/
```

<a id="deployment"></a>
## ☁️ Deployment

### Docker

```bash
docker build -t random:local -f docker/Dockerfile .
docker run --rm -p 8080:8080 --name random random:local
```

Using Compose:

```bash
docker compose -f docker/docker-compose.yml up --build
```

### AWS Lambda

The app switches to Lambda mode when `AWS_LAMBDA_FUNCTION_NAME` is present. The `deploy-lambda.yml` workflow builds a `bootstrap` binary and deploys it via the custom runtime.

Key environment values for deployment:

| Secret/Var | Purpose |
|-----------|---------|
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | Authentication for CI deployments |
| `AWS_REGION` | Target region |
| `LAMBDA_FUNCTION_NAME` | Lambda function to update |

The workflow produces a zip archive in `lambda-artifacts/` and publishes a new version.

<a id="configuration"></a>
## ⚙️ Configuration

The service intentionally keeps configuration surface area small. Important knobs:

| Option | Description |
|--------|-------------|
| Query `p` | Printable string length (default random 12–30) |
| Query `a` | Alphanumeric string length (default random 12–30) |
| Env `AWS_LAMBDA_FUNCTION_NAME` | Enables Lambda adapter mode |

Values outside 1–99 are clamped automatically.

<a id="api-endpoints"></a>
## 📋 API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | HTML UI with live length controls |
| GET | `/json` | JSON payload describing both strings |

Note on CLI clients
-------------------

By default the service will return JSON to programmatic or CLI clients (for example `curl`, `wget`, `powershell`, `httpie`, language HTTP libraries, etc.). This is detected using the `User-Agent` header. If you need to force HTML from a CLI client, request the `/` endpoint with an explicit `Accept: text/html` header or use a browser; to force JSON use `/json` or `Accept: application/json`.

Sample JSON response:

```json
{
  "printable": {
    "length": 33,
    "string": "P7d*93g1..."
  },
  "alphanumeric": {
    "length": 22,
    "string": "dN7Z0tXy4Kq1..."
  }
}
```

<a id="development"></a>
## 🔧 Development

```bash
# Format, lint, and test
go fmt ./...
go vet ./...
golangci-lint run
go test ./...

# Run the service locally during development
go run ./...
```

The CI pipeline in `.github/workflows/test.yml` enforces the formatting, linting, and testing steps above.

<a id="build-metadata"></a>
## 🏷️ Build Metadata

`main.go` exposes three variables injected at build time:

| Variable | Default | Purpose |
|----------|---------|---------|
| `Version` | `dev` | Semantic version or git tag |
| `BuildTime` | `unknown` | Build timestamp (ISO 8601 recommended) |
| `CommitHash` | `none` | Git commit SHA |

Inject values with Go build flags:

```bash
go build \
  -ldflags "-X main.Version=v1.2.3 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.CommitHash=$(git rev-parse --short HEAD)" \
  -o random .
```

The values are printed on startup and wired into the HTML template for cache busting of static assets.

<a id="links"></a>
## 🔗 Links

- **GitHub**: https://github.com/johnwmail/random
- **Issues**: https://github.com/johnwmail/random/issues
- **Actions**: https://github.com/johnwmail/random/actions

---

⭐ Star the project if this random string generator helps you out!
