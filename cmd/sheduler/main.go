package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	api "workScheduler/internal/api/app"
	handlers "workScheduler/internal/handlers"
	inmemoryrepository "workScheduler/internal/repository/inmemory_repository"

	"github.com/go-openapi/runtime/middleware"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	// ctx := context.Background()
	// data, err := mongo.NewMongoClient(ctx)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	// 	os.Exit(1)
	// }

	data := inmemoryrepository.NewInmemoryRepository()
	sheduller := api.NewApi(data)

	var sh http.Handler = middleware.SwaggerUI(middleware.SwaggerUIOpts{
		SpecURL: "./static/api.yaml",
		Path:    "/swagger",
	}, nil)

	r := mux.NewRouter()
	api.HandlerFromMux(sheduller, r)
	r.HandleFunc("/health", handlers.HealthCheckHandler).Methods("GET")
	r.Handle("/swagger", sh).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./openapi"))))

	loggedRouter := gorilla_handlers.LoggingHandler(os.Stdout, r)

	s := &http.Server{
		Handler: loggedRouter,
		Addr:    fmt.Sprintf("0.0.0.0:%d", *port),
	}
	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
