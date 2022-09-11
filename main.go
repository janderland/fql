package main

import "github.com/janderland/fdbq/internal/app"

func main() {
	if err := app.Fdbq.Execute(); err != nil {
		panic(err)
	}
}
