package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"workScheduler/internal/scheduler/app"
)

// TODO: Отслеживать сигналы, тушить все и на монге делать коннекшен клоус + кенсел контекста
// КОНТЕКСТЫ
// TODO: Отслеживать изменения конфига, разбирать его (возможно с использованием сигнала)
// Json теги для work структуры

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := app.NewScheduler(ctx)
	s.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	s.Stop()
	log.Println("Shutting down")
	os.Exit(0)
}
