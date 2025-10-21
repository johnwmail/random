---# Copilot Instructions

description: 'Copilot instructions for the random web service project'

applyTo: '**/*.go,**/*.html,**/*.js,**/*.css,**/*.yml,**/Dockerfile,**/docker-compose.yml'Random is a Go-based HTTP clipboard/pastebin service built with Gin that supports dual deployment modes: **server mode** (filesystem storage) and **Lambda mode** (S3 storage).

---

## Essential Architecture Knowledge

# Copilot Instructions for Random Web Service

### Deployment Mode Detection

This file provides guidance for GitHub Copilot when contributing to the **random** project—a tiny Go web service that generates and returns random strings via HTML and JSON endpoints.The codebase automatically detects deployment mode via `isLambdaEnvironment()` checking for `AWS_LAMBDA_FUNCTION_NAME` env var. This drives storage backend selection:

- **Server mode**: Uses `storage.NewFileSystemStore("./data")` 

## Project Overview- **Lambda mode**: Uses `storage.NewS3Store(cfg.S3Bucket, cfg.S3Prefix)`



**Purpose:** Provide a lightweight HTTP service that generates random printable and alphanumeric strings, with support for customizable string lengths via query parameters.### Storage Abstraction Pattern

All storage operations go through the `PasteStore` interface:

**Key Technologies:**```go

- **Language:** Go 1.24+type PasteStore interface {

- **Web Framework:** Gin    Store(paste *models.Paste) error

- **Deployment:** Local, Docker, AWS Lambda (with API Gateway)    Get(id string) (*models.Paste, error) 

- **CI/CD:** GitHub Actions    Exists(id string) (bool, error)

    Delete(id string) error

**Core Endpoints:**    IncrementReadCount(id string) error

- `GET /` — HTML page with interactive UI to generate random strings with custom lengths    StoreContent(id string, content []byte) error

- `GET /json` — JSON API endpoint returning structured random string data    GetContent(id string) ([]byte, error)

}

**Query Parameters (both endpoints):**```

- `p` — Printable string length (1–99, default: random 12–30)

- `a` — Alphanumeric string length (1–99, default: random 12–30)Key insight: Metadata and content are stored separately. Both filesystem and S3 implementations store:

- Content as raw bytes (`{slug}` file/object)

## Code Organization- Metadata as JSON (`{slug}.json` file/object)



```### Handler Architecture 

random/Handlers are organized by function, not REST resources:

├── main.go              # Entry point, route handlers, Lambda adapter- `handlers/paste.go` - Core upload/retrieval logic

├── main_test.go         # Unit tests for string generation functions- `handlers/webui.go` - HTML UI endpoints  

├── go.mod, go.sum       # Go module dependencies- `handlers/meta.go` - Metadata endpoints

├── README.md            # User-facing documentation- `handlers/system.go` - Health/status endpoints

├── static/- `handlers/upload/` and `handlers/retrieval/` - Specialized logic

│   └── index.html       # HTML template for web UI

├── docker/## Critical Implementation Patterns

│   ├── Dockerfile       # Multi-stage build for local/GHCR deployment

│   └── docker-compose.yml### Slug Generation with Collision Handling

├── infra/               # Infrastructure-as-code for LambdaUses batch generation with fallback to longer slugs in `generateUniqueSlug()`:

└── .github/```go

    ├── instructions/    # Development guidelineslengths := []int{5, 6, 7}  // Try incrementally longer slugs

    └── workflows/       # CI/CD pipelines (test.yml, build.yml, deploy-lambda.yml)candidates, err := utils.GenerateSlugBatch(batchSize, length)

``````

Also checks if existing slugs are expired before considering them collisions.

## Development Guidelines

### Required Before Each Commit
- Run `go fmt ./...` and `golangci-lint run` before committing any changes to ensure proper code formatting and linting
- This will run gofmt on all Go files to maintain consistent style

