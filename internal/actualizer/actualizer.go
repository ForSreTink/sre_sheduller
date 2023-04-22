package actualizer

import (
	"context"
	"log"
	"time"
	"workScheduler/internal/repository"
)

type Actualizer struct {
	Repository repository.ReadWriteRepository
}

func NewActualizer(repo repository.ReadWriteRepository) *Actualizer {
	return &Actualizer{
		Repository: repo,
	}
}

func (a *Actualizer) Run(ctx context.Context) {
	go a.actualize(ctx)
}

func (a *Actualizer) actualize(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			location, err := time.LoadLocation("")
			if err != nil {
				log.Printf("WARNING: Find error while parsing start time for actualizer, %s\n", err)
				continue
			}
			now := time.Now()
			first_time := time.Date(1970, 1, 1, 0, 0, 0, 0, location)
			works, err := a.Repository.List(ctx, first_time, now, []string{}, []string{"planned", "in_progress"})
			if err != nil {
				log.Printf("WARNING: Find error while getting works for actualizer from data, %s\n", err)
				continue
			}
			for _, work := range works {
				wStartUnix := work.StartDate.Unix()
				wEndunix := work.EndTime().Unix()
				nowUnix := now.Unix()
				if wStartUnix < nowUnix && wEndunix > nowUnix {
					work.Status = "in_progress"
					a.Repository.Update(ctx, work)
				} else if wStartUnix < nowUnix && wEndunix < nowUnix {
					work.Status = "completed"
					a.Repository.Update(ctx, work)
				}
			}
		}
	}
}
