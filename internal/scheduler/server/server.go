package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"workScheduler/internal/actualizer"
	api "workScheduler/internal/api/app"
	"workScheduler/internal/configuration"
	handlers "workScheduler/internal/handlers"
	"workScheduler/internal/scheduler/app"

	mongo "workScheduler/internal/repository/mongo_integrations"

	// inmemoryrepository "workScheduler/internal/repository/inmemory_repository"
	tmp "workScheduler/internal/scheduler/server/templates"

	"github.com/go-openapi/runtime/middleware"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	Ctx    context.Context
	Server *http.Server
	Config *configuration.Configurator
}

func NewServer(ctx context.Context, config string) *Server {
	c := configuration.NewConfigurator(ctx, config)
	c.Run()
	return &Server{
		Ctx:    ctx,
		Config: c,
	}
}

func (s *Server) Run() error {
	var err error
	log.Println("Start web server...")
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	data, err := mongo.NewMongoClient(s.Ctx)
	if err != nil {
		return err
	}

	a := actualizer.NewActualizer(data)
	a.Run(s.Ctx)

	// data := inmemoryrepository.NewInmemoryRepository()
	scheduler := app.NewScheduler(s.Ctx, data, s.Config)
	Server := api.NewApi(data, scheduler, s.Config)

	var sh http.Handler = middleware.SwaggerUI(middleware.SwaggerUIOpts{
		SpecURL: "./static/api.yaml",
		Path:    "/swagger",
	}, nil)

	t := tmp.NewTemplate(data)

	r := mux.NewRouter()
	api.HandlerFromMux(Server, r)
	r.HandleFunc("/", t.Generate).Methods("GET")
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

func (s *Server) Stop() {
	s.Server.Shutdown(s.Ctx)
}
