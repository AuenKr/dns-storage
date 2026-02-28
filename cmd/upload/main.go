package main

import (
	"context"
	"flag"
	"fmt"

	"dns-storage/internal"
	"dns-storage/internal/handler"
	"dns-storage/pkg"
	"dns-storage/pkg/cli"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var _ = godotenv.Load()

func main() {
	filePath := flag.String("path", "", "path to file for upload")
	index := flag.String("index", "", "index/domain key for download/delete")
	flag.Parse()
	args := cli.CLIArgs{
		FilePath: *filePath,
		Index:    *index,
	}

	app := fx.New(
		internal.Module,
		pkg.Module,
		fx.Supply(args),
		fx.Invoke(runApp),
	)

	app.Run()
}

func runApp(fileHandler handler.FileHandler, args cli.CLIArgs) {
	if args.FilePath == "" {
		fmt.Println("missing required --path")
		return
	}
	ctx := context.Background() // File Upload
	indexFile, err := fileHandler.Upload(ctx, args.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(indexFile)
}
