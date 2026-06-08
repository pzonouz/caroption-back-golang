package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/routes"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

type config struct {
	addr string
}

type application struct {
	config config
	db     *pgxpool.Pool
	mux    *chi.Mux
}

func backupHandler(w http.ResponseWriter, r *http.Request) {
	// run the bash script
	cmd := exec.CommandContext(r.Context(), "bash", "./manual-backup.sh")

	err := cmd.Run()
	if err != nil {
		http.Error(w, "backup failed: "+err.Error(), http.StatusInternalServerError)

		return
	}

	// generate filename
	date := time.Now().Format("2006-01-02")
	file := fmt.Sprintf("/tmp/backup-%s.tar.gz", date)

	// force download
	w.Header().Set("Content-Disposition", "attachment; filename=backup-"+date+".tar.gz")
	w.Header().Set("Content-Type", "application/gzip")

	http.ServeFile(w, r, file)
}

func restoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	force := r.URL.Query().Get("force") == "true"

	tmpFile, err := os.CreateTemp("/tmp", "restore-*.tar.gz")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer os.Remove(tmpFile.Name())

	// stream upload directly to disk
	_, err = io.Copy(tmpFile, r.Body)
	if err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)

		return
	}

	err = tmpFile.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	// validate archive structure
	err = utils.ValidateBackup(tmpFile.Name())
	if err != nil {
		http.Error(w, "invalid backup: "+err.Error(), http.StatusBadRequest)

		return
	}

	cmd := exec.CommandContext(ctx, "bash", "./manual-restore.sh", tmpFile.Name())

	cmd.Stdout = w
	cmd.Stderr = w

	w.Header().Set("Content-Type", "text/plain")

	if !force {
		w.Write([]byte("Overwrite protection enabled. Use ?force=true to allow restore.\n"))

		return
	}

	err = cmd.Run()
	if err != nil {
		http.Error(w, "\nrestore failed: "+err.Error(), http.StatusInternalServerError)

		return
	}
}

func (app *application) mount() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	service := services.New(app.db)
	routes.GenerateEntityRoutes(router, service)
	routes.GenerateProductRoutes(router, service)
	routes.GenerateCategoryRoutes(router, service)
	routes.GenerateBrandRoutes(router, service)
	routes.GenerateImageRoutes(router, service)
	routes.GenerateParameterGroupsRoutes(router, service)
	routes.GenerateParametersRoutes(router, service)
	routes.GenerateAuthRoutes(router, service)
	routes.GenerateArticleRoutes(router, service)
	routes.GenerateInvoiceRoutes(router, service)
	routes.GeneratePersonRoutes(router, service)
	routes.GenerateReportsRoutes(router, service)
	routes.GenerateAccountRoutes(router, service)
	routes.GenerateSettingsRoutes(router, service)
	routes.GenerateBankRoutes(router, service)
	routes.GenerateCheckRoutes(router, service)
	routes.GenerateCheckBandRoutes(router, service)
	routes.GenerateVoucherRoutes(router, service)
	routes.GenerateFinancialOperationsRoutes(router, service)
	routes.GenerateAvvalDorehEditRoutes(router, service)
	routes.GenerateInvoiceProductsGroupsRoutes(router, service)
	router.Post("/upload-file", func(w http.ResponseWriter, r *http.Request) {
		_ = utils.Uploader(w, r)
	})
	router.With(middlewares.AdminOnly).Get("/backup", backupHandler)
	router.With(middlewares.AdminOnly).Post("/restore-backup", restoreBackupHandler)

	return router
}

func (app *application) initDB() {
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_DBNAME")

	conn, err := pgxpool.New(
		context.Background(),
		"postgres://root:"+databasePassword+"@localhost:5432/"+databaseName,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	log.Print("Connected to Database")

	app.db = conn
}

func (app *application) run(mux *chi.Mux) error {
	server := &http.Server{
		Addr:    ":" + app.config.addr,
		Handler: mux,
	}
	log.Printf("Starting Server on %s", app.config.addr)

	return server.ListenAndServe()
}
