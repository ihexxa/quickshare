package server

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/db/userstore"
	q "github.com/ihexxa/quickshare/src/handlers"
)

func TestPermissions(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	rootPath := "tmpTestData"
	config := `{
		"users": {
			"enableAuth": true,
			"minUserNameLen": 2,
			"minPwdLen": 4,
			"captchaEnabled": false,
			"uploadSpeedLimit": 409600,
			"downloadSpeedLimit": 409600,
			"spaceLimit": 1000,
			"limiterCapacity": 1000,
			"limiterCyc": 1000,
			"predefinedUsers": [				
				{
					"name": "admin",
					"pwd": "1234",
					"role": "admin"
				},
				{
					"name": "admin2",
					"pwd": "1234",
					"role": "admin"
				},
				{
					"name": "user",
					"pwd": "1234",
					"role": "user"
				},
				{
					"name": "user2",
					"pwd": "1234",
					"role": "user"
				},
				{
					"name": "share",
					"pwd": "1234",
					"role": "user"
				}
			]
		},
		"server": {
			"debug": true,
			"host": "127.0.0.1"
		},
		"fs": {
			"root": "tmpTestData"
		},
		"db": {
			"dbPath": "tmpTestData/quickshare"
		}
	}`

	adminName := "qs"
	adminPwd := "quicksh@re"
	setUpEnv(t, rootPath, adminName, adminPwd)
	defer os.RemoveAll(rootPath)

	srv := startTestServer(config)
	defer srv.Shutdown()
	if !isServerReady(addr) {
		t.Fatal("fail to start server")
	}

	// tests only check the status code for checking permission
	t.Run("Users API Permissions", func(t *testing.T) {
		testUsersAPIs := func(user string, pwd string, requireAuth bool, expectedCodes map[string]int) {
			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}
			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			expectedCode := expectedCodes["SetPwd"]
			newPwd := "12345"
			resp, _, errs := cl.SetPwd(pwd, newPwd, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}
			// set back the pwd
			resp, _, errs = cl.SetPwd(newPwd, pwd, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}

			expectedCode = expectedCodes["Self"]
			resp, selfResp, errs := cl.Self(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}

			expectedCode = expectedCodes["SetPreferences"]
			prefer := selfResp.Preferences
			resp, _, errs = cl.SetPreferences(prefer, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(resp.StatusCode)
			}

			expectedCode = expectedCodes["IsAuthed"]
			resp, _, errs = cl.IsAuthed(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			// test user operations
			expectedCode = expectedCodes["AddUser"]
			tmpUser, tmpPwd, tmpRole := "tmpUser", "1234", "admin"
			resp, addUserResp, errs := cl.AddUser(tmpUser, tmpPwd, tmpRole, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			expectedCode = expectedCodes["ListUsers"]
			resp, _, errs = cl.ListUsers(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			// TODO: the id here should be uint64
			tmpUserID := uint64(0)
			var err error
			if addUserResp.ID != "" {
				tmpUserID, err = strconv.ParseUint(addUserResp.ID, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
			}
			userID := uint64(0)
			if selfResp.ID != "" {
				userID, err = strconv.ParseUint(selfResp.ID, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
			}

			newRole := userstore.AdminRole
			newQuota := &userstore.Quota{
				SpaceLimit:         int64(2046),
				UploadSpeedLimit:   int(8 * 1024 * 1024),
				DownloadSpeedLimit: int(8 * 1024 * 1024),
			}
			// update self
			expectedCode = expectedCodes["SetUserSelf"]
			resp, _, errs = cl.SetUser(userID, newRole, newQuota, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}
			// update other users
			expectedCode = expectedCodes["SetUserOthers"]
			resp, _, errs = cl.SetUser(tmpUserID, userstore.AdminRole, newQuota, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}
			resp, _, errs = cl.SetUser(0, userstore.UserRole, newQuota, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			expectedCode = expectedCodes["DelUser"]
			resp, _, errs = cl.DelUser(addUserResp.ID, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			// test role operations
			expectedCode = expectedCodes["AddRole"]
			tmpNewRole := "tmpNewRole"
			resp, _, errs = cl.AddRole(tmpNewRole, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			expectedCode = expectedCodes["ListRoles"]
			resp, _, errs = cl.ListRoles(token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			expectedCode = expectedCodes["DelRole"]
			resp, _, errs = cl.DelRole(tmpNewRole, token)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatalf("%s %d %d", user, expectedCode, resp.StatusCode)
			}

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
			}
		}

		testUsersAPIs("admin", "1234", true, map[string]int{
			"SetPwd":         200,
			"Self":           200,
			"SetPreferences": 200,
			"IsAuthed":       200,
			"AddUser":        200,
			"ListUsers":      200,
			"SetUserSelf":    200,
			"SetUserOthers":  200,
			"SetOtherUser":   200,
			"DelUser":        200,
			"AddRole":        200,
			"ListRoles":      200,
			"DelRole":        200,
		})

		testUsersAPIs("user", "1234", true, map[string]int{
			"SetPwd":         200,
			"Self":           200,
			"SetPreferences": 200,
			"IsAuthed":       200,
			"AddUser":        403,
			"ListUsers":      403,
			"SetUserSelf":    403,
			"SetUserOthers":  403,
			"DelUser":        403,
			"AddRole":        403,
			"ListRoles":      403,
			"DelRole":        403,
		})

		testUsersAPIs("visitor", "", false, map[string]int{
			"SetPwd":         403,
			"Self":           403,
			"SetPreferences": 403,
			"IsAuthed":       403,
			"AddUser":        403,
			"ListUsers":      403,
			"SetUserSelf":    403,
			"SetUserOthers":  403,
			"DelUser":        403,
			"AddRole":        403,
			"ListRoles":      403,
			"DelRole":        403,
		})
	})

	t.Run("Files operation API Permissions", func(t *testing.T) {
		// ListUploadings(c *gin.Context) {
		// DelUploading(c *gin.Context) {

		// AddSharing(c *gin.Context) {
		// DelSharing(c *gin.Context) {
		// IsSharing(c *gin.Context) {
		// ListSharings(c *gin.Context) {
		// ListSharingIDs(c *gin.Context) {

		// GenerateHash(c *gin.Context) {
		// GetSharingDir(c *gin.Context) {

		testFolderOpPermission := func(user string, pwd string, requireAuth bool, expectedCodes map[string]int) {
			// List(c *gin.Context) {
			// ListHome(c *gin.Context) {
			// Mkdir(c *gin.Context) {
			// Move(c *gin.Context) {

			// Create(c *gin.Context) {
			// UploadChunk(c *gin.Context) {
			// UploadStatus(c *gin.Context) {
			// Metadata(c *gin.Context) {
			// Move(c *gin.Context) {
			// Download(c *gin.Context) {
			// GetStreamReader(userID uint64, fd io.Reader) (io.ReadCloser, error) {
			// Delete(c *gin.Context) {

			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}
			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			filesCl := client.NewFilesClient(addr, token)

			expectedCode := expectedCodes["ListHome"]
			resp, lhResp, errs := filesCl.ListHome()
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["List"]
			homePath := lhResp.Cwd
			if !requireAuth {
				homePath = "/"
			}
			resp, _, errs = filesCl.List(homePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["ListPaths"]
			for _, itemPath := range []string{
				"/",
				"admin/",
				"admin/files",
				"user2/",
				"user2/files",
			} {
				resp, _, errs = filesCl.List(itemPath)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != expectedCode {
					t.Fatal(user, resp.StatusCode, expectedCode)
				}
			}

			expectedCode = expectedCodes["Mkdir"]
			testPath := filepath.Join(lhResp.Cwd, "test")
			resp, _, errs = filesCl.Mkdir(testPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["Move"]
			newPath := filepath.Join(lhResp.Cwd, "test2")
			resp, _, errs = filesCl.Move(testPath, newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(user, resp.StatusCode, expectedCode)
				}
			}
		}

		testFolderOpPermission("admin", "1234", true, map[string]int{
			"ListHome":  200,
			"List":      200,
			"ListPaths": 200,
			"Mkdir":     200,
			"Move":      200,
		})
		testFolderOpPermission("user", "1234", true, map[string]int{
			"ListHome":  200,
			"List":      200,
			"ListPaths": 403,
			"Mkdir":     200,
			"Move":      200,
		})
		testFolderOpPermission("visitor", "", false, map[string]int{
			"ListHome":  403,
			"List":      403,
			"ListPaths": 403,
			"Mkdir":     403,
			"Move":      403,
		})

		testFileOpPermission := func(user string, pwd string, requireAuth bool, targetPath, targetFile string, expectedCodes map[string]int) {
			// Create(c *gin.Context) {
			// UploadChunk(c *gin.Context) {
			// UploadStatus(c *gin.Context) {
			// Metadata(c *gin.Context) {
			// Move(c *gin.Context) {
			// Download(c *gin.Context) {
			// Delete(c *gin.Context) {

			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}
			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(resp.StatusCode)
				}
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			expectedCode := expectedCodes["ListHome"]
			filesCl := client.NewFilesClient(addr, token)
			resp, _, errs := filesCl.ListHome()
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			fileContent := []byte("01010")
			filePath := filepath.Join(targetPath, "old")
			fileSize := int64(len(fileContent))
			expectedCode = expectedCodes["Create"]
			resp, _, errs = filesCl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["UploadStatus"]
			resp, _, errs = filesCl.UploadStatus(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["UploadChunk"]
			base64Content := base64.StdEncoding.EncodeToString([]byte(fileContent))
			resp, _, errs = filesCl.UploadChunk(filePath, base64Content, 0)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["Metadata"]
			resp, _, errs = filesCl.Metadata(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["MetadataTarget"]
			resp, _, errs = filesCl.Metadata(targetPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["Download"]
			resp, _, errs = filesCl.Download(filePath, map[string]string{})
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			if targetFile != "" {
				expectedCode = expectedCodes["DownloadTarget"]
				resp, _, errs = filesCl.Download(targetFile, map[string]string{})
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != expectedCode {
					t.Fatal(user, resp.StatusCode, expectedCode)
				}
			}

			expectedCode = expectedCodes["Move"]
			newPath := filepath.Join(targetPath, "new")
			resp, _, errs = filesCl.Move(filePath, newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			expectedCode = expectedCodes["Delete"]
			resp, _, errs = filesCl.Delete(newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != expectedCode {
				t.Fatal(user, resp.StatusCode, expectedCode)
			}

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if resp.StatusCode != 200 {
					t.Fatal(user, resp.StatusCode, expectedCode)
				}
			}
		}

		testFileOpPermission("admin", "1234", true, "admin/files", "", map[string]int{
			"ListHome":       200,
			"Create":         200,
			"UploadChunk":    200,
			"UploadStatus":   200,
			"Metadata":       200,
			"MetadataTarget": 200,
			"Move":           200,
			"Download":       200,
			"Delete":         200,
		})
		testFileOpPermission("user", "1234", true, "user/files", "", map[string]int{
			"ListHome":       200,
			"Create":         200,
			"UploadChunk":    200,
			"UploadStatus":   200,
			"Metadata":       200,
			"MetadataTarget": 200,
			"Move":           200,
			"Download":       200,
			"Delete":         200,
		})
		testFileOpPermission("visitor", "", false, "user/files", "", map[string]int{
			"ListHome":       403,
			"Create":         403,
			"UploadChunk":    403,
			"UploadStatus":   403,
			"Metadata":       403,
			"MetadataTarget": 403,
			"Move":           403,
			"Download":       403,
			"Delete":         403,
		})
		testFileOpPermission("admin", "1234", true, "user2/files", "", map[string]int{
			"ListHome":       200,
			"Create":         200,
			"UploadChunk":    200,
			"UploadStatus":   200,
			"Metadata":       200,
			"MetadataTarget": 200,
			"Move":           200,
			"Download":       200,
			"Delete":         200,
		})
		testFileOpPermission("user", "1234", true, "user2/files", "", map[string]int{
			"ListHome":       200,
			"Create":         403,
			"UploadChunk":    403,
			"UploadStatus":   403,
			"Metadata":       403,
			"MetadataTarget": 403,
			"Move":           403,
			"Download":       403,
			"Delete":         403,
		})

		// test sharing permission
		enableSharing := func() {
			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}

			resp, _, errs := cl.Login("share", "1234")
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}
			token = client.GetCookie(resp.Cookies(), q.TokenCookie)

			filesCl := client.NewFilesClient(addr, token)
			resp, _, errs = filesCl.AddSharing("share/files")
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}

			assertUploadOK(t, "share/files/share", "101", addr, token)
		}
		enableSharing()

		testFileOpPermission("user", "1234", true, "share/files", "share/files/share", map[string]int{
			"ListHome":       200,
			"Create":         403,
			"UploadChunk":    403,
			"UploadStatus":   403,
			"Metadata":       403,
			"MetadataTarget": 403,
			"Move":           403,
			"Download":       404,
			"DownloadTarget": 200,
			"Delete":         403,
			// List is not tested
		})
	})

	t.Run("Settings API Permissions", func(t *testing.T) {

	})
}
