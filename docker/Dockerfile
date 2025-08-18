FROM --platform=$BUILDPLATFORM golang:1.24 AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /src
COPY . ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -tags=netgo -trimpath \
       -ldflags="-w -s -X 'main.buildTime=$(date --utc +%Y-%m-%dT%H:%M:%SZ)' \
       -X 'main.commitHash=$(git rev-parse HEAD)' \
       -X 'main.version=${VERSION:-latest}'" \
       -v -o /random .

FROM scratch

COPY --from=build ./random /
WORKDIR /
USER 8080:8080
EXPOSE 8080
ENTRYPOINT ["/random"]
