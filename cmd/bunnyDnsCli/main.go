package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"dns-storage/internal"
	"dns-storage/internal/handler"
	"dns-storage/pkg"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		pkg.Module,
		internal.Module,
		fx.Provide(NewDNSTXTCliOption),
		fx.Invoke(Run),
	)

	if err := app.Start(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func Run(flags DNSTXTCliOption, dnsTxtProvider handler.DNSTXTProvider) {
	fmt.Println("Flags:", flags)
	ctx := context.Background()

	var record handler.Record
	var err error

	switch flags.Mode {
	case Create:
		record, err = dnsTxtProvider.CreateTXTRecord(ctx, flags.Subdomain, flags.Value)
	case Get:
		record, err = dnsTxtProvider.GetTXTRecords(ctx, flags.Subdomain)
	case Delete:
		err = dnsTxtProvider.DeleteTXTRecord(ctx, flags.ID)
	case ResetDomain:
		now := time.Now()
		records, err := dnsTxtProvider.GetAllRecord(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Time Taken to fetch all records", time.Since(now))
		batchSize := 50

		for i := 0; i < len(records); i += batchSize {
			startTime := time.Now()

			wg := sync.WaitGroup{}

			for j := 0; i+j < len(records) && j < batchSize; j++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					err = dnsTxtProvider.DeleteTXTRecord(ctx, strconv.Itoa(records[idx].ID))
					if err != nil {
						fmt.Println(err)
						return
					}
				}(i + j)
			}

			wg.Wait()
			fmt.Printf("Deleted %d records\n", i+batchSize)
			fmt.Println("Time Taken", time.Since(startTime))
		}

		fmt.Println("Total Time Taken", time.Since(now))
	case Test:
		flags.Subdomain = "temptemp2"
		temp := make([]int, 4000)
		txtRecord := ""
		for _, v := range temp {
			txtRecord = txtRecord + strconv.Itoa(v)
		}

		now := time.Now()
		for i := 2; ; {
			now := time.Now()
			record, err = dnsTxtProvider.CreateTXTRecord(ctx, flags.Subdomain, txtRecord)
			if err != nil {
				break
			}

			fmt.Printf("Created %d records\n", i)
			fmt.Println("Time Taken", time.Since(now))
			i++
		}

		fmt.Println("Time Taken", time.Since(now))

	default:
		err = errors.New("unknown mode" + string(flags.Mode))
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", record)
}

type TXTCLIMode string

const (
	Create      TXTCLIMode = "create"
	Get         TXTCLIMode = "get"
	Delete      TXTCLIMode = "delete"
	ResetDomain TXTCLIMode = "reset"
	Test        TXTCLIMode = "test"
)

type DNSTXTCliOption struct {
	Mode      TXTCLIMode
	Subdomain string
	Value     string
	ID        string
}

func NewDNSTXTCliOption() DNSTXTCliOption {
	flags, err := parseFlag()
	if err != nil {
		panic(err)
	}
	return flags
}

func parseFlag() (DNSTXTCliOption, error) {
	var mode, subdomain, value, id string
	flag.StringVar(&mode, "mode", "", "Mode to run")
	flag.StringVar(&subdomain, "subdomain", "", "Subdomain to add/update/delete")
	flag.StringVar(&value, "value", "", "Value to add/update")
	flag.StringVar(&id, "id", "", "ID of record to delete")
	flag.Parse()

	err := make([]error, 0, 2)
	switch TXTCLIMode(mode) {
	case Create:
		if subdomain == "" {
			err = append(err, errors.New("subdomain is required"))
		}
		if value == "" {
			err = append(err, errors.New("value is required"))
		}
	case Get:
		if subdomain == "" {
			err = append(err, errors.New("subdomain is required"))
		}
	case Delete:
		if id == "" {
			err = append(err, errors.New("id is required"))
		}
	case ResetDomain:
	// No input required
	case Test:
	// No input required
	default:
		err = append(err, errors.New("unknown mode"+mode))
	}

	flags := DNSTXTCliOption{}
	if len(err) > 0 {
		var errMsg strings.Builder
		for _, e := range err {
			errMsg.WriteString(e.Error() + "\n")
		}
		return flags, errors.New(errMsg.String())
	}

	flags.Mode = TXTCLIMode(mode)
	flags.Subdomain = subdomain
	flags.Value = value
	flags.ID = id

	fmt.Printf("Flags: %#v", flags)
	return flags, nil
}
