package apis

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/cfg"
	"github.com/ihexxa/quickshare/server/libs/encrypt"
	"github.com/ihexxa/quickshare/server/libs/errutil"
	"github.com/ihexxa/quickshare/server/libs/fileidx"
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/httpworker"
	"github.com/ihexxa/quickshare/server/libs/limiter"
	"github.com/ihexxa/quickshare/server/libs/logutil"
	"github.com/ihexxa/quickshare/server/libs/walls"
)

const testCap = 3

func initServiceForUploadTest(config *cfg.Config, indexMap map[string]*fileidx.FileInfo) *SrvShare {
	logger := logutil.NewSlog(os.Stdout, config.AppName)
	setLog := func(srv *SrvShare) {
		srv.Log = logger
	}

	setWorkerPool := func(srv *SrvShare) {
		workerPoolSize := config.WorkerPoolSize
		taskQueueSize := config.TaskQueueSize
		srv.WorkerPool = httpworker.NewWorkerPool(workerPoolSize, taskQueueSize, logger)
	}

	setWalls := func(srv *SrvShare) {
		encrypterMaker := encrypt.JwtEncrypterMaker
		ipLimiter := limiter.NewRateLimiter(config.LimiterCap, config.LimiterTtl, config.LimiterCyc, config.BucketCap, map[int16]int16{})
		opLimiter := limiter.NewRateLimiter(config.LimiterCap, config.LimiterTtl, config.LimiterCyc, config.BucketCap, map[int16]int16{})
		srv.Walls = walls.NewAccessWalls(config, ipLimiter, opLimiter, encrypterMaker)
	}

	setIndex := func(srv *SrvShare) {
		srv.Index = fileidx.NewMemFileIndexWithMap(len(indexMap)+testCap, indexMap)
	}

	setFs := func(srv *SrvShare) {
		srv.Fs = &stubFsUtil{}
	}

	setEncryptor := func(srv *SrvShare) {
		srv.Encryptor = &encrypt.HmacEncryptor{Key: config.SecretKeyByte}
	}

	errChecker := errutil.NewErrChecker(!config.Production, logger)
	setErr := func(srv *SrvShare) {
		srv.Err = errChecker
	}

	return InitSrvShare(config, setIndex, setWalls, setWorkerPool, setFs, setEncryptor, setLog, setErr)
}

func TestStartUpload(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		FileName string
	}
	type Output struct {
		Response interface{}
		IndexMap map[string]*fileidx.FileInfo
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "invalid file name",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{},
			},
			Input: Input{
				FileName: "",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{},
				Response: httputil.Err400,
			},
		},
		testCase{
			Desc: "succeed to start uploading",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{},
			},
			Input: Input{
				FileName: "filename",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					DefaultId: &fileidx.FileInfo{
						Id:        DefaultId,
						DownLimit: conf.DownLimit,
						ModTime:   time.Now().UnixNano(),
						PathLocal: "filename",
						Uploaded:  0,
						State:     fileidx.StateUploading,
					},
				},
				Response: &ByteRange{
					ShareId: DefaultId,
					Start:   0,
					Length:  conf.MaxUpBytesPerSec,
				},
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForUploadTest(conf, testCase.Init.IndexMap)

		// verify CreateFile
		expectCreateFileName = filepath.Join(conf.PathLocal, testCase.FileName)

		response := srv.startUpload(testCase.FileName)

		// verify index
		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			t.Fatalf("startUpload: index not equal got: %v, %v,  expect: %v", srv.Index.List(), response, testCase.Output.IndexMap)
		}

		// verify response
		switch expectRes := testCase.Output.Response.(type) {
		case *ByteRange:
			res := response.(*ByteRange)
			if res.ShareId != expectRes.ShareId ||
				res.Start != expectRes.Start ||
				res.Length != expectRes.Length {
				t.Fatalf(fmt.Sprintf("startUpload: res=%v expect=%v", res, expectRes))
			}
		case httputil.MsgRes:
			if response != expectRes {
				t.Fatalf(fmt.Sprintf("startUpload: response=%v expectRes=%v", response, expectRes))
			}
		default:
			t.Fatalf(fmt.Sprintf("startUpload: type not found: %T %T", testCase.Output.Response, httputil.Err400))
		}
	}
}

