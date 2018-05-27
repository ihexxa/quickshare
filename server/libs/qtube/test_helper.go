package qtube

type StubFile struct {
	Content string
	Offset  int64
}

func (file *StubFile) Read(p []byte) (int, error) {
	copied := copy(p[:], []byte(file.Content)[:len(p)])
	return copied, nil
}

func (file *StubFile) Seek(offset int64, whence int) (int64, error) {
	file.Offset = offset
	return offset, nil
}

func (file *StubFile) Close() error {
	return nil
}

type stubFiler struct {
	file *StubFile
}

func (filer *stubFiler) Open(filePath string) (ReadSeekCloser, error) {
	return filer.file, nil
}
