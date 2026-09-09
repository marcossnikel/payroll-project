package main

import (
	"log"
	"net/http"
	"os"

	"github.com/marcossnikel/payroll-project/internal/httpapi"
)

func main() {
	address := os.Getenv("PAYROLL_HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}

	app := httpapi.NewApp(httpapi.Config{})
	log.Printf("payroll operations API listening on http://localhost%s", address)
	if err := http.ListenAndServe(address, app.Handler()); err != nil {
		log.Fatal(err)
	}
}
