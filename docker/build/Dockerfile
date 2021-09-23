FROM centos:7

RUN yum -y install epel-release gcc-c++ ShellCheck &&\
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
RUN curl -Lf $GOLANGCI_LINT_URL | sh -s -- -b $(go env GOPATH)/bin v1.42.1

