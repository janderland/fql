FROM centos:7

RUN yum -y --setopt=skip_missing_names_on_install=False install epel-release gcc-4.8.5 &&\
    yum -y --setopt=skip_missing_names_on_install=False install ShellCheck-0.3.8 &&\
    yum -y --setopt=skip_missing_names_on_install=False install git-1.8.3.1 &&\
    yum -y clean all

ARG FDB_LIB_URL
RUN curl -Lo fdb.rpm $FDB_LIB_URL &&\
    rpm -i ./fdb.rpm &&\
    rm ./fdb.rpm

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
