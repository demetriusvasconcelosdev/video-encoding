package main

import (
	"fmt"
	"log"

	"github.com/dl/encoding-video/internal/config"
	"github.com/dl/encoding-video/internal/encoder"
)

func main() {
	cfg := config.Load()
	if err := encoder.Run(cfg); err != nil {
		log.Fatal(err)
	}
	fmt.Println("done")
}
