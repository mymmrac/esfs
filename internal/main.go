package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/mymmrac/esfs"
)

//go:embed testdata
var dist embed.FS

func main() {
	// err := esfs.ServeDir("localhost:8080", "internal",
	err := esfs.ServeFS("localhost:8080", dist,
		esfs.WithSubDir("testdata"),
		esfs.WithPathRewriteToRoot(),
		esfs.WithGracefulShutdown(),
		esfs.WithCompressBrotli(),
	)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
