package fileshdr

import (
	"context"
	"errors"
	"fmt"

	"github.com/ihexxa/quickshare/src/db"
	q "github.com/ihexxa/quickshare/src/handlers"
	"golang.org/x/crypto/bcrypt"
)

func (h *FileHandlers) Init(ctx context.Context, adminName string) (string, error) {
	var err error

	fsPath := q.FsRootPath(adminName, "")
	if err = h.deps.FS().MkdirAll(fsPath); err != nil {
		return "", err
	}
	uploadFolder := q.UploadFolder(adminName)
	if err = h.deps.FS().MkdirAll(uploadFolder); err != nil {
		return "", err
	}

	usersInterface, ok := h.cfg.Slice("Users.PredefinedUsers")
	spaceLimit := int64(h.cfg.IntOr("Users.SpaceLimit", 100*1024*1024))
	uploadSpeedLimit := h.cfg.IntOr("Users.UploadSpeedLimit", 100*1024)
	downloadSpeedLimit := h.cfg.IntOr("Users.DownloadSpeedLimit", 100*1024)
	if downloadSpeedLimit < q.DownloadChunkSize {
		return "", fmt.Errorf("download speed limit can not be lower than chunk size: %d", q.DownloadChunkSize)
	}
	if ok {
		userCfgs, ok := usersInterface.([]*db.UserCfg)
		if !ok {
			return "", fmt.Errorf("predefined user is invalid: %s", err)
		}
		for _, userCfg := range userCfgs {
			_, err := h.deps.Users().GetUserByName(ctx, userCfg.Name)
			if err != nil {
				if errors.Is(err, db.ErrUserNotFound) {
					// no op, need initing
				} else {
					return "", err
				}
			} else {
				h.deps.Log().Warn("warning: users exists, skip initing(%s)", userCfg.Name)
				continue
			}

			// TODO: following operations must be atomic
			// TODO: check if the folders already exists
			fsRootFolder := q.FsRootPath(userCfg.Name, "")
			if err = h.deps.FS().MkdirAll(fsRootFolder); err != nil {
				return "", err
			}
			uploadFolder := q.UploadFolder(userCfg.Name)
			if err = h.deps.FS().MkdirAll(uploadFolder); err != nil {
				return "", err
			}

			pwdHash, err := bcrypt.GenerateFromPassword([]byte(userCfg.Pwd), 10)
			if err != nil {
				return "", err
			}

			preferences := db.DefaultPreferences
			user := &db.User{
				ID:   h.deps.ID().Gen(),
				Name: userCfg.Name,
				Pwd:  string(pwdHash),
				Role: userCfg.Role,
				Quota: &db.Quota{
					SpaceLimit:         spaceLimit,
					UploadSpeedLimit:   uploadSpeedLimit,
					DownloadSpeedLimit: downloadSpeedLimit,
				},
				Preferences: &preferences,
			}

			err = h.deps.Users().AddUser(ctx, user)
			if err != nil {
				h.deps.Log().Warn("warning: failed to add user(%s): %s", user, err)
				return "", err
			}
			h.deps.Log().Infof("user(%s) is added", user.Name)
		}
	}
	return "", nil
}
