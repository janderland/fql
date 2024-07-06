# 'builder' includes all build & testing dependencies.
FROM debian:12 AS builder

RUN apt-get update &&\
    apt-get install -y \
      build-essential \
      shellcheck=0.9.0-1 \
      git=1:2.39.2-1.1 \
      curl=7.88.1-10+deb12u6 &&\
    apt-get clean

ARG FDB_LIB_URL
RUN curl -Lo /fdb.deb $FDB_LIB_URL &&\
    dpkg -i /fdb.deb

ARG GO_URL
RUN curl -Lo go.tar.gz $GO_URL &&\
    tar -C /usr/local -xzf ./go.tar.gz &&\
    rm ./go.tar.gz
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"
ENV GOCACHE="/cache/gocache"
ENV GOMODCACHE="/cache/gomod"

ARG GOLANGCI_LINT_URL="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
ARG GOLANGCI_LINT_VER
RUN curl -Lo golint.sh $GOLANGCI_LINT_URL &&\
    sh ./golint.sh -b "$(go env GOPATH)/bin" $GOLANGCI_LINT_VER &&\
    rm ./golint.sh
ENV GOLANGCI_LINT_CACHE="/cache/golangci-lint"

ARG HADOLINT_URL
RUN curl -Lo /usr/local/bin/hadolint $HADOLINT_URL &&\
    chmod +x /usr/local/bin/hadolint

ARG JQ_URL
RUN curl -Lo /usr/local/bin/jq $JQ_URL &&\
    chmod +x /usr/local/bin/jq

# Configure git so it allows any user to run git commands
# on the /fdbq directory. This allows the user which runs
# CI to be different from the user which built the Docker
# image.
RUN git config --global --add safe.directory /fdbq


# 'gobuild' executes 'go build'.
FROM --platform=linux/amd64 builder AS gobuild

COPY . /src
WORKDIR /src

ARG FDBQ_VER
RUN go build -o /fdbq -ldflags="-X 'github.com/janderland/fdbq/internal/app.Version=${FDBQ_VER}'"


# The final stage builds the 'fdbq' image.
FROM debian:12

COPY --from=gobuild /fdb.deb /fdb.deb
RUN dpkg -i ./fdb.deb &&\
    rm /fdb.deb

ENV TERM="xterm-256color"

COPY --from=gobuild /fdbq /fdbq
COPY ./docker/shim.sh /shim.sh
ENTRYPOINT ["/shim.sh"]
