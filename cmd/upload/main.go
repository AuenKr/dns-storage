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

func main() {
	app := fx.New(
		internal.Module,
		pkg.Module,
		fx.Invoke(runApp),
	)

	app.Run()
}

func runApp(cloudflareCli *handler.CloudflareDNS, dnsCli handler.DNSTXTHandler, logger *zap.Logger, config *defaults.DefaultConfig) {
	ctx := context.Background()

	// TXTRecord := []byte("test.auenkr.qzz.io")
	//
	// createResult, err := cloudflareCli.CreateTXTRecord("test2", TXTRecord)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("CreateTXTRecord")
	// fmt.Printf("%#v\n", createResult)
	//
	url := fmt.Sprintf("%s.%s.", "example", config.BaseURL)
	// listResult, err := cloudflareCli.GetTXTRecords(url, 1)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("GetTXTRecords")
	// fmt.Printf("%#v\n", listResult)

	res, err := dnsCli.ReadTXTRecord(ctx, url)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("ReadTXTRecord")
	fmt.Printf("%#v\n", res)

	// deleteResult, err := cloudflareCli.DeleteTXTRecord(createResult.Result.ID)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("DeleteTXTRecord")
	// fmt.Printf("%#v\n", deleteResult)
}
