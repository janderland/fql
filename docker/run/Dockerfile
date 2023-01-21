FROM janderland/fdbq-build:latest AS build

COPY . /src
RUN go build /src -o /fdbq

FROM build AS fdbq

ARG FDB_LIB_URL="https://github.com/apple/foundationdb/releases/download/6.2.30/foundationdb-clients-6.2.30-1.el7.x86_64.rpm"
RUN curl -Lo fdb.rpm $FDB_LIB_URL &&\
    rpm -i ./fdb.rpm              &&\
    rm ./fdb.rpm

COPY --from=build /fdbq /fdbq
COPY ./docker/run/shim.sh /shim.sh
ENTRYPOINT ["/shim.sh"]
