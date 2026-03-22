package main

import (
	"context"
	"fmt"
	"net/http"

	"dns-storage/internal"
	"dns-storage/internal/handler"
	"dns-storage/pkg"
	"dns-storage/pkg/defaults"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var _ = godotenv.Load()

func main() {
	app := fx.New(
		internal.Module,
		pkg.Module,
		fx.Invoke(runServer),
	)

	app.Run()
}

func runServer(lc fx.Lifecycle, server handler.APIHandler, config *defaults.DefaultConfig) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", server.Health)
	mux.HandleFunc("/upload", server.Upload)
	mux.HandleFunc("/download", server.Download)
	mux.HandleFunc("/delete", server.Delete)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Println("http server error:", err)
				}
			}()
			fmt.Println("Server running on port", config.HTTPPort)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return httpServer.Shutdown(ctx)
		},
	})
}
