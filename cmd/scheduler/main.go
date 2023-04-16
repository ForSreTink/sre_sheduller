package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"workScheduler/internal/scheduler/server"
)

// TODO: Отслеживать сигналы, тушить все и на монге делать коннекшен клоус + кенсел контекста
// TODO: Отслеживать изменения конфига, разбирать его (возможно с использованием сигнала)
// Json теги для work структуры

func main() {
	log.Println("Starting application...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := server.NewServer(ctx, "./config.yml")
	s.Run()
	log.Println("Application started!")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	s.Stop()
	log.Println("Shutting down")
	os.Exit(0)
}
