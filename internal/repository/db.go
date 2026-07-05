package repository

import (
	"gofr.dev/pkg/gofr/container"
)

// getDB is a convenience alias — repositories receive container.DB at construction.
func getDB(c *container.Container) container.DB {
	return c.SQL
}
