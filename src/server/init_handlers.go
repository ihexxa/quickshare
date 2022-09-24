package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/depidx"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
	"github.com/ihexxa/quickshare/src/handlers/multiusers"
	"github.com/ihexxa/quickshare/src/handlers/settings"
	qsstatic "github.com/ihexxa/quickshare/static"
)

func initHandlers(cfg gocfg.ICfg, deps *depidx.Deps) (*gin.Engine, error) {
	router := gin.Default()

	// handlers
	userHdrs, err := multiusers.NewMultiUsersSvc(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new users svc error: %w", err)
	}

	adminName := cfg.GrabString("ENV.DEFAULTADMIN")
	_, err = userHdrs.Init(context.TODO(), adminName)
	if err != nil {
		return nil, fmt.Errorf("failed to init user handlers: %w", err)
	}

	fileHdrs, err := fileshdr.NewFileHandlers(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new files service error: %w", err)
	}
	settingsSvc, err := settings.NewSettingsSvc(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("new setting service error: %w", err)
	}

	// middlewares
	router.Use(userHdrs.AuthN())
	router.Use(userHdrs.APIAccessControl())

	publicPath, ok := cfg.String("Server.PublicPath")
	if !ok || publicPath == "" {
		return nil, errors.New("publicPath not found or empty")
	}
	if cfg.BoolOr("Server.Debug", false) {
		router.Use(static.Serve("/", static.LocalFile(publicPath, false)))
	} else {
		embedFs, err := qsstatic.NewEmbedStaticFS()
		if err != nil {
			return nil, err
		}
		router.Use(static.Serve("/", embedFs))
	}

	// handlers
	v1 := router.Group("/v1")

	usersAPI := v1.Group("/users")
	usersAPI.POST("/login", userHdrs.Login)
	usersAPI.POST("/logout", userHdrs.Logout)
	usersAPI.GET("/isauthed", userHdrs.IsAuthed)
	usersAPI.PATCH("/pwd", userHdrs.SetPwd)
	usersAPI.PATCH("/pwd/force-set", userHdrs.ForceSetPwd)
	usersAPI.POST("/", userHdrs.AddUser)
	usersAPI.DELETE("/", userHdrs.DelUser)
	usersAPI.GET("/list", userHdrs.ListUsers)
	usersAPI.GET("/self", userHdrs.Self)
	usersAPI.PATCH("/", userHdrs.SetUser)
	usersAPI.PATCH("/preferences", userHdrs.SetPreferences)
	usersAPI.PUT("/used-space", userHdrs.ResetUsedSpace)

	rolesAPI := v1.Group("/roles")
	// rolesAPI.POST("/", userHdrs.AddRole)
	// rolesAPI.DELETE("/", userHdrs.DelRole)
	rolesAPI.GET("/list", userHdrs.ListRoles)

	captchaAPI := v1.Group("/captchas")
	captchaAPI.GET("/", userHdrs.GetCaptchaID)
	captchaAPI.GET("/imgs", userHdrs.GetCaptchaImg)

	filesAPI := v1.Group("/fs")
	filesAPI.POST("/files", fileHdrs.Create)
	filesAPI.DELETE("/files", fileHdrs.Delete)
	filesAPI.GET("/files", fileHdrs.Download)
	filesAPI.PATCH("/files/chunks", fileHdrs.UploadChunk)
	filesAPI.GET("/files/chunks", fileHdrs.UploadStatus)
	filesAPI.PATCH("/files/copy", fileHdrs.Copy)
	filesAPI.PATCH("/files/move", fileHdrs.Move)

	filesAPI.GET("/dirs", fileHdrs.List)
	filesAPI.GET("/dirs/home", fileHdrs.ListHome)
	filesAPI.POST("/dirs", fileHdrs.Mkdir)
	// files.POST("/dirs/copy", fileHdrs.CopyDir)

	filesAPI.GET("/uploadings", fileHdrs.ListUploadings)
	filesAPI.DELETE("/uploadings", fileHdrs.DelUploading)

	filesAPI.POST("/sharings", fileHdrs.AddSharing)
	filesAPI.DELETE("/sharings", fileHdrs.DelSharing)
	filesAPI.GET("/sharings", fileHdrs.ListSharings)
	filesAPI.GET("/sharings/ids", fileHdrs.ListSharingIDs)
	filesAPI.GET("/sharings/exist", fileHdrs.IsSharing)
	filesAPI.GET("/sharings/dirs", fileHdrs.GetSharingDir)

	filesAPI.GET("/metadata", fileHdrs.Metadata)
	filesAPI.GET("/search", fileHdrs.SearchItems)
	filesAPI.PUT("/reindex", fileHdrs.Reindex)

	filesAPI.POST("/hashes/sha1", fileHdrs.GenerateHash)

	settingsAPI := v1.Group("/settings")
	settingsAPI.OPTIONS("/health", settingsSvc.Health)
	settingsAPI.GET("/client", settingsSvc.GetClientCfg)
	settingsAPI.PATCH("/client", settingsSvc.SetClientCfg)
	settingsAPI.POST("/errors", settingsSvc.ReportErrors)
	settingsAPI.GET("/workers/queue-len", settingsSvc.WorkerQueueLen)

	return router, nil
}
