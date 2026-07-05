package handler

import (
	"strconv"

	"chatgoo/internal/pkg/response"
	"chatgoo/internal/service"

	"gofr.dev/pkg/gofr"
)

// GetFile returns file info.
func GetFile(fileSvc service.FileService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		fileID, err := strconv.ParseInt(c.PathParam("fileID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid fileID")
		}

		file, err := fileSvc.GetByID(c, fileID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(file), nil
	}
}
