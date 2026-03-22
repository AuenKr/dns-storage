package cli

import (
	"context"
	"fmt"

	"dns-storage/internal/handler"
)

type CommandLine interface {
	Upload(ctx context.Context, filePath string, subdomain string) error
	Download(ctx context.Context, indexFileRecord string, filePath string) error
	Delete(ctx context.Context, indexFileRecord string) error
	Stream(ctx context.Context, indexFileRecord string) error
}

type cli struct {
	FileHandler handler.FileHandlerProvider
}

// Upload implements [CommandLine].
func (c *cli) Upload(ctx context.Context, filePath string, subdomain string) error {
	statusChan, errChan := c.FileHandler.Upload(ctx, filePath, subdomain)
	return waitForResult(ctx, statusChan, errChan, func(status handler.FileStatus) {
		fmt.Println(status)
	})
}

// Download implements [CommandLine].
func (c *cli) Download(ctx context.Context, indexFileRecord string, filePath string) error {
	statusChan, errChan := c.FileHandler.Download(ctx, indexFileRecord, filePath)
	return waitForResult(ctx, statusChan, errChan, func(status handler.FileStatus) {
		fmt.Println("status:", status)
	})
}

// Delete implements [CommandLine].
func (c *cli) Delete(ctx context.Context, indexFileRecord string) error {
	statusChan, errChan := c.FileHandler.Delete(ctx, indexFileRecord)
	return waitForResult(ctx, statusChan, errChan, func(status handler.FileStatus) {
		fmt.Println(status)
	})
}

// Stream implements [CommandLine].
func (c *cli) Stream(ctx context.Context, indexFileRecord string) error {
	statusChan, errChan := c.FileHandler.Stream(ctx, indexFileRecord)
	return waitForResult(ctx, statusChan, errChan, func(status handler.FileStream) {
		fmt.Printf("%#v", status)
	})
}

func waitForResult[T any](ctx context.Context, statusChan <-chan T, errChan <-chan error, printStatus func(T)) error {
	for statusChan != nil || errChan != nil {
		select {
		case status, ok := <-statusChan:
			if !ok {
				statusChan = nil
				continue
			}
			printStatus(status)
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func NewCommandLine(
	fileHandler handler.FileHandlerProvider,
) CommandLine {
	return &cli{
		FileHandler: fileHandler,
	}
}
