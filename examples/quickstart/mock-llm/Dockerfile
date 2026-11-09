FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY main.go .
RUN go mod init mock-llm && CGO_ENABLED=0 go build -o /out/mock-llm .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=builder /out/mock-llm /mock-llm
EXPOSE 8000

# The distroless runtime has no shell or wget, so HEALTHCHECK execs the
# binary itself, which probes /healthz over loopback.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD ["/mock-llm", "healthcheck"]

# Run as the distroless non-root user (UID/GID 65532); the explicit USER
# statement makes the intent auditable.
USER nonroot:nonroot

ENTRYPOINT ["/mock-llm"]
