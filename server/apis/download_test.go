package apis

import (
	"net/http"
	"os"
	"testing"
	"time"
)

import (
	"quickshare/server/libs/cfg"
	"quickshare/server/libs/errutil"
	"quickshare/server/libs/fileidx"
	"quickshare/server/libs/httputil"
	"quickshare/server/libs/logutil"
	"quickshare/server/libs/qtube"
)

func initServiceForDownloadTest(config *cfg.Config, indexMap map[string]*fileidx.FileInfo, content string) *SrvShare {
	setDownloader := func(srv *SrvShare) {
		srv.Downloader = stubDownloader{Content: content}
	}

	setIndex := func(srv *SrvShare) {
		srv.Index = fileidx.NewMemFileIndexWithMap(len(indexMap), indexMap)
	}

	setFs := func(srv *SrvShare) {
		srv.Fs = &stubFsUtil{
			MockFile: &qtube.StubFile{
				Content: content,
				Offset:  0,
			},
		}
	}

	logger := logutil.NewSlog(os.Stdout, config.AppName)
	setLog := func(srv *SrvShare) {
		srv.Log = logger
	}

	setErr := func(srv *SrvShare) {
		srv.Err = errutil.NewErrChecker(!config.Production, logger)
	}

	return InitSrvShare(config, setDownloader, setIndex, setFs, setLog, setErr)
}

func TestDownload(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		Content  string
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId string
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response interface{}
		Body     string
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "empty file index",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{},
				Response: httputil.Err404,
			},
		},
		testCase{
			Desc: "file info not found",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"1": &fileidx.FileInfo{},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"1": &fileidx.FileInfo{},
				},
				Response: httputil.Err404,
			},
		},
		testCase{
			Desc: "file not found because of state=uploading",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: 1,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateUploading,
						Uploaded:  1,
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: 1,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateUploading,
						Uploaded:  1,
					},
				},
				Response: httputil.Err404,
			},
		},
		testCase{
			Desc: "download failed because download limit = 0",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: 0,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: 0,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
				Response: httputil.Err412,
			},
		},
		testCase{
			Desc: "succeed to download",
			Init: Init{
				Content: "content",
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: 1,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: 0,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
				Response: 0,
				Body:     "content",
			},
		},
		testCase{
			Desc: "succeed to download DownLimit == -1",
			Init: Init{
				Content: "content",
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: -1,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: -1,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "path",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
				Response: 0,
				Body:     "content",
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForDownloadTest(conf, testCase.Init.IndexMap, testCase.Content)
		writer := &stubWriter{Headers: map[string][]string{}}
		response := srv.download(
			testCase.ShareId,
			writer,
			&http.Request{},
		)

		// verify downlimit
		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			info, _ := srv.Index.Get(testCase.ShareId)
			t.Fatalf(
				"download: index incorrect got=%v  want=%v",
				info,
				testCase.Output.IndexMap[testCase.ShareId],
			)
		}

		// verify response
		if response != testCase.Output.Response {
			t.Fatalf(
				"download: response incorrect response=%v testCase=%v",
				response,
				testCase.Output.Response,
			)
		}

		// verify writerContent
		if string(writer.Response) != testCase.Output.Body {
			t.Fatalf(
				"download: body incorrect got=%v want=%v",
				string(writer.Response),
				testCase.Output.Body,
			)
		}

	}
}
