package mem

// type MemFS struct {
// 	files map[string][]byte
// 	dirs  map[string][]string
// }

// type MemFileInfo struct {
// 	name  string
// 	size  int64
// 	isDir bool
// }

// func (fi *MemFileInfo) Name() string {
// 	return fi.name
// }
// func (fi *MemFileInfo) Size() int64 {
// 	return fi.size
// }
// func (fi *MemFileInfo) Mode() os.FileMode {
// 	return 0666
// }
// func (fi *MemFileInfo) ModTime() time.Time {
// 	return time.Now()
// }
// func (fi *MemFileInfo) IsDir() bool {
// 	return fi.isDir
// }
// func (fi *MemFileInfo) Sys() interface{} {
// 	return ""
// }

// func NewMemFS() *MemFS {
// 	return &MemFS{
// 		files: map[string][]byte{},
// 		dirs:  map[string][]string{},
// 	}
// }

// // Create(filePath string) error
// // MkdirAll(filePath string) error
// // Remove(filePath string) error

// func (fs *MemFS) Create(filePath string) error {
// 	dirPath := path.Dir(filePath)
// 	files, ok := fs.dirs[dirPath]
// 	if !ok {
// 		fs.dirs[dirPath] = []string{}
// 	}

// 	fs.dirs[dirPath] = append(fs.dirs[dirPath], filePath)
// 	fs.files[filePath] = []byte("")
// 	return nil
// }

// func (fs *MemFS) MkdirAll(dirPath string) error {
// 	_, ok := fs.dirs[dirPath]
// 	if ok {
// 		return os.ErrExist
// 	}
// 	fs.dirs[dirPath] = []string{}
// 	return nil
// }

// func (fs *MemFS) Remove(filePath string) error {
// 	files, ok := fs.dirs[filePath]
// 	if ok {
// 		for _, fileName := range files {
// 			d
// 		}
// 	}

// 	delete(fs.dirs, filePath)
// 	delete(fs.files, filePath)
// 	return nil
// }

// func (fs *MemFS) Rename(oldpath, newpath string) error {
// 	content, ok := fs.files[oldpath]
// 	if !ok {
// 		return os.ErrNotExist
// 	}
// 	delete(fs.files, oldpath)

// 	newDir := path.Dir(newpath)
// 	_, ok = fs.dirs[newDir]
// 	if !ok {
// 		fs.dirs[newDir] = []string{}
// 	}
// 	fs.dirs[newDir] = append(fs.dirs[newDir], newpath)
// 	fs.files[newpath] = content
// 	return nil
// }

// func (fs *MemFS) ReadAt(filePath string, b []byte, off int64) (n int, err error) {
// 	content, ok := fs.files[filePath]
// 	if !ok {
// 		return 0, os.ErrNotExist
// 	}

// 	if off >= int64(len(content)) {
// 		return 0, errors.New("offset > fileSize")
// 	}
// 	right := off + int64(len(b))
// 	if right > int64(len(content)) {
// 		right = int64(len(content))
// 	}
// 	return copy(b, content[off:right]), nil
// }

// func (fs *MemFS) WriteAt(filePath string, b []byte, off int64) (n int, err error) {
// 	content, ok := fs.files[filePath]
// 	if !ok {
// 		return 0, os.ErrNotExist
// 	}

// 	if off >= int64(len(content)) {
// 		return 0, errors.New("offset > fileSize")
// 	} else if off+int64(len(b)) > int64(len(content)) {
// 		fs.files[filePath] = append(
// 			fs.files[filePath],
// 			make([]byte, off+int64(len(b))-int64(len(content)))...,
// 		)
// 	}

// 	copy(fs.files[filePath][off:], b)
// 	return len(b), nil
// }

// func (fs *MemFS) Stat(filePath string) (os.FileInfo, error) {
// 	_, ok := fs.dirs[filePath]
// 	if ok {
// 		return &MemFileInfo{
// 			name:  filePath,
// 			size:  0,
// 			isDir: true,
// 		}, nil
// 	}

// 	content, ok := fs.files[filePath]
// 	if ok {
// 		return &MemFileInfo{
// 			name:  filePath,
// 			size:  int64(len(content)),
// 			isDir: false,
// 		}, nil
// 	}
// 	return nil, os.ErrNotExist
// }

// func (fs *MemFS) Close() error {
// 	return nil
// }

// func (fs *MemFS) Sync() error {
// 	return nil
// }

// func (fs *MemFS) GetFileReader(filePath string) (ReadCloseSeeker, error) {
// 	content, ok := fs.files[filePath]
// 	if !ok {
// 		return nil, os.ErrNotExist
// 	}
// 	return bytes.NewReader(content)
// }

// func (fs *MemFS) Root() string {
// 	return ""
// }

// func (fs *MemFS) ListDir(filePath string) ([]os.FileInfo, error) {
// 	files, ok := fs.dirs[filePath]
// 	if !ok {
// 		return nil, os.ErrNotExist
// 	}

// 	infos := []*MemFileInfo{}
// 	for _, fileName := range files {
// 		infos = append(infos, &MemFileInfo{
// 			name:  fileName,
// 			size:  int64(len(fs.files[fileName])),
// 			isDir: false,
// 		})
// 	}

// 	return infos
// }