### Development Flow
- Code changes, add new feature and bug fixes
  ** If adding new features, add corresponding unit tests, integration tests, and update documentation (Documents/* and README.md as needed)
- Ensure no linting errors with `golangci-lint run`, `go fmt ./...`, and `go vet ./...`
- Run unit tests with `go test ./...` to verify functionality
- Run integration tests with `bash scripts/integration-test.sh` (requires stop and re-running server)
- Address any issues found during testing and repeat until all tests pass
- Before pushing changes, ensure all above steps is passed and code is clean


### Content-Type Detection Chain

### Go Code Standards1. Use client-provided `Content-Type` header if present

2. Detect from filename extension (if filename provided)  

Follow the instructions in `.github/instructions/go.instructions.md` plus these project-specific conventions:3. Fall back to content-based detection via `utils/mime.go`



1. **String Generation Functions**### Buffer Size Limit Implementation

   - `GenerateRandomPrintable(length int) string` — Alphanumeric base with 1–3 random special character substitutions**CRITICAL**: Buffer size limits must be properly enforced to prevent silent truncation. The correct pattern handles multipart vs direct uploads differently:

   - `GenerateRandomAlphanumeric(length int) string` — Pure alphanumeric (a-z, A-Z, 0-9)

   - Both use the local `rnd *rand.Rand` instance (initialized in `init()`)```go

   - Always clamp lengths to [1, 99] range in handlers before generation// Check if this is a multipart upload

isMultipart := strings.HasPrefix(c.ContentType(), "multipart/form-data")

2. **Random Number Generation**

   - Use `rnd *rand.Rand` (initialized once in `init()` using `time.Now().UnixNano()`)// For direct POST requests, check Content-Length early (accurate for content size)

   - Do **not** use the deprecated global `math/rand` functions// Skip for multipart since Content-Length includes boundaries and headers

   - Ensure thread safety by keeping `rnd` as a package-level variableif !isMultipart {

    if contentLength := c.Request.ContentLength; contentLength > 0 && contentLength > bufferSize {

3. **Handler Functions**        return error

   - Both `/` and `/json` routes use the shared `generateStrings` handler with conditional response format    }

   - Query parameters: parse with `c.GetQuery()`, convert to int with `strconv.Atoi()`}

   - Clamp and validate lengths in the handler before generation

   - Set `Cache-Control: no-store, no-cache, must-revalidate` on `/json` to ensure fresh values// Read with io.LimitReader

content, err := io.ReadAll(io.LimitReader(reader, bufferSize))

4. **Error Handling**

   - Use `log.Fatal()` for startup errors (missing templates, Lambda initialization failures)// Always check for truncation by attempting to read one more byte

   - Return HTTP 500 for recoverable errors (template parsing)var oneByte [1]byte

   - Avoid panic() except during initializationn, _ := reader.Read(oneByte[:])

if n > 0 {

5. **Logging**    return error // Content was truncated

   - Log version info at startup (version, BuildTime, CommitHash)}

   - Log Lambda environment detection```

   - Use structured logging where applicable

This prevents silent truncation while being accurate for both direct POST and multipart file uploads.

### Testing

### Platform-Specific Buffer Size Limits ⚠️

- Write unit tests in `main_test.go` for string generation functions

- Test length clamping (edge cases: 0, 1, 99, 100)- **Lambda deployments**: Refer to `Documents/LAMBDA.md` for AWS Lambda 6MB payload limits

- Test special character substitution in printable strings- **API Gateway/CloudFront**: Check AWS service limits for your architecture

- Use `testing.T` for assertions and error reporting

- Aim for >80% code coverage for core string generation### Error Response Consistency 

The codebase has specific requirements for 404 handling:

Example test structure:- **CLI clients** (curl/wget): Return JSON `{"error": "message"}` 

```go- **Browser clients**: Render HTML error page using `view.html` template

func TestGenerateRandomAlphanumeric(t *testing.T) {- Detection via User-Agent string analysis

    // Test various lengths

    for _, len := range []int{1, 10, 99} {## Testing & Development Workflows

        result := GenerateRandomAlphanumeric(len)

        if len(result) != len {### Required Before Each Commit

            t.Errorf("Expected length %d, got %d", len, len(result))- Run `go fmt ./...` and `golangci-lint run` before committing any changes to ensure proper code formatting and linting

        }- This will run gofmt on all Go files to maintain consistent style

    }

}### Development Flow

```- Code changes, add new feature and bug fixes

  ** If adding new features, add corresponding unit tests, integration tests, and update documentation (Documents/* and README.md as needed)

### HTML/Template Guidelines- Ensure no linting errors with `golangci-lint run`, `go fmt ./...`, and `go vet ./...`

- Run unit tests with `go test ./...` to verify functionality

- Embed template data as a map in handlers- Address any issues found during testing and repeat until all tests pass

- Support dynamic length input fields on the HTML page- Before pushing changes, ensure all above steps is passed and code is clean

- Use client-side JavaScript to fetch fresh strings without page reload

- Ensure accessibility: use semantic HTML, proper labels, ARIA attributes where needed### Test Cleanup Requirements ⚠️

**CRITICAL**: All tests must clean up artifacts they create. The unified integration test script (`scripts/integration-test.sh`) uses:

### Docker and Deployment

- Cleanup function removes only recorded slugs or recently modified files (`-mmin -60`)

1. **Dockerfile**- Never use broad cleanup like `rm -rf ./data/*` - it may delete unrelated data

   - Use multi-stage builds (builder + runtime)- Cleanup runs automatically via EXIT trap handlers in lib.sh

   - Builder: golang:latest, compile with `-ldflags` for version info- Unit tests use `defer cleanupTestData(store.dataDir)` to remove test files

   - Runtime: Alpine Linux (smallest footprint)

   - Expose port 8080This requirement ensures reproducible, reliable tests and prevents leftover artifacts from affecting subsequent runs or deployments.

   - Health check: `curl http://localhost:8080/` or similar

### Buffer Size Testing

2. **Lambda Deployment**Buffer size limits are tested at multiple levels:

   - Auto-detect Lambda via `AWS_LAMBDA_FUNCTION_NAME`- **Unit tests**: `TestBufferSizeLimit` tests both direct POST and multipart uploads

   - Build `bootstrap` binary (custom runtime, `provided.al2`)- **Integration tests**: `test_buffer.sh` module in unified test suite tests end-to-end with real HTTP requests

   - Use `ginadapter` v1 or v2 based on event type- Both test that oversized uploads are rejected with 400 status and appropriate error messages

   - Publish new versions on deploy

### CI/CD Workflows- Integration tests assert unauthenticated uploads return 401 when auth is enabled and verify authenticated uploads succeed using the generated test key.



1. **test.yml** (on push/PR to main)### Environment Variables

   - Run `go fmt`, `go vet`, `golangci-lint run`, `go test`

2. **build.yml** (on tags v*)

   - Multi-arch build (linux/amd64, linux/arm64, etc.)- `DEBUG` - Enables verbose logging including all environment variables

   - Push to GHCR with tag

   - Run smoke test (start image, check endpoints return 200)## Deployment Specifics



3. **deploy-lambda.yml** (on push to deploy/lambda)### Docker/Kubernetes

   - Build Linux bootstrap binary- Uses non-root user (1001:1001) 

   - Create lambda-artifacts/ with deployment package- Read-only container filesystem

   - Deploy to Lambda with AWS credentials- Health check via `/health` endpoint

- Static assets copied to `/app/static/`

## Common Tasks

### Lambda Integration

### Adding a New Endpoint- Uses `awslabs/aws-lambda-go-api-proxy` for Gin integration

- Gin routes remain identical between server and Lambda modes

1. Define handler function in `main.go`- Lambda handler wraps existing Gin router - no code duplication

2. Register route in `setupRouter()` (or inline in `main()`)

3. Add unit tests in `main_test.go`### Version Management

4. Update `README.md` with endpoint documentationBuild-time injection pattern:

5. Test locally: `go run ./...` and curl the endpoint```bash

6. Update HTML page if user-facinggo build -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.CommitHash=$GIT_COMMIT"

```

### Modifying String Generation Logic

This architecture enables true "write once, deploy anywhere" with the same codebase running in containers, servers, and serverless environments.

1. Edit `GenerateRandomPrintable()` or `GenerateRandomAlphanumeric()`

2. Update special character set if changing printable characters
3. Add/update tests in `main_test.go`
4. Verify length clamping logic
5. Run: `go test -v ./...`

### Updating Dependencies

1. Run `go get -u ./...` to check for updates
2. Run `go mod tidy` to clean up
3. Run full test suite: `go test ./...`
4. Test Docker build: `docker build -t random:test .`
5. Commit `go.mod` and `go.sum` together

### Local Testing

```bash
# Start the service
go run ./...

# Test JSON endpoint
curl "http://localhost:8080/json?p=20&a=15"

# Test HTML endpoint
curl -s http://localhost:8080/ | head -20

# Run tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Docker Testing

```bash
# Build locally
docker build -t random:local -f docker/Dockerfile .

# Run
docker run --rm -p 8080:8080 random:local

# Docker Compose
docker compose -f docker/docker-compose.yml up --build
```

## Key Files and Their Purpose

| File | Purpose |
|------|---------|
| `main.go` | Route handlers, string generation, Lambda adapter |
| `main_test.go` | Unit tests for generation functions |
| `static/index.html` | Web UI template |
| `docker/Dockerfile` | Multi-stage build for deployment |
| `docker/docker-compose.yml` | Local development with Docker |
| `infra/` | Terraform or CDK for Lambda infrastructure (if present) |
| `.github/workflows/test.yml` | CI: format, lint, test |
| `.github/workflows/build.yml` | Build and push Docker images |
| `.github/workflows/deploy-lambda.yml` | Deploy to Lambda |

## Performance and Constraints

- **String Length:** Clamped to [1, 99]
- **Port:** 8080 (configurable via environment if needed)
- **Lambda Limits:** 6MB payload limit (not a concern for small random strings)
- **Concurrency:** Gin handles concurrent requests safely; `rnd` is thread-local per goroutine
- **Caching:** Disabled on JSON endpoint to ensure fresh values per request

## Security Considerations

1. **Input Validation:** Always clamp and validate query parameter lengths
2. **Template Injection:** Use `html/template` (auto-escaping) not `text/template`
3. **Dependencies:** Keep minimal; audit regularly via `go list -json -m all`
4. **Lambda:** Use least-privilege IAM roles; no sensitive data in logs

## Debugging Tips

- **Lambda issues:** Check CloudWatch logs; use `AWS_LAMBDA_LOG_LEVEL=debug`
- **Local debugging:** Set `GIN_MODE=debug` for verbose Gin output
- **Random seed issues:** Verify `rnd` is initialized once in `init()`
- **Template errors:** Enable template debugging by checking error details before responding

## Best Practices Specific to Random

1. **Always use the project-level `rnd`**, not `math/rand` global functions
2. **Clamp lengths before generation**, not inside generation functions
3. **Keep special characters for printable strings minimal** (current: `!#$%*+-=?@^_`)
4. **Return structured JSON** with both length and string in response
5. **Cache-Control headers** prevent stale values in browser caches and CDNs
6. **Version info** should be embedded at build time, not computed at runtime

## References

- [Go Effective Go](https://go.dev/doc/effective_go)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [AWS Lambda Go](https://github.com/aws/aws-lambda-go)
- [aws-lambda-go-api-proxy](https://github.com/awslabs/aws-lambda-go-api-proxy)
- Project `README.md` for user documentation

---

**Last Updated:** October 2025  
**Maintainers:** See repository CODEOWNERS
