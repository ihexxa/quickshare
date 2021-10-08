package server

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/ihexxa/gocfg"
	"github.com/ihexxa/quickshare/src/client"
	fspkg "github.com/ihexxa/quickshare/src/fs"
	"github.com/ihexxa/quickshare/src/db/userstore"
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

func setUpEnv(t *testing.T, rootPath string, adminName, adminPwd string) {
	os.Setenv("DEFAULTADMIN", adminName)
	os.Setenv("DEFAULTADMINPWD", adminPwd)
	os.RemoveAll(rootPath)
	err := os.MkdirAll(rootPath, 0700)
	if err != nil {
		t.Fatal(err)
	}
}

func getUserName(id int) string {
	return fmt.Sprintf("user_%d", id)
}

func addUsers(t *testing.T, addr, userPwd string, userCount int, adminToken *http.Cookie) map[string]string {
	usersCl := client.NewSingleUserClient(addr)
	users := map[string]string{}
	for i := range make([]int, userCount) {
		userName := getUserName(i)

		resp, adResp, errs := usersCl.AddUser(userName, userPwd, userstore.UserRole, adminToken)
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal("failed to add user")
		}

		users[userName] = adResp.ID
	}

	return users
}

func isServerReady(addr string) bool {
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
	reader, id, err := fs.GetFileReader(filePath)
	if err != nil {
		return false, err
	}
	defer func() {
		err = fs.CloseReader(fmt.Sprint(id))
		if err != nil {
			fmt.Println(err)
		}
	}()

	gotContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, err
	}

	return string(gotContent) == expectedContent, nil
}

func assertUploadOK(t *testing.T, filePath, content, addr string, token *http.Cookie) bool {
	cl := client.NewFilesClient(addr, token)

	fileSize := int64(len([]byte(content)))
	res, _, errs := cl.Create(filePath, fileSize)
	if len(errs) > 0 {
		t.Fatal(errs)
		return false
	} else if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
		return false
	}

	base64Content := base64.StdEncoding.EncodeToString([]byte(content))
	res, _, errs = cl.UploadChunk(filePath, base64Content, 0)
	if len(errs) > 0 {
		t.Fatal(errs)
		return false
	} else if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
		return false
	}

	return true
}

func assertDownloadOK(t *testing.T, filePath, content, addr string, token *http.Cookie) bool {
	var (
		res      *http.Response
		body     string
		errs     []error
		fileSize = int64(len([]byte(content)))
	)

	cl := client.NewFilesClient(addr, token)

	rd := rand.Intn(3)
	switch rd {
	case 0:
		res, body, errs = cl.Download(filePath, map[string]string{})
	case 1:
		res, body, errs = cl.Download(filePath, map[string]string{
			"Range": fmt.Sprintf("bytes=0-%d", fileSize-1),
		})
	case 2:
		res, body, errs = cl.Download(filePath, map[string]string{
			"Range": fmt.Sprintf("bytes=0-%d, %d-%d", (fileSize-1)/2, (fileSize-1)/2+1, fileSize-1),
		})
	}

	fileName := path.Base(filePath)
	contentDispositionHeader := res.Header.Get("Content-Disposition")
	if len(errs) > 0 {
		t.Error(errs)
		return false
	}
	if res.StatusCode != 200 && res.StatusCode != 206 {
		t.Error(res.StatusCode)
		return false
	}
	if contentDispositionHeader != fmt.Sprintf(`attachment; filename="%s"`, fileName) {
		t.Errorf("incorrect Content-Disposition header: %s", contentDispositionHeader)
		return false
	}

	switch rd {
	case 0:
		if body != content {
			t.Errorf("body not equal got(%s) expect(%s)\n", body, content)
			return false
		}
	case 1:
		if body[2:] != content { // body returned by gorequest contains the first CRLF
			t.Errorf("body not equal got(%s) expect(%s)\n", body[2:], content)
			return false
		}
	default:
		body = body[2:] // body returned by gorequest contains the first CRLF
		realBody := ""
		boundaryEnd := strings.Index(body, "\r\n")
		boundary := body[0:boundaryEnd]
		bodyParts := strings.Split(body, boundary)

		for i, bodyPart := range bodyParts {
			if i == 0 || i == len(bodyParts)-1 {
				continue
			}
			start := strings.Index(bodyPart, "\r\n\r\n")

			fmt.Printf("<%s>", bodyPart[start+4:len(bodyPart)-2]) // ignore the last CRLF
			realBody += bodyPart[start+4 : len(bodyPart)-2]
		}
		if realBody != content {
			t.Errorf("multi body not equal got(%s) expect(%s)\n", realBody, content)
			return false
		}
	}

	return true
}
