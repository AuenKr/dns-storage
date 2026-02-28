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

const IndexFile = "image.jpeg.auenkr.qzz.io"

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
	err := fileHandler.Delete(ctx, IndexFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("DeleteFile Path:", IndexFile)
}
