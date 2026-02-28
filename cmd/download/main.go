package main

import (
	"context"
	"fmt"

	"dns-storage/internal"
	"dns-storage/internal/handler"
	"dns-storage/pkg"
	"dns-storage/pkg/defaults"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

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
	downloadPath, err := fileHandler.Download(ctx, IndexFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Download Path:", downloadPath)
}