func TestUpload(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId string
		Start   int64
		Len     int64
		Chunk   io.Reader
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response interface{}
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "shareid does not exist",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{},
			},
			Input: Input{
				ShareId: DefaultId,
				Start:   0,
				Len:     1,
				Chunk:   strings.NewReader(""),
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{},
				Response: httputil.Err404,
			},
		},
		testCase{
			Desc: "succeed",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					DefaultId: &fileidx.FileInfo{
						Id:        DefaultId,
						DownLimit: conf.MaxShares,
						PathLocal: "path/filename",
						State:     fileidx.StateUploading,
						Uploaded:  0,
					},
				},
			},
			Input: Input{
				ShareId: DefaultId,
				Start:   0,
				Len:     1,
				Chunk:   strings.NewReader("a"),
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					DefaultId: &fileidx.FileInfo{
						Id:        DefaultId,
						DownLimit: conf.MaxShares,
						PathLocal: "path/filename",
						State:     fileidx.StateUploading,
						Uploaded:  1,
					},
				},
				Response: &ByteRange{
					ShareId: DefaultId,
					Start:   1,
					Length:  conf.MaxUpBytesPerSec,
				},
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForUploadTest(conf, testCase.Init.IndexMap)

		response := srv.upload(
			testCase.Input.ShareId,
			testCase.Input.Start,
			testCase.Input.Len,
			testCase.Input.Chunk,
		)

		// TODO: not verified copyChunk

		// verify index
		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			t.Fatalf("upload: index not identical got: %v want: %v", srv.Index.List(), testCase.Output.IndexMap)
		}
		// verify response
		switch response.(type) {
		case *ByteRange:
			br := testCase.Output.Response.(*ByteRange)
			res := response.(*ByteRange)
			if res.ShareId != br.ShareId || res.Start != br.Start || res.Length != br.Length {
				t.Fatalf(fmt.Sprintf("upload: response=%v expectRes=%v", res, br))
			}
		default:
			if response != testCase.Output.Response {
				t.Fatalf(fmt.Sprintf("upload: response=%v expectRes=%v", response, testCase.Output.Response))
			}
		}
	}
}

func TestFinishUpload(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId string
		Start   int64
		Len     int64
		Chunk   io.Reader
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response interface{}
	}
	type testCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []testCase{
		testCase{
			Desc: "success",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					DefaultId: &fileidx.FileInfo{
						Id:        DefaultId,
						DownLimit: conf.MaxShares,
						PathLocal: "path/filename",
						State:     fileidx.StateUploading,
						Uploaded:  1,
					},
				},
			},
			Input: Input{
				ShareId: DefaultId,
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					DefaultId: &fileidx.FileInfo{
						Id:        DefaultId,
						DownLimit: conf.MaxShares,
						PathLocal: "path/filename",
						State:     fileidx.StateDone,
						Uploaded:  1,
					},
				},
				Response: &ShareInfo{
					ShareId: DefaultId,
				},
			},
		},
		testCase{
			Desc: "shareId exists",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{},
			},
			Input: Input{
				ShareId: DefaultId,
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{},
				Response: httputil.Err404,
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForUploadTest(conf, testCase.Init.IndexMap)

		response := srv.finishUpload(testCase.ShareId)

		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			t.Fatalf("finishUpload: index not identical got: %v, want: %v", srv.Index.List(), testCase.Output.IndexMap)
		}

		switch res := response.(type) {
		case httputil.MsgRes:
			expectRes := testCase.Output.Response.(httputil.MsgRes)
			if res != expectRes {
				t.Fatalf(fmt.Sprintf("finishUpload: response=%v expectRes=%v", res, expectRes))
			}
		case *ShareInfo:
			info, found := testCase.Output.IndexMap[res.ShareId]
			if !found || info.State != fileidx.StateDone {
				// TODO: should use isValidUrl or better to verify result
				t.Fatalf(fmt.Sprintf("finishUpload: share info is not correct: received: %v expect: %v", res.ShareId, testCase.ShareId))
			}
		default:
			t.Fatalf(fmt.Sprintf("finishUpload: type not found: %T %T", response, testCase.Output.Response))
		}
	}
}
