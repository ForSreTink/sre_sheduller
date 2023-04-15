package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	api "workScheduler/internal/api/app"
	inmemoryrepository "workScheduler/internal/api/models/repository/inmemory_repository"

	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"

	"github.com/gorilla/mux"
)

func main() {
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	data := inmemoryrepository.NewInmemoryRepository()
	sheduller := api.NewApi(data)

	r := mux.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(sheduller, r)

	s := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", *port),
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
