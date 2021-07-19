package server

import (
	"io/ioutil"
	// "path"
	"time"

	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/client"
	fspkg "github.com/ihexxa/quickshare/src/fs"
)

func startTestServer(config string) *Server {
	defaultCfg, err := DefaultConfig()
	if err != nil {
		panic(err)
	}

	cfg, err := gocfg.New(NewConfig()).
		Load(
			gocfg.JSONStr(defaultCfg),
			gocfg.JSONStr(config),
		)
	if err != nil {
		panic(err)
	}

	srv, err := NewServer(cfg)
	if err != nil {
		panic(err)
	}

	go srv.Start()
	return srv
}

func waitForReady(addr string) bool {
	retry := 20
	setCl := client.NewSettingsClient(addr)

	for retry > 0 {
		_, _, errs := setCl.Health()
		if len(errs) > 0 {
			time.Sleep(100 * time.Millisecond)
		} else {
			return true
		}
		retry--
	}

	return false
}

func compareFileContent(fs fspkg.ISimpleFS, uid, filePath string, expectedContent string) (bool, error) {
	reader, err := fs.GetFileReader(filePath)
	if err != nil {
		return false, err
	}

	gotContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, err
	}

	return string(gotContent) == expectedContent, nil
}
