package boltdb

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ihexxa/quickshare/src/db"
)

func (fi *FileInfoStore) getShareID(payload string) (string, error) {
	if len(payload) == 0 {
		return "", ErrEmpty
	}

	for i := 0; i < maxHashingTime; i++ {
		msg := strings.Repeat(payload, i+1)
		h := sha1.New()
		_, err := io.WriteString(h, msg)
		if err != nil {
			return "", err
		}

		shareID := fmt.Sprintf("%x", h.Sum(nil))[:7]
		shareDir, ok := fi.store.GetStringIn(db.ShareIDNs, shareID)
		if !ok {
			return shareID, nil
		} else if ok && shareDir == payload {
			return shareID, nil
		}
	}

	return "", ErrConflicted
}

func (fi *FileInfoStore) GetSharingDir(hashID string) (string, error) {
	dirPath, ok := fi.store.GetStringIn(db.ShareIDNs, hashID)
	if !ok {
		return "", ErrSharingNotFound
	}
	return dirPath, nil
}

func (fi *FileInfoStore) AddSharing(dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.getInfo(dirPath)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		info = &db.FileInfo{
			IsDir: true,
		}
	}

	// TODO: ensure Atomicity
	shareID, err := fi.getShareID(dirPath)
	if err != nil {
		return err
	}
	err = fi.store.SetStringIn(db.ShareIDNs, shareID, dirPath)
	if err != nil {
		return err
	}

	info.Shared = true
	info.ShareID = shareID
	return fi.setInfo(dirPath, info)
}

func (fi *FileInfoStore) DelSharing(dirPath string) error {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	info, err := fi.getInfo(dirPath)
	if err != nil {
		return err
	}

	// TODO: ensure Atomicity
	// In the bolt, if the key does not exist
	// then nothing is done and a nil error is returned

	// because before this version, shareIDs are not removed correctly
	// so it iterates all shareIDs and cleans remaining entries
	shareIDtoDir, err := fi.store.ListStringsIn(db.ShareIDNs)
	if err != nil {
		return err
	}

	for shareID, shareDir := range shareIDtoDir {
		if shareDir == dirPath {
			err = fi.store.DelStringIn(db.ShareIDNs, shareID)
			if err != nil {
				return err
			}
		}
	}

	info.ShareID = ""
	info.Shared = false
	return fi.setInfo(dirPath, info)
}

func (fi *FileInfoStore) GetSharing(dirPath string) (bool, bool) {
	fi.mtx.Lock()
	defer fi.mtx.Unlock()

	// TODO: differentiate error and not exist
	info, err := fi.getInfo(dirPath)
	if err != nil {
		return false, false
	}
	return info.IsDir && info.Shared, true
}

func (fi *FileInfoStore) ListSharings(prefix string) (map[string]string, error) {
	infoStrs, err := fi.store.ListStringsByPrefixIn(prefix, db.FileInfoNs)
	if err != nil {
		return nil, err
	}

	info := &db.FileInfo{}
	sharings := map[string]string{}
	for itemPath, infoStr := range infoStrs {
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, fmt.Errorf("list sharing error: %w", err)
		}

		if info.IsDir && info.Shared {
			sharings[itemPath] = info.ShareID
		}
	}

	return sharings, nil
}
