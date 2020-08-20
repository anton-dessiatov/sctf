package main

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/cmd"
)

func main() {
	instance, err := app.New()
	if err != nil {
		log.Fatal(fmt.Errorf("app.New: %w", err))
	}
	app.Instance = instance
	cmd.Execute()
}
