# FDBQ

FDBQ is a protype
[layer](https://apple.github.io/foundationdb/layer-concept.html)
for FoundationDB. Some of the things this project aims to
acheive are:
- [x] Provide a textual description of key-value schemas.
- [x] Provide an intuitive query language for FDB.
- [x] Provide a Go API which is structurally equivalent to
  the query language.
- [ ] Improve the ergonomics of the FoundationDB API.
  - [x] Merge the directory & tuple layers with the core FDB
    API.
  - [ ] Standardize the encoding of primitives (int, float,
    bool) as an FDB value.
  - [ ] Gracefully handle multi-transaction range-reads.
  - [ ] Gracefully handle transient errors.

Here is the [syntax definiton](syntax.ebnf) for the query
language.
