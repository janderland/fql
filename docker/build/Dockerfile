FROM centos:7

RUN yum -y --setopt=skip_missing_names_on_install=False install epel-release gcc-4.8.5 &&\
    yum -y --setopt=skip_missing_names_on_install=False install ShellCheck-0.3.8 &&\
    yum -y --setopt=skip_missing_names_on_install=False install git-1.8.3.1 &&\
    yum -y clean all

ARG FDB_LIB_URL="https://github.com/apple/foundationdb/releases/download/6.2.30/foundationdb-clients-6.2.30-1.el7.x86_64.rpm"
RUN curl -Lo fdb.rpm $FDB_LIB_URL &&\
    rpm -i ./fdb.rpm &&\
    rm ./fdb.rpm

ARG GO_URL="https://go.dev/dl/go1.19.1.linux-amd64.tar.gz"
RUN curl -Lo go.tar.gz $GO_URL &&\
    tar -C /usr/local -xzf ./go.tar.gz &&\
    rm ./go.tar.gz
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"
ENV GOCACHE="/cache/gocache"
ENV GOMODCACHE="/cache/gomod"

ARG GOLANGCI_LINT_URL="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
RUN curl -Lo golint.sh $GOLANGCI_LINT_URL &&\
    sh ./golint.sh -b "$(go env GOPATH)/bin" v1.49.0 &&\
    rm ./golint.sh
ENV GOLANGCI_LINT_CACHE="/cache/golangci-lint"

ARG HADOLINT_URL="https://github.com/hadolint/hadolint/releases/download/v2.7.0/hadolint-Linux-x86_64"
RUN curl -Lo /usr/local/bin/hadolint $HADOLINT_URL &&\
    chmod +x /usr/local/bin/hadolint

ARG JQ_URL="https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"
RUN curl -Lo /usr/local/bin/jq $JQ_URL &&\
    chmod +x /usr/local/bin/jq
