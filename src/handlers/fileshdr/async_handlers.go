package fileshdr

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ihexxa/quickshare/src/worker"
)

const (
	MsgTypeSha1 = "sha1"
)

type Sha1Params struct {
	FilePath string
}

func (h *FileHandlers) genSha1(msg worker.IMsg) error {
	taskInputs := &Sha1Params{}
	err := json.Unmarshal([]byte(msg.Body()), taskInputs)
	if err != nil {
		return fmt.Errorf("fail to unmarshal sha1 msg: %w", err)
	}

	f, id, err := h.deps.FS().GetFileReader(taskInputs.FilePath)
	if err != nil {
		return fmt.Errorf("fail to get reader: %s", err)
	}
	defer func() {
		err := h.deps.FS().CloseReader(fmt.Sprint(id))
		if err != nil {
			h.deps.Log().Errorf("failed to close file: %s", err)
		}
	}()

	hasher := sha1.New()
	buf := make([]byte, 4096)
	_, err = io.CopyBuffer(hasher, f, buf)
	if err != nil {
		return fmt.Errorf("faile to copy buffer: %w", err)
	}

	sha1Sign := fmt.Sprintf("%x", hasher.Sum(nil))
	err = h.deps.FileInfos().SetSha1(taskInputs.FilePath, sha1Sign)
	if err != nil {
		return fmt.Errorf("fail to set sha1: %s", err)
	}

	return nil
}
