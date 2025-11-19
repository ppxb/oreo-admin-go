package config

import (
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/ppxb/oreo-admin-go/pkg/log"
)

// TODO: NEED REFACTOR

type ConfBox struct {
	Ctx context.Context
	Fs  embed.FS
	Dir string
}

func (c ConfBox) Get(filename string) []byte {
	if filename == "" {
		return nil
	}

	path := c.buildPath(filename)
	if data := c.readFromFileSystem(path); data != nil {
		return data
	}
	return c.readFromEmbed(path)
}

func (c ConfBox) buildPath(filename string) string {
	return filepath.Join(c.Dir, filename)
}

func (c ConfBox) readFromFileSystem(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.WithContext(c.Ctx).WithError(err).Debug("[CONF BOX] Read file %s from file system failed, will try embed", path)
		return nil
	}

	log.WithContext(c.Ctx).Debug("[CONF BOX] Read file %s from file system success", path)
	return data
}

func (c ConfBox) readFromEmbed(path string) []byte {
	data, err := c.Fs.ReadFile(path)
	if err != nil {
		log.WithContext(c.Ctx).WithError(err).Warn("[CONF BOX] Read file %s from embed failed", path)
		return nil
	}

	if len(data) == 0 {
		log.WithContext(c.Ctx).Warn("[CONF BOX] File %s is empty in embed", path)
		return nil
	}

	log.WithContext(c.Ctx).Debug("[CONF BOX] Read file %s from embed success", path)
	return data
}
