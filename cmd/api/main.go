package main

import (
	"log"

	"github.com/go-chi/chi/v5"

	env "github.com/pzonouz/pzonouz-caroption-back-golang/internal"
)

func main() {
	cfg := &config{
		addr: env.GetString("ADDR", "8080"),
	}
	app := &application{
		config: *cfg,
		mux:    chi.NewMux(),
	}
	app.initDB()

	mux := app.mount()
	defer app.db.Close()

	err := app.run(mux)
	if err != nil {
		log.Panicln(err.Error())
	}
}
