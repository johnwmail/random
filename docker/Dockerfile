FROM --platform=$BUILDPLATFORM golang:1.24 AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /src
COPY . ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -tags=netgo -trimpath \
       -ldflags="-w -s -X 'main.BuildTime=$(date --utc +%Y-%m-%dT%H:%M:%SZ)' \
       -X 'main.CommitHash=$(git rev-parse HEAD)' \
       -X 'main.Version=${VERSION:-latest}'" \
       -v -o /random .

FROM scratch

# copy static assets from build stage (they live at /src/static inside the build stage)
COPY --from=build /src/static /static
# copy the compiled binary placed at /random in the build stage
COPY --from=build /random /random
WORKDIR /
USER 8080:8080
EXPOSE 8080
ENTRYPOINT ["/random"]
