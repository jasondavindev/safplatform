FROM golang:1.27

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        openssh-client \
        make \
        unzip \
        jq \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV GOPATH=/go \
    GOROOT=/usr/local/go \
    PATH=/usr/local/go/bin:/go/bin:$PATH

RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install golang.org/x/tools/cmd/goimports@latest

RUN mkdir -p /go/pkg/mod /root/.cache/go-build
