package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"dns-storage/internal"
	"dns-storage/internal/handler"
	"dns-storage/pkg"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var _ = godotenv.Load()

func main() {
	app := fx.New(
		internal.Module,
		pkg.Module,
		fx.Invoke(runDNSProvider),
	)

	app.Run()
}

func runDNSProvider(dnsProviderCli handler.DNSTXTProvider) {
	ctx := context.Background()

	subdomain := "example"
	txtRecord := []byte("testing for example subdomain from dns-storage")
	base64value := base64.StdEncoding.EncodeToString(txtRecord)
	createResult, err := dnsProviderCli.CreateTXTRecord(ctx, subdomain, base64value)
	if err != nil {
		fmt.Println("CreateTXTRecord", err)
		return
	}
	fmt.Println("CreateTXTRecord:")
	fmt.Printf("%#v\n", createResult)
	time.Sleep(time.Second * 2)

	getResult, err := dnsProviderCli.GetTXTRecords(ctx, subdomain)
	if err != nil {
		fmt.Println("GetTXTRecords", err)
		return
	}
	fmt.Println("GetTXTRecords:")
	fmt.Printf("%#v\n", getResult)
	time.Sleep(time.Second * 2)

	newTxtRecord := []byte("testing for example subdomain from dns-storage updated")
	base64value = base64.StdEncoding.EncodeToString(newTxtRecord)
	createResult.Content = base64value
	updateResult, err := dnsProviderCli.UpdateTXTRecord(ctx, createResult.ID, createResult)
	if err != nil {
		fmt.Println("UpdateTXTRecord", err)
		return
	}
	fmt.Println("UpdateTXTRecord:")
	fmt.Printf("%#v\n", updateResult)
	time.Sleep(time.Second * 2)

	err = dnsProviderCli.DeleteTXTRecord(ctx, updateResult.ID)
	if err != nil {
		fmt.Println("DeleteTXTRecord", err)
		return
	}
	fmt.Println("DeleteTXTRecord")
}
