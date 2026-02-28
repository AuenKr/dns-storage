package main

import (
	"context"
	"fmt"

	"dns-storage/internal"
	"dns-storage/internal/handler"
	"dns-storage/pkg"
	"dns-storage/pkg/defaults"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var _ = godotenv.Load()

const FilePath = ".temp/perfect.mp3"

// const FilePath = ".temp/test.txt"

func main() {
	app := fx.New(
		internal.Module,
		pkg.Module,
		fx.Invoke(runApp),
	)

	app.Run()
}

func runApp(cloudflareCli *handler.CloudflareDNS, dnsCli handler.DNSTXTHandler, fileHandler handler.FileHandler, logger *zap.Logger, config *defaults.DefaultConfig) {
	ctx := context.Background()

	// File Upload
	indexFile, err := fileHandler.Upload(ctx, FilePath)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(indexFile)
}
