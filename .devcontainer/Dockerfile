# Imagem oficial do Go: fornece /usr/local/go/bin/go usado pela extensao Go do VS Code
FROM golang:1.27

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Usuario nao-root (a imagem golang so traz root)
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m -s /bin/bash $USERNAME \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        openssh-client \
        make \
        sudo \
        unzip \
    && echo "$USERNAME ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV GOPATH=/go \
    GOROOT=/usr/local/go \
    PATH=/usr/local/go/bin:/go/bin:$PATH

# Ferramentas usadas pela extensao Go (language server, debugger, lint, format)
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && chmod -R a+rwX /go

RUN mkdir -p /go/pkg/mod /home/$USERNAME/.cache/go-build \
    && chown -R $USER_UID:$USER_GID /go /home/$USERNAME

USER $USERNAME
