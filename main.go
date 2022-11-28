/*
fdbq is a query language for Foundation DB

Usage:

	fdbq [flags] query ...

Flags:

	-b, --bytes            print full byte strings instead of just their length
	-c, --cluster string   path to cluster file
	-h, --help             help for fdbq
	    --limit int        limit the number of KVs read in range-reads
	-l, --little           encode/decode values as little endian instead of big endian
	    --log              perform debug logging
	-r, --reverse          query range-reads in reverse order
	-s, --strict           throw an error if a KV is read which doesn't match the schema
	-w, --write            allow write queries
*/
package main

import "github.com/janderland/fdbq/internal/app"

func main() {
	if err := app.Fdbq.Execute(); err != nil {
		panic(err)
	}
}
