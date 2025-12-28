# 'gobuild' stage compiles the FQL binary using the pre-built builder image from fenv.
# The builder image (fenv-ext-fql) is built separately via ./build.sh --image build
ARG FENV_EXT_DOCKER_TAG
FROM fenv-ext-fql:${FENV_EXT_DOCKER_TAG} AS gobuild

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
