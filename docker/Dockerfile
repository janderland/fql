# 'fenv-builder' extends the fenv base image with FQL build & testing dependencies.
# This stage is used by fenv for local development and CI/CD verification tasks.
ARG FENV_DOCKER_TAG
FROM fenv:${FENV_DOCKER_TAG} AS fenv-builder

# Install additional system packages needed for build
RUN apt-get update && \
    apt-get install --no-install-recommends -y \
      build-essential=12.9 \
      ca-certificates=20* \
      git=1:2.39.* \
      curl=7.88.* && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

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

# Install shellcheck
ARG SHELLCHECK_URL="https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.linux.x86_64.tar.xz"
RUN curl -fsSL ${SHELLCHECK_URL} -o /tmp/shellcheck.tar.xz && \
    tar -xf /tmp/shellcheck.tar.xz -C /tmp && \
    mv /tmp/shellcheck-*/shellcheck /usr/local/bin && \
    rm -rf /tmp/shellcheck*

# Install hadolint
ARG HADOLINT_URL="https://github.com/hadolint/hadolint/releases/download/v2.7.0/hadolint-Linux-x86_64"
RUN curl -fsSL ${HADOLINT_URL} -o /usr/local/bin/hadolint && \
    chmod +x /usr/local/bin/hadolint

# Install jq
ARG JQ_URL="https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"
RUN curl -fsSL ${JQ_URL} -o /usr/local/bin/jq && \
    chmod +x /usr/local/bin/jq

# Install pandoc
ARG PANDOC_URL="https://github.com/jgm/pandoc/releases/download/3.3/pandoc-3.3-1-amd64.deb"
RUN curl -fsSL ${PANDOC_URL} -o /tmp/pandoc.deb && \
    dpkg -i /tmp/pandoc.deb && \
    rm /tmp/pandoc.deb

# Configure git to allow any user to run git commands on /src
# This allows the user which runs CI to be different from the
# user which built the Docker image.
RUN git config --global --add safe.directory /src


# 'gobuild' stage compiles the FQL binary.
FROM fenv-builder AS gobuild

COPY . /src
WORKDIR /src

ARG FQL_VER
RUN go build -o /fql -ldflags="-X 'github.com/janderland/fql/internal/app.Version=${FQL_VER}'"


# 'fdb-installer' stage downloads the FDB client .deb for the final stage.
FROM debian:12 AS fdb-installer

RUN apt-get update && \
    apt-get install --no-install-recommends -y curl=7.88.* && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ARG FDB_LIB_URL
RUN curl -fsSL ${FDB_LIB_URL} -o /fdb.deb


# The final stage builds the minimal 'fql' runtime image.
FROM debian:12

COPY --from=fdb-installer /fdb.deb /fdb.deb
RUN dpkg -i ./fdb.deb && \
    rm /fdb.deb

ENV TERM="xterm-256color"

COPY --from=gobuild /fql /fql
COPY ./docker/shim.sh /shim.sh
ENTRYPOINT ["/shim.sh"]
