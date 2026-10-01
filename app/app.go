package app

import (
	"github.com/egermano/balde/core"
)

// App is the application service layer above core. It exposes typed
// operations with DTOs carrying the stable JSON wire schema — core types
// never cross the wire.
type App struct {
	budget *core.Budget
	store  core.Store
}

// New creates an App for the given budget backed by store.
func New(budgetID string, store core.Store) *App {
	return &App{
		budget: core.NewBudget(budgetID, store),
		store:  store,
	}
}

// bucketDTO converts a core bucket and adds the computed fill percentage.
func (a *App) bucketDTO(b core.Bucket) Bucket {
	dto := toBucketDTO(b)
	dto.FillPercent = a.budget.CalculateFillPercentage(b)
	return dto
}
