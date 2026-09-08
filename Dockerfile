FROM golang:1.27 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w" \
    -o /out/server .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
