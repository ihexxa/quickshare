package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	"github.com/ihexxa/quickshare/src/db"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/handlers/settings"
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
	t.Run("users API authz tests", func(t *testing.T) {
		// Login
		// SetPwd
		// Self
		// SetPreferences
		// IsAuthed
		// AddUser
		// AddUser
		// ListUsers
		// ForceSetPwd
		// ForceSetPwdOther
		// ForceSetPwdOtherAdmin
		// SetUserSelf
		// SetUserOthers
		// SetUserOthers
		// DelUser
		// AddRole
		// ListRoles
		// DelRole
		// Logout

		testUsersAPIs := func(desc, user string, pwd string, requireAuth bool, expectedCodes map[string]int) {
			newPwd := "12345"
			newRole := db.AdminRole
			newQuota := &db.Quota{
				SpaceLimit:         int64(2046),
				UploadSpeedLimit:   int(8 * 1024 * 1024),
				DownloadSpeedLimit: int(8 * 1024 * 1024),
			}
			tmpUser, tmpPwd, tmpRole := "tmpUser", "1234", "user"
			tmpAdmin, tmpAdminPwd := "tmpAdmin", "1234"
			tmpNewRole := "tmpNewRole"

			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}
			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				assertResp(t, resp, errs, 200, desc)
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			} else {
				resp, _, errs := cl.Login(user, pwd)
				assertResp(t, resp, errs, 403, desc)
			}

			resp, _, errs := cl.SetPwd(pwd, newPwd, token)
			assertResp(t, resp, errs, expectedCodes["SetPwd"], fmt.Sprintf("%s-%s", desc, "SetPwd"))
			// set back the password
			resp, _, errs = cl.SetPwd(newPwd, pwd, token)
			assertResp(t, resp, errs, expectedCodes["SetPwd"], fmt.Sprintf("%s-%s", desc, "SetPwd"))

			resp, selfResp, errs := cl.Self(token)
			assertResp(t, resp, errs, expectedCodes["Self"], fmt.Sprintf("%s-%s", desc, "Self"))

			prefer := selfResp.Preferences

			resp, _, errs = cl.SetPreferences(prefer, token)
			assertResp(t, resp, errs, expectedCodes["SetPreferences"], fmt.Sprintf("%s-%s", desc, "SetPreferences"))

			resp, _, errs = cl.IsAuthed(token)
			assertResp(t, resp, errs, expectedCodes["IsAuthed"], fmt.Sprintf("%s-%s", desc, "IsAuthed"))

			// user management
			resp, addUserResp, errs := cl.AddUser(tmpUser, tmpPwd, tmpRole, token)
			assertResp(t, resp, errs, expectedCodes["AddUser"], fmt.Sprintf("%s-%s", desc, "AddUser"))
			resp, addAdminResp, errs := cl.AddUser(tmpAdmin, tmpAdminPwd, db.AdminRole, token)
			assertResp(t, resp, errs, expectedCodes["AddUser"], fmt.Sprintf("%s-%s", desc, "AddUser"))

			resp, _, errs = cl.ListUsers(token)
			assertResp(t, resp, errs, expectedCodes["ListUsers"], fmt.Sprintf("%s-%s", desc, "ListUsers"))

			// TODO: the id here should be uint64
			tmpUserID := uint64(12345)
			var err error
			if addUserResp.ID != "" {
				tmpUserID, err = strconv.ParseUint(addUserResp.ID, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
			}
			userID := uint64(12345)
			if selfResp.ID != "" {
				userID, err = strconv.ParseUint(selfResp.ID, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
			}
			tmpAdminID := uint64(12345)
			if addAdminResp.ID != "" {
				tmpAdminID, err = strconv.ParseUint(addAdminResp.ID, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
			}

			resp, _, errs = cl.ForceSetPwd(selfResp.ID, newPwd, token)
			assertResp(t, resp, errs, expectedCodes["ForceSetPwd"], fmt.Sprintf("%s-%s", desc, "ForceSetPwd"))
			resp, _, errs = cl.ForceSetPwd(selfResp.ID, pwd, token)
			assertResp(t, resp, errs, expectedCodes["ForceSetPwd"], fmt.Sprintf("%s-%s", desc, "ForceSetPwdBack"))

			resp, _, errs = cl.ForceSetPwd(addUserResp.ID, newPwd, token)
			assertResp(t, resp, errs, expectedCodes["ForceSetPwdOther"], fmt.Sprintf("%s-%s", desc, "ForceSetPwdOther"))
			resp, _, errs = cl.ForceSetPwd(addUserResp.ID, pwd, token)
			assertResp(t, resp, errs, expectedCodes["ForceSetPwdOther"], fmt.Sprintf("%s-%s", desc, "ForceSetPwdOtherBack"))

			resp, _, errs = cl.ForceSetPwd(addAdminResp.ID, newPwd, token)
			assertResp(t, resp, errs, expectedCodes["ForceSetPwdOtherAdmin"], fmt.Sprintf("%s-%s", desc, "ForceSetPwdOtherAdmin"))

			// update self
			resp, _, errs = cl.SetUser(userID, newRole, newQuota, token)
			assertResp(t, resp, errs, expectedCodes["SetUserSelf"], fmt.Sprintf("%s-%s", desc, "SetUserSelf"))
			// update other users
			resp, _, errs = cl.SetUser(tmpUserID, db.AdminRole, newQuota, token)
			assertResp(t, resp, errs, expectedCodes["SetUserOthers"], fmt.Sprintf("%s-%s", desc, "SetUserOthers"))
			resp, _, errs = cl.SetUser(tmpAdminID, db.UserRole, newQuota, token)
			assertResp(t, resp, errs, expectedCodes["SetUserOthersAdmin"], fmt.Sprintf("%s-%s", desc, "SetUserOthersAdmin"))

			resp, _, errs = cl.DelUser(addUserResp.ID, token)
			assertResp(t, resp, errs, expectedCodes["DelUser"], fmt.Sprintf("%s-%s", desc, "DelUser"))
			resp, _, errs = cl.DelUser(addAdminResp.ID, token)
			assertResp(t, resp, errs, expectedCodes["DelUserAdmin"], fmt.Sprintf("%s-%s", desc, "DelUserAdmin"))

			// role management
			resp, _, errs = cl.AddRole(tmpNewRole, token)
			assertResp(t, resp, errs, expectedCodes["AddRole"], fmt.Sprintf("%s-%s", desc, "AddRole"))

			resp, _, errs = cl.ListRoles(token)
			assertResp(t, resp, errs, expectedCodes["ListRoles"], fmt.Sprintf("%s-%s", desc, "ListRoles"))

			resp, _, errs = cl.DelRole(tmpNewRole, token)
			assertResp(t, resp, errs, expectedCodes["DelRole"], fmt.Sprintf("%s-%s", desc, "DelRole"))

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				assertResp(t, resp, errs, 200, fmt.Sprintf("%s-%s", desc, "logout"))
			} else {
				resp, _, errs := cl.Logout(token)
				assertResp(t, resp, errs, 403, fmt.Sprintf("%s-%s", desc, "logout"))
			}
		}

		testUsersAPIs("admin user operations", "admin", "1234", true, map[string]int{
			"SetPwd":                200,
			"Self":                  200,
			"SetPreferences":        200,
			"IsAuthed":              200,
			"AddUser":               200,
			"ListUsers":             200,
			"ForceSetPwd":           403, // force setting admin's password is not allowed
			"ForceSetPwdOther":      200,
			"ForceSetPwdOtherAdmin": 403,
			"SetUserSelf":           200,
			"SetUserOthers":         200,
			"SetUserOthersAdmin":    200,
			"DelUser":               200,
			"DelUserAdmin":          200,
			"AddRole":               200,
			"ListRoles":             200,
			"DelRole":               200,
		})

		testUsersAPIs("user user operations", "user", "1234", true, map[string]int{
			"SetPwd":                200,
			"Self":                  200,
			"SetPreferences":        200,
			"IsAuthed":              200,
			"AddUser":               403,
			"ListUsers":             403,
			"ForceSetPwd":           403,
			"ForceSetPwdOther":      403,
			"ForceSetPwdOtherAdmin": 403,
			"SetUserSelf":           403,
			"SetUserOthers":         403,
			"SetUserOthersAdmin":    403,
			"DelUser":               403,
			"DelUserAdmin":          403,
			"AddRole":               403,
			"ListRoles":             403,
			"DelRole":               403,
		})

		testUsersAPIs("visitor user operations", "visitor", "", false, map[string]int{
			"SetPwd":                403,
			"Self":                  403,
			"SetPreferences":        403,
			"IsAuthed":              403,
			"AddUser":               403,
			"ListUsers":             403,
			"ForceSetPwd":           403,
			"ForceSetPwdOther":      403,
			"ForceSetPwdOtherAdmin": 403,
			"SetUserSelf":           403,
			"SetUserOthers":         403,
			"SetUserOthersAdmin":    403,
			"DelUser":               403,
			"DelUserAdmin":          403,
			"AddRole":               403,
			"ListRoles":             403,
			"DelRole":               403,
		})
	})

	t.Run("files operation API authz tests", func(t *testing.T) {
		testFolderOpPermission := func(desc, user string, pwd string, requireAuth bool, targetPaths []string, expectedCodes map[string]int) {
			// ListHome
			// List
			// ListTarget
			// Mkdir
			// Move

			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}

			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				assertResp(t, resp, errs, 200, desc)
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			filesCl := client.NewFilesClient(addr, token)

			resp, lhResp, errs := filesCl.ListHome()
			assertResp(t, resp, errs, expectedCodes["ListHome"], fmt.Sprintf("%s-%s", desc, "ListHome"))

			homePath := lhResp.Cwd
			if !requireAuth {
				homePath = "/"
			}
			testPath := filepath.Join(lhResp.Cwd, "test")
			newPath := filepath.Join(lhResp.Cwd, "test2")

			resp, _, errs = filesCl.List(homePath)
			assertResp(t, resp, errs, expectedCodes["List"], fmt.Sprintf("%s-%s", desc, "List"))

			for _, itemPath := range targetPaths {
				resp, _, errs = filesCl.List(itemPath)
				assertResp(t, resp, errs, expectedCodes["ListTarget"], fmt.Sprintf("%s-%s", desc, "ListTarget"))
			}

			resp, _, errs = filesCl.Mkdir(testPath)
			assertResp(t, resp, errs, expectedCodes["Mkdir"], fmt.Sprintf("%s-%s", desc, "Mkdir"))

			for _, itemPath := range targetPaths {
				resp, _, errs = filesCl.Mkdir(filepath.Join(itemPath, "test"))
				assertResp(t, resp, errs, expectedCodes["MkdirTarget"], fmt.Sprintf("%s-%s", desc, "MkdirTarget"))
			}

			resp, _, errs = filesCl.Move(testPath, newPath)
			assertResp(t, resp, errs, expectedCodes["Move"], fmt.Sprintf("%s-%s", desc, "Move"))

			for i, srcDir := range targetPaths {
				srcPath := filepath.Join(srcDir, "test")
				dstPath := filepath.Join(homePath, fmt.Sprintf("folder_%d", i))

				resp, _, errs = filesCl.Move(srcPath, dstPath)
				assertResp(t, resp, errs, expectedCodes["MoveFrom"], fmt.Sprintf("%s-%s", desc, "MoveFrom"))
			}

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				assertResp(t, resp, errs, 200, desc)
			}
		}

		targetPaths := []string{
			"/",
			"admin2/",
			"admin2/files",
			"user2/",
			"user2/files",
		}

		testFolderOpPermission("admin folder operations", "admin", "1234", true, targetPaths, map[string]int{
			"ListHome":    200,
			"List":        200,
			"ListTarget":  200,
			"Mkdir":       200,
			"MkdirTarget": 200,
			"Move":        200,
			"MoveFrom":    200,
		})
		testFolderOpPermission("user folder operations", "user", "1234", true, targetPaths, map[string]int{
			"ListHome":    200,
			"List":        200,
			"ListTarget":  403,
			"Mkdir":       200,
			"MkdirTarget": 403,
			"Move":        200,
			"MoveFrom":    403,
		})
		testFolderOpPermission("visitor folder operations", "visitor", "", false, targetPaths, map[string]int{
			"ListHome":    403,
			"List":        403,
			"ListTarget":  403,
			"Mkdir":       403,
			"MkdirTarget": 403,
			"Move":        403,
			"MoveFrom":    403,
		})

		uploadSample := func() {
			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}

			resp, _, errs := cl.Login("user2", "1234")
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if resp.StatusCode != 200 {
				t.Fatal(resp.StatusCode)
			}
			token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			assertUploadOK(t, "user2/files/upload", "101", addr, token)
		}
		uploadSample()

		// test file operations in targetPath with targetFile
		testFileOpPermission := func(desc, user string, pwd string, requireAuth bool, targetPath, targetFile string, expectedCodes map[string]int) {
			// Create
			// UploadChunk
			// UploadStatus
			// Metadata
			// Move
			// Download
			// Delete

			// ListUploadings
			// DelUploading

			// GenerateHash

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

			resp, lhResp, errs := filesCl.ListHome()
			assertResp(t, resp, errs, expectedCodes["ListHome"], fmt.Sprintf("%s-%s", desc, "ListHome"))
			resp, _, errs = filesCl.List(targetPath)
			assertResp(t, resp, errs, expectedCodes["ListTarget"], fmt.Sprintf("%s-%s", desc, "ListTarget"))

			fileContent := []byte("01010")
			fileSize := int64(len(fileContent))
			base64Content := base64.StdEncoding.EncodeToString([]byte(fileContent))
			filePath := filepath.Join(lhResp.Cwd, "old")
			homeNewPath := filepath.Join(lhResp.Cwd, "new")
			targetPathFile := filepath.Join(targetPath, user)

			resp, _, errs = filesCl.Create(filePath, fileSize)
			assertResp(t, resp, errs, expectedCodes["Create"], fmt.Sprintf("%s-%s", desc, "Create1"))
			resp, _, errs = filesCl.Create(targetPathFile, fileSize)
			assertResp(t, resp, errs, expectedCodes["CreateTarget"], fmt.Sprintf("%s-%s", desc, "CreateTarget"))

			resp, _, errs = filesCl.UploadStatus(filePath)
			assertResp(t, resp, errs, expectedCodes["UploadStatus"], fmt.Sprintf("%s-%s", desc, "UploadStatus"))
			resp, _, errs = filesCl.UploadStatus(targetPathFile)
			assertResp(t, resp, errs, expectedCodes["UploadStatusTarget"], fmt.Sprintf("%s-%s", desc, "UploadStatusTarget"))

			resp, _, errs = filesCl.ListUploadings()
			assertResp(t, resp, errs, expectedCodes["ListUploadings"], fmt.Sprintf("%s-%s", desc, "ListUploadings"))

			resp, _, errs = filesCl.DelUploading(filePath)
			assertResp(t, resp, errs, expectedCodes["DelUploading"], fmt.Sprintf("%s-%s", desc, "DelUploading"))
			resp, _, errs = filesCl.DelUploading(targetPathFile)
			assertResp(t, resp, errs, expectedCodes["DelUploadingTarget"], fmt.Sprintf("%s-%s", desc, "DelUploadingTarget"))

			// create again
			resp, _, errs = filesCl.Create(filePath, fileSize)
			assertResp(t, resp, errs, expectedCodes["Create"], fmt.Sprintf("%s-%s", desc, "Create2"))
			resp, _, errs = filesCl.Create(targetPathFile, fileSize)
			assertResp(t, resp, errs, expectedCodes["CreateTarget"], fmt.Sprintf("%s-%s", desc, "CreateTarget"))

			resp, _, errs = filesCl.UploadChunk(filePath, base64Content, 0)
			assertResp(t, resp, errs, expectedCodes["UploadChunk"], fmt.Sprintf("%s-%s", desc, "UploadChunk"))
			resp, _, errs = filesCl.UploadChunk(targetPathFile, base64Content, 0)
			assertResp(t, resp, errs, expectedCodes["UploadChunkTarget"], fmt.Sprintf("%s-%s", desc, "UploadChunkTarget"))

			resp, _, errs = filesCl.Metadata(filePath)
			assertResp(t, resp, errs, expectedCodes["Metadata"], fmt.Sprintf("%s-%s", desc, "Metadata"))
			resp, _, errs = filesCl.Metadata(targetPath)
			assertResp(t, resp, errs, expectedCodes["MetadataTarget"], fmt.Sprintf("%s-%s", desc, "MetadataTarget"))

			resp, _, errs = filesCl.GenerateHash(filePath)
			assertResp(t, resp, errs, expectedCodes["GenerateHash"], fmt.Sprintf("%s-%s", desc, "GenerateHash"))
			resp, _, errs = filesCl.GenerateHash(targetPathFile)
			assertResp(t, resp, errs, expectedCodes["GenerateHashTarget"], fmt.Sprintf("%s-%s", desc, "GenerateHashTarget"))

			resp, _, errs = filesCl.Download(filePath, map[string]string{})
			assertResp(t, resp, errs, expectedCodes["Download"], fmt.Sprintf("%s-%s", desc, "Download"))
			resp, _, errs = filesCl.Download(targetPathFile, map[string]string{})
			assertResp(t, resp, errs, expectedCodes["DownloadTarget"], fmt.Sprintf("%s-%s", desc, "DownloadTarget"))

			if targetFile != "" {
				resp, _, errs = filesCl.Download(targetFile, map[string]string{})
				assertResp(t, resp, errs, expectedCodes["DownloadTargetFile"], fmt.Sprintf("%s-%s", desc, "DownloadTarget"))
			}

			resp, _, errs = filesCl.Move(filePath, homeNewPath)
			assertResp(t, resp, errs, expectedCodes["Move"], fmt.Sprintf("%s-%s", desc, "Move"))
			// move target is already tested

			resp, _, errs = filesCl.Delete(homeNewPath)
			assertResp(t, resp, errs, expectedCodes["Delete"], fmt.Sprintf("%s-%s", desc, "Delete"))
			resp, _, errs = filesCl.Delete(targetPathFile)
			assertResp(t, resp, errs, expectedCodes["DeleteTarget"], fmt.Sprintf("%s-%s", desc, "DeleteTarget"))

			if requireAuth {
				resp, _, errs := cl.Logout(token)
				assertResp(t, resp, errs, 200, desc)
			}
		}

		testFileOpPermission("admin file operations", "admin", "1234", true, "user2/files", "", map[string]int{
			"ListHome":           200,
			"ListTarget":         200,
			"Create":             200,
			"CreateTarget":       200,
			"ListUploadings":     200,
			"DelUploading":       200,
			"DelUploadingTarget": 200,
			"UploadChunk":        200,
			"UploadChunkTarget":  200,
			"UploadStatus":       200,
			"UploadStatusTarget": 200,
			"Metadata":           200,
			"MetadataTarget":     200,
			"GenerateHash":       200,
			"GenerateHashTarget": 200,
			"Move":               200,
			"Download":           200,
			"DownloadTarget":     200,
			"Delete":             200,
			"DeleteTarget":       200,
		})
		testFileOpPermission("user file operations", "user", "1234", true, "user2/files", "", map[string]int{
			"ListHome":           200,
			"ListTarget":         403,
			"Create":             200,
			"CreateTarget":       403,
			"ListUploadings":     200,
			"DelUploading":       200,
			"DelUploadingTarget": 403,
			"UploadChunk":        200,
			"UploadChunkTarget":  403,
			"UploadStatus":       200,
			"UploadStatusTarget": 403,
			"Metadata":           200,
			"MetadataTarget":     403,
			"GenerateHash":       200,
			"GenerateHashTarget": 403,
			"Move":               200,
			"Download":           200,
			"DownloadTarget":     403,
			"Delete":             200,
			"DeleteTarget":       403,
		})
		testFileOpPermission("visitor file operations", "visitor", "", false, "user2/files", "", map[string]int{
			"ListHome":           403,
			"ListTarget":         403,
			"Create":             403,
			"CreateTarget":       403,
			"ListUploadings":     403,
			"DelUploading":       403,
			"DelUploadingTarget": 403,
			"UploadChunk":        403,
			"UploadChunkTarget":  403,
			"UploadStatus":       403,
			"UploadStatusTarget": 403,
			"Metadata":           403,
			"MetadataTarget":     403,
			"GenerateHash":       403,
			"GenerateHashTarget": 403,
			"Move":               403,
			"Download":           403,
			"DownloadTarget":     403,
			"Delete":             403,
			"DeleteTarget":       403,
		})

		// sharing permission tests
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

		testFileOpPermission("user file operations in sharing", "user", "1234", true, "share/files", "share/files/share", map[string]int{
			"ListHome":           200,
			"ListTarget":         200,
			"Create":             200,
			"CreateTarget":       403,
			"ListUploadings":     200,
			"DelUploading":       200,
			"DelUploadingTarget": 403,
			"UploadChunk":        200,
			"UploadChunkTarget":  403,
			"UploadStatus":       200,
			"UploadStatusTarget": 403,
			"Metadata":           200,
			"MetadataTarget":     403,
			"GenerateHash":       200,
			"GenerateHashTarget": 403,
			"Move":               200,
			"Download":           200,
			"DownloadTarget":     404, // uploading file to sharing folder is not allowd
			"DownloadTargetFile": 200,
			"Delete":             200,
			"DeleteTarget":       403,
		})

		testShareOpPermission := func(user string, pwd string, requireAuth bool, targetPath string, expectedCodes map[string]int) {
			// AddSharing
			// DelSharing
			// IsSharing
			// ListSharings deprecated
			// ListSharingIDs
			// GetSharingDir

			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}
			homePath := "/"
			desc := user

			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				assertResp(t, resp, errs, 200, desc)
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			filesCl := client.NewFilesClient(addr, token)

			if requireAuth {
				resp, lhResp, errs := filesCl.ListHome()
				assertResp(t, resp, errs, 200, desc)
				homePath = lhResp.Cwd
			}

			resp, _, errs := filesCl.AddSharing(homePath)
			assertResp(t, resp, errs, expectedCodes["AddSharing"], desc)

			resp, _, errs = filesCl.AddSharing(targetPath)
			assertResp(t, resp, errs, expectedCodes["AddSharingTarget"], desc)

			resp, _, errs = filesCl.IsSharing(homePath)
			assertResp(t, resp, errs, expectedCodes["IsSharing"], desc)

			resp, _, errs = filesCl.IsSharing(targetPath)
			assertResp(t, resp, errs, expectedCodes["IsSharingTarget"], desc)

			resp, lsResp, errs := filesCl.ListSharingIDs()
			assertResp(t, resp, errs, expectedCodes["ListSharingIDs"], desc)

			shareID := ""
			if len(lsResp.IDs) == 0 && requireAuth {
				t.Fatalf("sharing is not added: %s", desc)
			} else if len(lsResp.IDs) > 0 {
				for _, id := range lsResp.IDs {
					shareID = id
					break
				}
			}

			// TODO: visitor accessing is not tested
			resp, _, errs = filesCl.GetSharingDir(shareID)
			assertResp(t, resp, errs, expectedCodes["GetSharingDir"], desc)

			resp, _, errs = filesCl.DelSharing(homePath)
			assertResp(t, resp, errs, expectedCodes["DelSharing"], desc)

			resp, _, errs = filesCl.DelSharing(targetPath)
			assertResp(t, resp, errs, expectedCodes["DelSharingTarget"], desc)
		}

		testShareOpPermission("admin", "1234", true, "user2/files", map[string]int{
			"AddSharing":       200,
			"AddSharingTarget": 200,
			"IsSharing":        200,
			"IsSharingTarget":  200, // sharing is added by admin
			"ListSharingIDs":   200,
			"GetSharingDir":    200,
			"DelSharing":       200,
			"DelSharingTarget": 200, // it returns 200 even it is not in sharing
		})

		testShareOpPermission("user", "1234", true, "user2/files", map[string]int{
			"AddSharing":       200,
			"AddSharingTarget": 403,
			"IsSharing":        200,
			"IsSharingTarget":  404, // sharing is deleted in previous test...
			"ListSharingIDs":   200,
			"GetSharingDir":    200,
			"DelSharing":       200,
			"DelSharingTarget": 403,
		})

		testShareOpPermission("visitor", "", false, "user2/files", map[string]int{
			"AddSharing":       403,
			"AddSharingTarget": 403,
			"IsSharing":        404,
			"IsSharingTarget":  404, // sharing is deleted in previous test...
			"ListSharingIDs":   403,
			"GetSharingDir":    400,
			"DelSharing":       403,
			"DelSharingTarget": 403,
		})
	})

	t.Run("Settings API Permissions", func(t *testing.T) {
		testSettingsOpPermission := func(user string, pwd string, requireAuth bool, expectedCodes map[string]int) {
			// Health
			// GetClientCfg
			// SetClientCfg
			// ReportErrors

			cl := client.NewSingleUserClient(addr)
			token := &http.Cookie{}
			desc := user
			errReports := &settings.ClientErrorReports{
				Reports: []*settings.ClientErrorReport{
					&settings.ClientErrorReport{
						Report:  "report1",
						Version: "v1",
					},
					&settings.ClientErrorReport{
						Report:  "report2",
						Version: "v2",
					},
				},
			}

			if requireAuth {
				resp, _, errs := cl.Login(user, pwd)
				assertResp(t, resp, errs, 200, desc)
				token = client.GetCookie(resp.Cookies(), q.TokenCookie)
			}

			settingsCl := client.NewSettingsClient(addr)

			resp, _, errs := settingsCl.Health()
			assertResp(t, resp, errs, expectedCodes["Health"], fmt.Sprintf("%s-%s", desc, "Health"))

			resp, gccResp, errs := settingsCl.GetClientCfg(token)
			assertResp(t, resp, errs, expectedCodes["GetClientCfg"], fmt.Sprintf("%s-%s", desc, "GetClientCfg"))

			clientCfgMsg := gccResp
			clientCfgMsg.ClientCfg.SiteName = "new name"

			resp, _, errs = settingsCl.SetClientCfg(clientCfgMsg, token)
			assertResp(t, resp, errs, expectedCodes["SetClientCfg"], fmt.Sprintf("%s-%s", desc, "SetClientCfg"))

			resp, _, errs = settingsCl.ReportErrors(errReports, token)
			assertResp(t, resp, errs, expectedCodes["ReportErrors"], fmt.Sprintf("%s-%s", desc, "ReportErrors"))
		}

		testSettingsOpPermission("admin", "1234", true, map[string]int{
			"Health":       200,
			"GetClientCfg": 200,
			"SetClientCfg": 200,
			"ReportErrors": 200,
		})

		testSettingsOpPermission("user", "1234", true, map[string]int{
			"Health":       200,
			"GetClientCfg": 200,
			"SetClientCfg": 403,
			"ReportErrors": 200,
		})

		testSettingsOpPermission("visitor", "", false, map[string]int{
			"Health":       200,
			"GetClientCfg": 200,
			"SetClientCfg": 403,
			"ReportErrors": 403,
		})
	})
}
