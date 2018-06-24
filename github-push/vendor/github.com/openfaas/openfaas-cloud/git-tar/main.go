package main

import (
	"io/ioutil"
	"log"
	"os"

	"handler/function"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Unable to read standard input: %s", err.Error())
	}

	os.Stdout.Write(function.Handle(input))
}
