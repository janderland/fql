FROM centos:7

ARG FDB_LIB_URL="https://www.foundationdb.org/downloads/6.2.30/rhel7/installers/foundationdb-clients-6.2.30-1.el7.x86_64.rpm"
RUN curl -Lo fdb.rpm $FDB_LIB_URL &&\
    rpm -i ./fdb.rpm              &&\
    rm ./fdb.rpm

COPY fdbq /fdbq
COPY docker/fdbq/shim.sh /shim.sh
ENTRYPOINT ["/shim.sh"]
