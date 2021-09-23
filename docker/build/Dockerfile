FROM centos:7

RUN yum -y --setopt=skip_missing_names_on_install=False install epel-release gcc-4.8.5 &&\
    yum -y --setopt=skip_missing_names_on_install=False install ShellCheck-0.3.8 &&\
    yum -y clean all

ARG FDB_LIB_URL="https://www.foundationdb.org/downloads/6.3.15/rhel7/installers/foundationdb-clients-6.3.15-1.el7.x86_64.rpm"
RUN curl -Lo fdb.rpm $FDB_LIB_URL &&\
    rpm -i ./fdb.rpm &&\
    rm ./fdb.rpm

ARG GO_URL="https://golang.org/dl/go1.16.8.linux-amd64.tar.gz"
RUN curl -Lo go.tar.gz $GO_URL &&\
    tar -C /usr/local -xzf ./go.tar.gz &&\
    rm ./go.tar.gz
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"

ARG GOLANGCI_LINT_URL="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
RUN curl -Lo golint.sh $GOLANGCI_LINT_URL &&\
    sh ./golint.sh -b "$(go env GOPATH)/bin" v1.42.1 &&\
    rm ./golint.sh

ARG HADOLINT_URL="https://github.com/hadolint/hadolint/releases/download/v2.7.0/hadolint-Linux-x86_64"
RUN curl -Lo /usr/local/bin/hadolint $HADOLINT_URL &&\
    chmod +x /usr/local/bin/hadolint

ARG JQ_URL="https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"
RUN curl -Lo /usr/local/bin/jq $JQ_URL &&\
    chmod +x /usr/local/bin/jq

