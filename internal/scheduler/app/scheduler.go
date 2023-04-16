package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	api "workScheduler/internal/api/app"
	"workScheduler/internal/configuration"
	handlers "workScheduler/internal/handlers"

	mongo "workScheduler/internal/repository/mongo_integrations"

	// inmemoryrepository "workScheduler/internal/repository/inmemory_repository"

	"github.com/go-openapi/runtime/middleware"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Scheduler struct {
	Ctx    context.Context
	Server *http.Server
	Config *configuration.Configurator
}

func NewScheduler(ctx context.Context, config string) *Scheduler {
	c := configuration.NewConfigurator(ctx, config)
	c.Run()
	return &Scheduler{
		Ctx:    ctx,
		Config: c,
	}
}

func (s *Scheduler) Run() error {
	var err error
	log.Println("Start web server...")
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	data, err := mongo.NewMongoClient(s.Ctx)
	if err != nil {
		return err
	}

	// data := inmemoryrepository.NewInmemoryRepository()
	scheduler := api.NewApi(data)

	var sh http.Handler = middleware.SwaggerUI(middleware.SwaggerUIOpts{
		SpecURL: "./static/api.yaml",
		Path:    "/swagger",
	}, nil)

	r := mux.NewRouter()
	api.HandlerFromMux(scheduler, r)
	r.HandleFunc("/health", handlers.HealthCheckHandler).Methods("GET")
	r.Handle("/swagger", sh).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./openapi"))))

	loggedRouter := gorilla_handlers.LoggingHandler(os.Stdout, r)

	s.Server = &http.Server{
		Handler:     loggedRouter,
		Addr:        fmt.Sprintf("0.0.0.0:%d", *port),
		BaseContext: func(l net.Listener) context.Context { return s.Ctx },
	}

	go func() {
		err = s.Server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("Web server started!")
	return nil

}

func (s *Scheduler) Stop() {
	s.Server.Shutdown(s.Ctx)
}
