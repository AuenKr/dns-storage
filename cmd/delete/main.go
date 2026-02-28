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

const IndexFile = "46112583-acf0-4b8d-ba71-f541454d3480.auenkr.qzz.io"

func main() {
	filePath := flag.String("filePath", "", "path to file for upload")
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
	if args.Index == "" {
		fmt.Println("missing required --index")
		return
	}
	ctx := context.Background()
	err := fileHandler.Delete(ctx, args.Index)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("DeleteFile Path:", args.Index)
}
