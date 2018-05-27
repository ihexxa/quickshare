package qtube

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/fileidx"
)

// Range format examples:
// Range: <unit>=<range-start>-
// Range: <unit>=<range-start>-<range-end>
// Range: <unit>=<range-start>-<range-end>, <range-start>-<range-end>
// Range: <unit>=<range-start>-<range-end>, <range-start>-<range-end>, <range-start>-<range-end>
func TestGetRanges(t *testing.T) {
	type Input struct {
		HeaderRange string
		Size        int64
	}
	type Output struct {
		Ranges   []httpRange
		ErrorMsg string
	}
	type testCase struct {
		Desc string
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "invalid range",
			Input: Input{
				HeaderRange: "bytes=start-invalid end",
				Size:        0,
			},
			Output: Output{
				ErrorMsg: ErrorInvalidRange,
			},
		},
		testCase{
			Desc: "invalid range total size",
			Input: Input{
				HeaderRange: "bytes=0-1, 2-3, 0-1, 0-2",
				Size:        3,
			},
			Output: Output{
				ErrorMsg: ErrorInvalidSize,
			},
		},
		testCase{
			Desc: "range ok",
			Input: Input{
				HeaderRange: "bytes=0-1, 2-3",
				Size:        4,
			},
			Output: Output{
				Ranges: []httpRange{
					httpRange{start: 0, length: 2},
					httpRange{start: 2, length: 2},
				},
				ErrorMsg: "",
			},
		},
	}

	for _, tCase := range testCases {
		ranges, err := getRanges(tCase.HeaderRange, tCase.Size)
		if err != nil {
			if err.Error() != tCase.ErrorMsg || len(tCase.Ranges) != 0 {
				t.Fatalf("getRanges: incorrect errorMsg want: %v got: %v", tCase.ErrorMsg, err.Error())
			} else {
				continue
			}
		} else {
			for id, ra := range ranges {
				if ra.GetStart() != tCase.Ranges[id].GetStart() {
					t.Fatalf("getRanges: incorrect range start, got: %v want: %v", ra.GetStart(), tCase.Ranges[id])
				}
				if ra.GetLength() != tCase.Ranges[id].GetLength() {
					t.Fatalf("getRanges: incorrect range length, got: %v want: %v", ra.GetLength(), tCase.Ranges[id])
				}
			}
		}
	}
}

