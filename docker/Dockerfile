# 'builder' extends the fenv base image with FQL build & testing dependencies.
# This stage is used by fenv for local development and CI/CD verification tasks.
ARG FENV_DOCKER_TAG
FROM fenv:${FENV_DOCKER_TAG} AS builder

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


# 'gobuild' stage compiles the FQL binary.
FROM builder AS gobuild

COPY . /src
WORKDIR /src

ARG FQL_VER
RUN go build -o /fql -ldflags="-X 'github.com/janderland/fql/internal/app.Version=${FQL_VER}'"


# The final stage builds the minimal 'fql' runtime image.
FROM debian:12

RUN apt-get update && \
    apt-get install --no-install-recommends -y curl=7.88.* && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ARG FENV_FDB_VER
RUN curl -fsSL "https://github.com/apple/foundationdb/releases/download/${FENV_FDB_VER}/foundationdb-clients_${FENV_FDB_VER}-1_amd64.deb" -o /fdb.deb && \
    dpkg -i /fdb.deb && \
    rm /fdb.deb

ENV TERM="xterm-256color"

COPY --from=gobuild /fql /fql
COPY ./docker/shim.sh /shim.sh
ENTRYPOINT ["/shim.sh"]
