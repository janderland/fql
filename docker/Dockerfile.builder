# Builder image extends the fenv base image with FQL build & testing dependencies.
# This image is used by fenv for local development and CI/CD verification tasks.
ARG FENV_DOCKER_TAG
FROM fenv:${FENV_DOCKER_TAG}

# Install Go
ARG GO_URL="https://go.dev/dl/go1.19.1.linux-amd64.tar.gz"
RUN curl -fsSL ${GO_URL} | tar -C /usr/local -xz
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"
ENV GOCACHE="/cache/gocache"
ENV GOMODCACHE="/cache/gomod"

# Install golangci-lint
ARG GOLANGCI_LINT_URL="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
ARG GOLANGCI_LINT_VER="v1.49.0"
RUN curl -fsSL ${GOLANGCI_LINT_URL} | sh -s -- -b "$(go env GOPATH)/bin" ${GOLANGCI_LINT_VER}
ENV GOLANGCI_LINT_CACHE="/cache/golangci-lint"

# Install pandoc
ARG PANDOC_URL="https://github.com/jgm/pandoc/releases/download/3.3/pandoc-3.3-1-amd64.deb"
RUN curl -fsSL ${PANDOC_URL} -o /tmp/pandoc.deb && \
    dpkg -i /tmp/pandoc.deb && \
    rm /tmp/pandoc.deb