func TestThrottledCopyN(t *testing.T) {
	type Init struct {
		BytesPerSec int64
		MaxRangeLen int64
	}
	type Input struct {
		Src    string
		Length int64
	}
	// after starting throttledCopyN by DstAtTime.AtMs millisecond,
	// copied valueshould equal to DstAtTime.Dst.
	type DstAtTime struct {
		AtMS int
		Dst  string
	}
	type Output struct {
		ExpectDsts []DstAtTime
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	verifyDsts := func(dst *bytes.Buffer, expectDsts []DstAtTime) {
		for _, expectDst := range expectDsts {
			// fmt.Printf("sleep: %d\n", time.Now().UnixNano())
			time.Sleep(time.Duration(expectDst.AtMS) * time.Millisecond)
			dstStr := string(dst.Bytes())
			// fmt.Printf("check: %d\n", time.Now().UnixNano())
			if dstStr != expectDst.Dst {
				panic(
					fmt.Sprintf(
						"throttledCopyN want: <%s> | got: <%s> | at: %d",
						expectDst.Dst,
						dstStr,
						expectDst.AtMS,
					),
				)
			}
		}
	}

	testCases := []testCase{
		testCase{
			Desc: "4 byte per sec",
			Init: Init{
				BytesPerSec: 5,
				MaxRangeLen: 10,
			},
			Input: Input{
				Src:    "aaaa_aaaa_",
				Length: 10,
			},
			Output: Output{
				ExpectDsts: []DstAtTime{
					DstAtTime{AtMS: 200, Dst: "aaaa_"},
					DstAtTime{AtMS: 200, Dst: "aaaa_"},
					DstAtTime{AtMS: 200, Dst: "aaaa_"},
					DstAtTime{AtMS: 600, Dst: "aaaa_aaaa_"},
					DstAtTime{AtMS: 200, Dst: "aaaa_aaaa_"},
					DstAtTime{AtMS: 200, Dst: "aaaa_aaaa_"},
				},
			},
		},
	}

	for _, tCase := range testCases {
		tb := NewQTube("", tCase.BytesPerSec, tCase.MaxRangeLen, &stubFiler{}).(*QTube)
		dst := bytes.NewBuffer(make([]byte, len(tCase.Src)))
		dst.Reset()

		go verifyDsts(dst, tCase.ExpectDsts)
		tb.throttledCopyN(dst, strings.NewReader(tCase.Src), tCase.Length)
	}
}

// TODO: using same stub with testhelper
type stubWriter struct {
	Headers    http.Header
	Response   []byte
	StatusCode int
}

func (w *stubWriter) Header() http.Header {
	return w.Headers
}

func (w *stubWriter) Write(body []byte) (int, error) {
	w.Response = append(w.Response, body...)
	return len(body), nil
}

func (w *stubWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func TestCopyRange(t *testing.T) {
	type Init struct {
		Content string
	}
	type Input struct {
		Range httpRange
		Info  fileidx.FileInfo
	}
	type Output struct {
		StatusCode int
		Headers    map[string][]string
		Body       string
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "copy ok",
			Init: Init{
				Content: "abcd_abcd_",
			},
			Input: Input{
				Range: httpRange{
					start:  6,
					length: 3,
				},
				Info: fileidx.FileInfo{
					ModTime:   0,
					Uploaded:  10,
					PathLocal: "filename.jpg",
				},
			},
			Output: Output{
				StatusCode: 206,
				Headers: map[string][]string{
					"Accept-Ranges":       []string{"bytes"},
					"Content-Disposition": []string{`attachment; filename="filename.jpg"`},
					"Content-Type":        []string{"application/octet-stream"},
					"Content-Range":       []string{"bytes 6-8/10"},
					"Content-Length":      []string{"3"},
					"Last-Modified":       []string{time.Unix(0, 0).UTC().Format(http.TimeFormat)},
				},
				Body: "abc",
			},
		},
	}

	for _, tCase := range testCases {
		filer := &stubFiler{
			&StubFile{
				Content: tCase.Content,
				Offset:  0,
			},
		}
		tb := NewQTube("", 100, 100, filer).(*QTube)
		res := &stubWriter{
			Headers:  make(map[string][]string),
			Response: make([]byte, 0),
		}
		err := tb.copyRange(res, tCase.Range, &tCase.Info)
		if err != nil {
			t.Fatalf("copyRange: %v", err)
		}
		if res.StatusCode != tCase.Output.StatusCode {
			t.Fatalf("copyRange: statusCode not match got: %v want: %v", res.StatusCode, tCase.Output.StatusCode)
		}
		if string(res.Response) != tCase.Output.Body {
			t.Fatalf("copyRange: body not match \ngot: %v \nwant: %v", string(res.Response), tCase.Output.Body)
		}
		for key, vals := range tCase.Output.Headers {
			if res.Header().Get(key) != vals[0] {
				t.Fatalf("copyRange: header not match %v got: %v want: %v", key, res.Header().Get(key), vals[0])
			}
		}
		if res.StatusCode != tCase.Output.StatusCode {
			t.Fatalf("copyRange: statusCodes are not match %v", res.StatusCode, tCase.Output.StatusCode)
		}
	}
}

func TestServeAll(t *testing.T) {
	type Init struct {
		Content string
	}
	type Input struct {
		Info fileidx.FileInfo
	}
	type Output struct {
		StatusCode int
		Headers    map[string][]string
		Body       string
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "copy ok",
			Init: Init{
				Content: "abcd_abcd_",
			},
			Input: Input{
				Info: fileidx.FileInfo{
					ModTime:   0,
					Uploaded:  10,
					PathLocal: "filename.jpg",
				},
			},
			Output: Output{
				StatusCode: 200,
				Headers: map[string][]string{
					"Accept-Ranges":       []string{"bytes"},
					"Content-Disposition": []string{`attachment; filename="filename.jpg"`},
					"Content-Type":        []string{"application/octet-stream"},
					"Content-Length":      []string{"10"},
					"Last-Modified":       []string{time.Unix(0, 0).UTC().Format(http.TimeFormat)},
				},
				Body: "abcd_abcd_",
			},
		},
	}

	for _, tCase := range testCases {
		filer := &stubFiler{
			&StubFile{
				Content: tCase.Content,
				Offset:  0,
			},
		}
		tb := NewQTube("", 100, 100, filer).(*QTube)
		res := &stubWriter{
			Headers:  make(map[string][]string),
			Response: make([]byte, 0),
		}
		err := tb.serveAll(res, &tCase.Info)
		if err != nil {
			t.Fatalf("serveAll: %v", err)
		}
		if res.StatusCode != tCase.Output.StatusCode {
			t.Fatalf("serveAll: statusCode not match got: %v want: %v", res.StatusCode, tCase.Output.StatusCode)
		}
		if string(res.Response) != tCase.Output.Body {
			t.Fatalf("serveAll: body not match \ngot: %v \nwant: %v", string(res.Response), tCase.Output.Body)
		}
		for key, vals := range tCase.Output.Headers {
			if res.Header().Get(key) != vals[0] {
				t.Fatalf("serveAll: header not match %v got: %v want: %v", key, res.Header().Get(key), vals[0])
			}
		}
		if res.StatusCode != tCase.Output.StatusCode {
			t.Fatalf("serveAll: statusCodes are not match %v", res.StatusCode, tCase.Output.StatusCode)
		}
	}
}
