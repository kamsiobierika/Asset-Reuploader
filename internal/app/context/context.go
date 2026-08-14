package context

import (
	"github.com/kartFr/Asset-Reuploader/internal/app/response"
	"github.com/kartFr/Asset-Reuploader/internal/files"
	"github.com/kartFr/Asset-Reuploader/internal/roblox"
)

type Context struct {
	Client          *roblox.Client
	Logger          *logger
	PauseController *pauseController
	Response        *response.Response
	DebugWriter     *files.DebugWriter
}

func New(c *roblox.Client, resp *response.Response) *Context {
	return &Context{
		Client:          c,
		Logger:          newLogger(),
		PauseController: newPauseController(),
		Response:        resp,
		DebugWriter:     nil,
	}
}

func NewWithDebug(c *roblox.Client, resp *response.Response, assetType string) *Context {
	return &Context{
		Client:          c,
		Logger:          newLogger(),
		PauseController: newPauseController(),
		Response:        resp,
		DebugWriter:     files.NewDebugWriter(assetType),
	}
}
