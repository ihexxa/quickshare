package fileidx

import (
	"sync"
)

const (
	// StateStarted = after startUpload before upload
	StateStarted = "started"
	// StateUploading =after upload before finishUpload
	StateUploading = "uploading"
	// StateDone = after finishedUpload
	StateDone = "done"
)

type FileInfo struct {
	Id        string
	DownLimit int
	ModTime   int64
	PathLocal string
	State     string
	Uploaded  int64
}

type FileIndex interface {
	Add(fileInfo *FileInfo) int
	Del(id string)
	SetId(id string, newId string) bool
	SetDownLimit(id string, downLimit int) bool
	DecrDownLimit(id string) (int, bool)
	SetState(id string, state string) bool
	IncrUploaded(id string, uploaded int64) int64
	Get(id string) (*FileInfo, bool)
	List() map[string]*FileInfo
}

func NewMemFileIndex(cap int) *MemFileIndex {
	return &MemFileIndex{
		cap:   cap,
		infos: make(map[string]*FileInfo, 0),
	}
}

func NewMemFileIndexWithMap(cap int, infos map[string]*FileInfo) *MemFileIndex {
	return &MemFileIndex{
		cap:   cap,
		infos: infos,
	}
}

type MemFileIndex struct {
	cap   int
	infos map[string]*FileInfo
	mux   sync.RWMutex
}

func (idx *MemFileIndex) Add(fileInfo *FileInfo) int {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	if len(idx.infos) >= idx.cap {
		return 1
	}

	if _, found := idx.infos[fileInfo.Id]; found {
		return -1
	}

	idx.infos[fileInfo.Id] = fileInfo
	return 0
}

func (idx *MemFileIndex) Del(id string) {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	delete(idx.infos, id)
}

func (idx *MemFileIndex) SetId(id string, newId string) bool {
	if id == newId {
		return true
	}

	idx.mux.Lock()
	defer idx.mux.Unlock()

	info, found := idx.infos[id]
	if !found {
		return false
	}

	if _, foundNewId := idx.infos[newId]; foundNewId {
		return false
	}

	idx.infos[newId] = info
	idx.infos[newId].Id = newId
	delete(idx.infos, id)
	return true
}

func (idx *MemFileIndex) SetDownLimit(id string, downLimit int) bool {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	info, found := idx.infos[id]
	if !found {
		return false
	}

	info.DownLimit = downLimit
	return true
}

func (idx *MemFileIndex) DecrDownLimit(id string) (int, bool) {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	info, found := idx.infos[id]
	if !found || info.State != StateDone {
		return 0, false
	}

	if info.DownLimit == 0 {
		return 1, false
	}

	if info.DownLimit > 0 {
		// info.DownLimit means unlimited
		info.DownLimit = info.DownLimit - 1
	}
	return 1, true
}

func (idx *MemFileIndex) SetState(id string, state string) bool {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	info, found := idx.infos[id]
	if !found {
		return false
	}

	info.State = state
	return true
}

func (idx *MemFileIndex) IncrUploaded(id string, uploaded int64) int64 {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	info, found := idx.infos[id]
	if !found {
		return 0
	}

	info.Uploaded = info.Uploaded + uploaded
	return info.Uploaded
}

func (idx *MemFileIndex) Get(id string) (*FileInfo, bool) {
	idx.mux.RLock()
	defer idx.mux.RUnlock()

	infos, found := idx.infos[id]
	return infos, found
}

func (idx *MemFileIndex) List() map[string]*FileInfo {
	idx.mux.RLock()
	defer idx.mux.RUnlock()

	return idx.infos
}

// TODO: add unit tests
