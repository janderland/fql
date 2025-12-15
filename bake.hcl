# Docker Bake configuration for FQL.
#
# Variables can be overridden via environment variables.
# The CI workflow passes these as env vars from the matrix.

variable "DOCKER_TAG" {
  default = "latest"
}

variable "FDB_VER" {
  default = "6.2.30"
}

variable "FDB_LIB_URL" {
  default = "https://github.com/apple/foundationdb/releases/download/6.2.30/foundationdb-clients_6.2.30-1_amd64.deb"
}

variable "GO_URL" {
  default = "https://go.dev/dl/go1.19.1.linux-amd64.tar.gz"
}

variable "GOLANGCI_LINT_VER" {
  default = "v1.49.0"
}

variable "SHELLCHECK_URL" {
  default = "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.linux.x86_64.tar.xz"
}

variable "HADOLINT_URL" {
  default = "https://github.com/hadolint/hadolint/releases/download/v2.7.0/hadolint-Linux-x86_64"
}

variable "JQ_URL" {
  default = "https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"
}

variable "PANDOC_URL" {
  default = "https://github.com/jgm/pandoc/releases/download/3.3/pandoc-3.3-1-amd64.deb"
}

# Shared build arguments used by both targets.
function "build_args" {
  params = []
  result = {
    FQL_VER           = DOCKER_TAG
    FDB_LIB_URL       = FDB_LIB_URL
    GO_URL            = GO_URL
    GOLANGCI_LINT_VER = GOLANGCI_LINT_VER
    SHELLCHECK_URL    = SHELLCHECK_URL
    HADOLINT_URL      = HADOLINT_URL
    JQ_URL            = JQ_URL
    PANDOC_URL        = PANDOC_URL
  }
}

group "default" {
  targets = ["build", "fql"]
}

target "build" {
  context    = "./docker"
  dockerfile = "Dockerfile"
  target     = "builder"
  tags       = ["docker.io/janderland/fql-build:${DOCKER_TAG}"]
  platforms  = ["linux/amd64"]
  args       = build_args()
  cache-from = ["type=gha,scope=build-${FDB_VER}"]
  cache-to   = ["type=gha,mode=max,scope=build-${FDB_VER}"]
}

target "fql" {
  context    = "."
  dockerfile = "./docker/Dockerfile"
  tags       = ["docker.io/janderland/fql:${DOCKER_TAG}"]
  platforms  = ["linux/amd64"]
  args       = build_args()
  cache-from = ["type=gha,scope=fql-${FDB_VER}"]
  cache-to   = ["type=gha,mode=max,scope=fql-${FDB_VER}"]
}
