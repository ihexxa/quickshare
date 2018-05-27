package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/apis"
	"github.com/ihexxa/quickshare/server/libs/cfg"
)

func main() {
	config := cfg.NewConfigFrom("config.json")
	srvShare := apis.NewSrvShare(config)

	// TODO: using httprouter instead
	mux := http.NewServeMux()
	mux.HandleFunc(config.PathLogin, srvShare.LoginHandler)
	mux.HandleFunc(config.PathStartUpload, srvShare.StartUploadHandler)
	mux.HandleFunc(config.PathUpload, srvShare.UploadHandler)
	mux.HandleFunc(config.PathFinishUpload, srvShare.FinishUploadHandler)
	mux.HandleFunc(config.PathDownload, srvShare.DownloadHandler)
	mux.HandleFunc(config.PathFileInfo, srvShare.FileInfoHandler)
	mux.HandleFunc(config.PathClient, srvShare.ClientHandler)

	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.HostName, config.Port),
		Handler:        mux,
		MaxHeaderBytes: config.MaxHeaderBytes,
		ReadTimeout:    time.Duration(config.ReadTimeout) * time.Millisecond,
		WriteTimeout:   time.Duration(config.WriteTimeout) * time.Millisecond,
		IdleTimeout:    time.Duration(config.IdleTimeout) * time.Millisecond,
	}

	log.Printf("quickshare starts @ %s:%d", config.HostName, config.Port)
	log.Fatal(server.ListenAndServe())
}
