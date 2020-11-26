package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihexxa/gocfg"

	"github.com/ihexxa/quickshare/src/handlers"
)

type Server struct {
	server *http.Server
}

type ServerCfg struct {
	Addr           string `json:"addr"`
	ReadTimeout    int    `json:"readTimeout"`
	WriteTimeout   int    `json:"writeTimeout"`
	MaxHeaderBytes int    `json:"maxHeaderBytes"`
}

func NewServer(cfg gocfg.ICfg) (*Server, error) {
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router, err := addHandlers(router)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		// TODO: set more options
		Addr:           cfg.StringOr("ServerCfg.Addr", ":8080"),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.IntOr("ServerCfg.ReadTimeout", 1)) * time.Second,
		WriteTimeout:   time.Duration(cfg.IntOr("ServerCfg.ReadTimeout", 1)) * time.Second,
		MaxHeaderBytes: cfg.IntOr("ServerCfg.MaxHeaderBytes", 512),
	}

	return &Server{
		server: srv,
	}, nil
}

func addHandlers(router *gin.Engine) (*gin.Engine, error) {
	v1 := router.Group("/v1")

	users := v1.Group("/users")
	users.POST("/login", handlers.Login)
	users.POST("/logout", handlers.Logout)

	files := v1.Group("files")
	files.POST("/upload/", handlers.Upload)
	files.GET("/list/", handlers.List)
	files.DELETE("/delete/", handlers.Delete)
	files.GET("/metadata/", handlers.Metadata)
	files.POST("/copy/", handlers.Copy)
	files.POST("/move/", handlers.Move)

	return router, nil
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}
