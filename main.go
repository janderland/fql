package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/janderland/fdbq/parser"
)

func main() {
	query, err := parser.ParseQuery(strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Println(errors.Wrap(err, "failed to parse query"))
		os.Exit(1)
	}
	str, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		fmt.Println(errors.Wrap(err, "failed to marshal JSON"))
		os.Exit(1)
	}
	fmt.Println(string(str))
}
