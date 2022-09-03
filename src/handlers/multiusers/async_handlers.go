package multiusers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ihexxa/quickshare/src/worker"
)

const (
	MsgTypeResetUsedSpace = "reset-used-space"
)

type UsedSpaceParams struct {
	UserID       uint64
	UserHomePath string
}

func (h *MultiUsersSvc) resetUsedSpace(msg worker.IMsg) error {
	params := &UsedSpaceParams{}
	err := json.Unmarshal([]byte(msg.Body()), params)
	if err != nil {
		return fmt.Errorf("fail to unmarshal sha1 msg: %w", err)
	}

	usedSpace := int64(0)
	dirQueue := []string{params.UserHomePath}
	for len(dirQueue) > 0 {
		dirPath := dirQueue[0]
		dirQueue = dirQueue[1:]

		infos, err := h.deps.FS().ListDir(dirPath)
		if err != nil {
			return err
		}

		for _, info := range infos {
			if info.IsDir() {
				dirQueue = append(dirQueue, filepath.Join(dirPath, info.Name()))
			} else {
				usedSpace += info.Size()
			}
		}
	}

	return h.deps.Users().ResetUsed(context.TODO(), params.UserID, usedSpace) // TODO: use source context
}
