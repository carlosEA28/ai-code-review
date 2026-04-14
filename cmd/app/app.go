package main

import (
	"log"

	"github.com/carlosEA28/ai-code-review/internal/web/server"
)

func main() {

	port := "3000"
	srv := server.NewServer(port)

	if err := srv.Start(); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
