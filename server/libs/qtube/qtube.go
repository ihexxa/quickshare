package qtube

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

import (
	"quickshare/server/libs/fileidx"
)

var (
	ErrCopy    = errors.New("ServeFile: copy error")
	ErrUnknown = errors.New("ServeFile: unknown error")
)

type httpRange struct {
	start, length int64
}

func (ra *httpRange) GetStart() int64 {
	return ra.start
}
func (ra *httpRange) GetLength() int64 {
	return ra.length
}
func (ra *httpRange) SetStart(start int64) {
	ra.start = start
}
func (ra *httpRange) SetLength(length int64) {
	ra.length = length
}

func NewQTube(root string, copySpeed, maxRangeLen int64, filer FileReadSeekCloser) Downloader {
	return &QTube{
		Root:        root,
		BytesPerSec: copySpeed,
		MaxRangeLen: maxRangeLen,
		Filer:       filer,
	}
}

type QTube struct {
	Root        string
	BytesPerSec int64
	MaxRangeLen int64
	Filer       FileReadSeekCloser
}

type FileReadSeekCloser interface {
	Open(filePath string) (ReadSeekCloser, error)
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

const (
	ErrorInvalidRange = "ServeFile: invalid Range"
	ErrorInvalidSize  = "ServeFile: invalid Range total size"
)

func (tb *QTube) ServeFile(res http.ResponseWriter, req *http.Request, fileInfo *fileidx.FileInfo) error {
	headerRange := req.Header.Get("Range")

	switch {
	case req.Method == http.MethodHead:
		res.Header().Set("Accept-Ranges", "bytes")
		res.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Uploaded))
		res.Header().Set("Content-Type", "application/octet-stream")
		res.WriteHeader(http.StatusOK)

		return nil
	case headerRange == "":
		return tb.serveAll(res, fileInfo)
	default:
		return tb.serveRanges(res, headerRange, fileInfo)
	}
}

func (tb *QTube) serveAll(res http.ResponseWriter, fileInfo *fileidx.FileInfo) error {
	res.Header().Set("Accept-Ranges", "bytes")
	res.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(fileInfo.PathLocal)))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Uploaded))
	res.Header().Set("Content-Type", "application/octet-stream")
	res.Header().Set("Last-Modified", time.Unix(fileInfo.ModTime, 0).UTC().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)

	// TODO: need verify path
	file, openErr := tb.Filer.Open(filepath.Join(tb.Root, fileInfo.PathLocal))
	defer file.Close()
	if openErr != nil {
		return openErr
	}

	copyErr := tb.throttledCopyN(res, file, fileInfo.Uploaded)
	if copyErr != nil && copyErr != io.EOF {
		return copyErr
	}

	return nil
}

func (tb *QTube) serveRanges(res http.ResponseWriter, headerRange string, fileInfo *fileidx.FileInfo) error {
	ranges, rangeErr := getRanges(headerRange, fileInfo.Uploaded)
	if rangeErr != nil {
		http.Error(res, rangeErr.Error(), http.StatusRequestedRangeNotSatisfiable)
		return errors.New(rangeErr.Error())
	}

	switch {
	case len(ranges) == 1 || len(ranges) > 1:
		if tb.copyRange(res, ranges[0], fileInfo) != nil {
			return ErrCopy
		}
	default:
		// TODO: add support for multiple ranges
		return ErrUnknown
	}

	return nil
}

func getRanges(headerRange string, size int64) ([]httpRange, error) {
	ranges, raParseErr := parseRange(headerRange, size)
	// TODO: check max number of ranges, range start end
	if len(ranges) <= 0 || raParseErr != nil {
		return nil, errors.New(ErrorInvalidRange)
	}
	if sumRangesSize(ranges) > size {
		return nil, errors.New(ErrorInvalidSize)
	}

	return ranges, nil
}

func (tb *QTube) copyRange(res http.ResponseWriter, ra httpRange, fileInfo *fileidx.FileInfo) error {
	// TODO: comfirm this wont cause problem
	if ra.GetLength() > tb.MaxRangeLen {
		ra.SetLength(tb.MaxRangeLen)
	}

	// TODO: add headers(ETag): https://tools.ietf.org/html/rfc7233#section-4.1  p11 2nd paragraph
	res.Header().Set("Accept-Ranges", "bytes")
	res.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(fileInfo.PathLocal)))
	res.Header().Set("Content-Type", "application/octet-stream")
	res.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.start+ra.length-1, fileInfo.Uploaded))
	res.Header().Set("Content-Length", strconv.FormatInt(ra.GetLength(), 10))
	res.Header().Set("Last-Modified", time.Unix(fileInfo.ModTime, 0).UTC().Format(http.TimeFormat))
	res.WriteHeader(http.StatusPartialContent)

	// TODO: need verify path
	file, openErr := tb.Filer.Open(filepath.Join(tb.Root, fileInfo.PathLocal))
	defer file.Close()
	if openErr != nil {
		return openErr
	}

	if _, seekErr := file.Seek(ra.start, io.SeekStart); seekErr != nil {
		return seekErr
	}

	copyErr := tb.throttledCopyN(res, file, ra.length)
	if copyErr != nil && copyErr != io.EOF {
		return copyErr
	}

	return nil
}

func (tb *QTube) throttledCopyN(dst io.Writer, src io.Reader, length int64) error {
	sum := int64(0)
	timeSlot := time.Duration(1 * time.Second)

	for sum < length {
		start := time.Now()
		chunkSize := length - sum
		if length-sum > tb.BytesPerSec {
			chunkSize = tb.BytesPerSec
		}

		copied, err := io.CopyN(dst, src, chunkSize)
		if err != nil {
			return err
		}

		sum += copied
		end := time.Now()
		if end.Before(start.Add(timeSlot)) {
			time.Sleep(start.Add(timeSlot).Sub(end))
		}
	}

	return nil
}

func parseRange(headerRange string, size int64) ([]httpRange, error) {
	if headerRange == "" {
		return nil, nil // header not present
	}

	const keyByte = "bytes="
	if !strings.HasPrefix(headerRange, keyByte) {
		return nil, errors.New("byte= not found")
	}

	var ranges []httpRange
	noOverlap := false
	for _, ra := range strings.Split(headerRange[len(keyByte):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}

		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("- not found")
		}

		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r httpRange
		if start == "" {
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, errors.New("parseRanges: no overlap")
	}
	return ranges, nil
}

func sumRangesSize(ranges []httpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}
