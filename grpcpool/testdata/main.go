package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/spf13/cobra"
)

var message string

func init() {
	nonce := make([]byte, 256)
	io.ReadFull(rand.Reader, nonce)

	message = hex.EncodeToString(nonce)
}

func main() {
	root := &cobra.Command{
		Use: "testdata",
	}

	root.AddCommand(
		server(),
		client(),
		dummy(),
	)

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
