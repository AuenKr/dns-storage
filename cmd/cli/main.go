package main

import (
	"context"
	"errors"
	"fmt"

	"dns-storage/internal"
	"dns-storage/internal/cli"
	"dns-storage/pkg"
	cliPkg "dns-storage/pkg/cli"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		pkg.Module,
		internal.Module,
		fx.Invoke(Run),
	)

	if err := app.Start(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func Run(commandLine cli.CommandLine, flags cliPkg.Flags) {
	fmt.Println("Flags:", flags)
	ctx := context.Background()
	var err error

	switch flags.Mode {
	case cliPkg.Download:
		err = commandLine.Download(ctx, flags.Subdomain, flags.Path)
	case cliPkg.Upload:
		err = commandLine.Upload(ctx, flags.Path, flags.Subdomain)
	case cliPkg.Delete:
		err = commandLine.Delete(ctx, flags.Subdomain)
	case cliPkg.Stream:
		err = commandLine.Stream(ctx, flags.Subdomain)
	default:
		err = errors.New("unknown mode" + string(flags.Mode))
	}

	if err != nil {
		fmt.Println(err)
	}
}
