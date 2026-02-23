FROM --platform=${BUILDPLATFORM} golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
# cache dependencies
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -a -ldflags "-s -w" -o manager cmd/manager/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /build/manager /app/manager

ENTRYPOINT ["/app/manager"]
