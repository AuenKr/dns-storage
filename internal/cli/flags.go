package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	pkgCli "dns-storage/pkg/cli"
)

func NewFlags() pkgCli.Flags {
	flags, err := parseFlags()
	if err != nil {
		panic(fmt.Sprint("Invalid Flags\n", err))
	}
	return flags
}

func parseFlags() (pkgCli.Flags, error) {
	arguments := pkgCli.Flags{}
	var mode string
	var subdomain string
	var path string

	// Flag defination
	flag.StringVar(&mode, "mode", "", "modes: download, upload, delete, stream")
	flag.StringVar(&subdomain, "subdomain", "", "subdomain for index file")
	flag.StringVar(&path, "path", "", "path of file")

	// Parsing Flags
	flag.Parse()

	allError := make([]error, 0, 2)
	switch pkgCli.Mode(mode) {
	case pkgCli.Download:
		if subdomain == "" {
			allError = append(allError, errors.New("--subdomain flag: subdomain(index file) is required"))
		}
		if path == "" {
			allError = append(allError, errors.New("--path flag: download filepath is required"))
		}

	case pkgCli.Upload:
		if subdomain == "" {
			allError = append(allError, errors.New("--subdomain flag: subdomain(index file) is required"))
		}
		if path == "" {
			allError = append(allError, errors.New("--path flag: upload file path is required"))
		}
	case pkgCli.Delete:
		if subdomain == "" {
			allError = append(allError, errors.New("--subdomain flag: subdomain(index file) is required"))
		}
	case pkgCli.Stream:
		if subdomain == "" {
			allError = append(allError, errors.New("--subdomain flag: subdomain(index file) is required"))
		}
	default:
		return arguments, errors.New("invalid mode " + mode)
	}

	if len(allError) > 0 {
		var errMsg strings.Builder
		for _, err := range allError {
			errMsg.WriteString(err.Error() + "\n")
		}
		return arguments, fmt.Errorf("%v", errMsg.String())
	}

	arguments.Mode = pkgCli.Mode(mode)
	arguments.Subdomain = subdomain
	arguments.Path = path

	return arguments, nil
}
