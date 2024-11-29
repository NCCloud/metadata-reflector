FROM golang:1.23 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w" -o manager cmd/manager/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /home/metadata-reflector

COPY --from=builder /workspace/manager .

ENTRYPOINT ["/home/metadata-reflector/manager"]
