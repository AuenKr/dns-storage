package cli

import (
	"context"
	"errors"
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

	for {
		select {
		case status, ok := <-statusChan:
			if !ok {
				return errors.New("status channel closed")
			}
			fmt.Println(status)
		case err, ok := <-errChan:
			if !ok {
				return errors.New("error channel closed")
			}
			fmt.Println(err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Download implements [CommandLine].
func (c *cli) Download(ctx context.Context, indexFileRecord string, filePath string) error {
	statusChan, errChan := c.FileHandler.Download(ctx, indexFileRecord, filePath)

	for {
		select {
		case status, ok := <-statusChan:
			if !ok {
				return errors.New("status channel closed")
			}
			fmt.Println(status)
		case err, ok := <-errChan:
			if !ok {
				return errors.New("error channel closed")
			}
			fmt.Println(err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Delete implements [CommandLine].
func (c *cli) Delete(ctx context.Context, indexFileRecord string) error {
	statusChan, errChan := c.FileHandler.Delete(ctx, indexFileRecord)

	for {
		select {
		case status, ok := <-statusChan:
			if !ok {
				return errors.New("status channel closed")
			}
			fmt.Println(status)
		case err, ok := <-errChan:
			if !ok {
				return errors.New("error channel closed")
			}
			fmt.Println(err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Stream implements [CommandLine].
func (c *cli) Stream(ctx context.Context, indexFileRecord string) error {
	panic("unimplemented")
}

func NewCommandLine(
	fileHandler handler.FileHandlerProvider,
) CommandLine {
	return &cli{
		FileHandler: fileHandler,
	}
}
