package main

import (
	"fmt"
	"url-shortener/cmd/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
}
