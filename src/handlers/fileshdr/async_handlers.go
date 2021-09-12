package fileshdr

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ihexxa/quickshare/src/worker"
)
const MsgTypeSha1 = "sha1"

type Sha1Params struct {
	FilePath string
}

func (h *FileHandlers) genSha1(msg worker.IMsg) error {
	taskInputs := &Sha1Params{}
	err := json.Unmarshal([]byte(msg.Body()), taskInputs)
	if err != nil {
		return fmt.Errorf("fail to unmarshal sha1 msg: %w", err)
	}

	f, err := h.deps.FS().GetFileReader(taskInputs.FilePath)
	if err != nil {
		return fmt.Errorf("fail to get reader: %s", err)
	}

	hasher := sha1.New()
	buf := make([]byte, 4096)
	_, err = io.CopyBuffer(hasher, f, buf)
	if err != nil {
		return err
	}

	sha1Sign := fmt.Sprintf("%x", hasher.Sum(nil))
	err = h.deps.FileInfos().SetSha1(taskInputs.FilePath, sha1Sign)
	if err != nil {
		return fmt.Errorf("fail to set sha1: %s", err)
	}
	return nil
}
