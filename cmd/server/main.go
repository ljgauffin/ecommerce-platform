package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/database"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	apphttp "github.com/lucasgauffin/ecommerce-platform/internal/presentation/http"
	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbPort := os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "5433"
		}
		dsn = fmt.Sprintf("host=%s user=ecommerce password=ecommerce_secret dbname=ecommerce_db port=%s sslmode=disable", dbHost, dbPort)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to database...")
	db, err := database.NewGormDB(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database connected and schema auto-migrated successfully.")

	productRepo := persistence.NewGormProductRepo(db.Gorm)
	productHandler := apphttp.NewProductHandler(productRepo)
	csvUC := usecase.NewCsvImportUseCase(productRepo)
	csvHandler := apphttp.NewCsvHandler(csvUC)
	swaggerHandler := apphttp.NewSwaggerHandler()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Idempotency-Key"},
		ExposedHeaders:   []string{"Link", "Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger", swaggerHandler.UI)
	r.Get("/swagger/*", swaggerHandler.UI)
	r.Get("/swagger/doc.json", swaggerHandler.Doc)

	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "UP"})
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sqlDB, err := db.Gorm.DB()
		if err != nil || sqlDB.PingContext(r.Context()) != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status":   "DOWN",
				"database": "disconnected",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "READY",
			"database": "connected",
		})
	})

	r.Route("/api/v1/products", func(r chi.Router) {
		r.Get("/", productHandler.List)
		r.Post("/", productHandler.Create)
		r.Post("/import-csv", csvHandler.ImportCSV)
		r.Get("/{id}", productHandler.GetByID)
		r.Put("/{id}", productHandler.Update)
		r.Delete("/{id}", productHandler.Delete)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancelShutdown := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancelShutdown()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("Server listening on port :%s", port)
	log.Printf("Swagger UI available at http://localhost:%s/swagger", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	<-serverCtx.Done()
	log.Printf("Server stopped gracefully.")
}
