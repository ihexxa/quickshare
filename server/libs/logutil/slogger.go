package logutil

import (
	"io"
	"log"
)

func NewSlog(out io.Writer, prefix string) LogUtil {
	return log.New(out, prefix, log.Ldate|log.Ltime|log.Lshortfile)
}

type Slog *log.Logger
