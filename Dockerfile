# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/smyth ./cmd/smyth

FROM gcr.io/distroless/static-debian12

ARG REPOSITORY_URL

LABEL org.opencontainers.image.source=$REPOSITORY_URL

COPY --from=builder /out/smyth /usr/local/bin/smyth

ENTRYPOINT ["/usr/local/bin/smyth"]
CMD ["--help"]
