package main

import (
	"os"

	"github.com/teiyou416/simblock_go/internal/app"
)

func main() {
	app.Run(os.Args[1:])
}
