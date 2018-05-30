package apis

import (
	"os"
	"path/filepath"
	"testing"
)

import (
	"github.com/ihexxa/quickshare/server/libs/cfg"
	"github.com/ihexxa/quickshare/server/libs/errutil"
	"github.com/ihexxa/quickshare/server/libs/fileidx"
	"github.com/ihexxa/quickshare/server/libs/httputil"
	"github.com/ihexxa/quickshare/server/libs/logutil"
)

const mockShadowId = "shadowId"
const mockPublicId = "publicId"

func initServiceForFileInfoTest(
	config *cfg.Config,
	indexMap map[string]*fileidx.FileInfo,
	useShadowEnc bool,
	localFileInfos []*fileidx.FileInfo,
) *SrvShare {
	setIndex := func(srv *SrvShare) {
		srv.Index = fileidx.NewMemFileIndexWithMap(len(indexMap), indexMap)
	}

	setFs := func(srv *SrvShare) {
		srv.Fs = &stubFsUtil{MockLocalFileInfos: localFileInfos}
	}

	logger := logutil.NewSlog(os.Stdout, config.AppName)
	setLog := func(srv *SrvShare) {
		srv.Log = logger
	}

	errChecker := errutil.NewErrChecker(!config.Production, logger)
	setErr := func(srv *SrvShare) {
		srv.Err = errChecker
	}

	var setEncryptor AddDep
	if useShadowEnc {
		setEncryptor = func(srv *SrvShare) {
			srv.Encryptor = &stubEncryptor{MockResult: mockShadowId}
		}
	} else {
		setEncryptor = func(srv *SrvShare) {
			srv.Encryptor = &stubEncryptor{MockResult: mockPublicId}
		}
	}

	return InitSrvShare(config, setIndex, setFs, setEncryptor, setLog, setErr)
}

func TestList(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type TestCase struct {
		Desc string
		Output
	}

	testCases := []TestCase{
		TestCase{
			Desc: "success",
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id: "0",
					},
					"1": &fileidx.FileInfo{
						Id: "1",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForFileInfoTest(conf, testCase.Output.IndexMap, true, []*fileidx.FileInfo{})
		response := srv.list()
		resInfos := response.(*ResInfos)

		for _, info := range resInfos.List {
			infoFromSrv, found := srv.Index.Get(info.Id)
			if !found || infoFromSrv.Id != info.Id {
				t.Fatalf("list: file infos are not identical")
			}
		}

		if len(resInfos.List) != len(srv.Index.List()) {
			t.Fatalf("list: file infos are not identical")
		}
	}
}

func TestDel(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId string
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response httputil.MsgRes
	}
	type TestCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []TestCase{
		TestCase{
			Desc: "success",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id: "0",
					},
					"1": &fileidx.FileInfo{
						Id: "1",
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"1": &fileidx.FileInfo{
						Id: "1",
					},
				},
				Response: httputil.Ok200,
			},
		},
		TestCase{
			Desc: "not found",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"1": &fileidx.FileInfo{
						Id: "1",
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"1": &fileidx.FileInfo{
						Id: "1",
					},
				},
				Response: httputil.Err404,
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForFileInfoTest(conf, testCase.Init.IndexMap, true, []*fileidx.FileInfo{})
		response := srv.del(testCase.ShareId)
		res := response.(httputil.MsgRes)

		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			t.Fatalf("del: index incorrect")
		}

		if res != testCase.Output.Response {
			t.Fatalf("del: response incorrect got: %v, want: %v", res, testCase.Output.Response)
		}
	}
}

func TestShadowId(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId string
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response interface{}
	}
	type TestCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []TestCase{
		TestCase{
			Desc: "success",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id: "0",
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					mockShadowId: &fileidx.FileInfo{
						Id: mockShadowId,
					},
				},
				Response: &ShareInfo{
					ShareId: mockShadowId,
				},
			},
		},
		TestCase{
			Desc: "original id not exists",
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
		TestCase{
			Desc: "dest id exists",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id: "0",
					},
					mockShadowId: &fileidx.FileInfo{
						Id: mockShadowId,
					},
				},
			},
			Input: Input{
				ShareId: "0",
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id: "0",
					},
					mockShadowId: &fileidx.FileInfo{
						Id: mockShadowId,
					},
				},
				Response: httputil.Err412,
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForFileInfoTest(conf, testCase.Init.IndexMap, true, []*fileidx.FileInfo{})
		response := srv.shadowId(testCase.ShareId)

		switch response.(type) {
		case *ShareInfo:
			res := response.(*ShareInfo)

			if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
				info, found := srv.Index.Get(mockShadowId)
				t.Fatalf(
					"shadowId: index incorrect got %v found: %v want %v",
					info,
					found,
					testCase.Output.IndexMap[mockShadowId],
				)
			}

			if res.ShareId != mockShadowId {
				t.Fatalf("shadowId: mockId incorrect")
			}

		case httputil.MsgRes:
			res := response.(httputil.MsgRes)

			if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
				t.Fatalf("shadowId: map not identical")
			}

			if res != testCase.Output.Response {
				t.Fatalf("shadowId: response incorrect")
			}
		default:
			t.Fatalf("shadowId: return type not found")
		}
	}
}

func TestPublishId(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId string
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response interface{}
	}
	type TestCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []TestCase{
		TestCase{
			Desc: "success",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					mockShadowId: &fileidx.FileInfo{
						Id: mockShadowId,
					},
				},
			},
			Input: Input{
				ShareId: mockShadowId,
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					mockPublicId: &fileidx.FileInfo{
						Id: mockPublicId,
					},
				},
				Response: &ShareInfo{
					ShareId: mockPublicId,
				},
			},
		},
		TestCase{
			Desc: "original id not exists",
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
		TestCase{
			Desc: "dest id exists",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					mockShadowId: &fileidx.FileInfo{
						Id: mockShadowId,
					},
					mockPublicId: &fileidx.FileInfo{
						Id: mockPublicId,
					},
				},
			},
			Input: Input{
				ShareId: mockShadowId,
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					mockShadowId: &fileidx.FileInfo{
						Id: mockShadowId,
					},
					mockPublicId: &fileidx.FileInfo{
						Id: mockPublicId,
					},
				},
				Response: httputil.Err412,
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForFileInfoTest(conf, testCase.Init.IndexMap, false, []*fileidx.FileInfo{})
		response := srv.publishId(testCase.ShareId)

		switch response.(type) {
		case *ShareInfo:
			res := response.(*ShareInfo)

			if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
				info, found := srv.Index.Get(mockPublicId)
				t.Fatalf(
					"shadowId: index incorrect got %v found: %v want %v",
					info,
					found,
					testCase.Output.IndexMap[mockPublicId],
				)
			}

			if res.ShareId != mockPublicId {
				t.Fatalf("shadowId: mockId incorrect %v %v", res.ShareId, mockPublicId)
			}

		case httputil.MsgRes:
			res := response.(httputil.MsgRes)

			if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
				t.Fatalf("shadowId: map not identical")
			}

			if res != testCase.Output.Response {
				t.Fatalf("shadowId: response incorrect got: %v want: %v", res, testCase.Output.Response)
			}
		default:
			t.Fatalf("shadowId: return type not found")
		}
	}
}

func TestSetDownLimit(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false
	mockDownLimit := 100

	type Init struct {
		IndexMap map[string]*fileidx.FileInfo
	}
	type Input struct {
		ShareId   string
		DownLimit int
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response httputil.MsgRes
	}
	type TestCase struct {
		Desc string
		Init
		Input
		Output
	}

	testCases := []TestCase{
		TestCase{
			Desc: "success",
			Init: Init{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id: "0",
					},
				},
			},
			Input: Input{
				ShareId:   "0",
				DownLimit: mockDownLimit,
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					"0": &fileidx.FileInfo{
						Id:        "0",
						DownLimit: mockDownLimit,
					},
				},
				Response: httputil.Ok200,
			},
		},
		TestCase{
			Desc: "not found",
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
	}

	for _, testCase := range testCases {
		srv := initServiceForFileInfoTest(conf, testCase.Init.IndexMap, true, []*fileidx.FileInfo{})
		response := srv.setDownLimit(testCase.ShareId, mockDownLimit)
		res := response.(httputil.MsgRes)

		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			info, _ := srv.Index.Get(testCase.ShareId)
			t.Fatalf(
				"setDownLimit: index incorrect got: %v want: %v",
				info,
				testCase.Output.IndexMap[testCase.ShareId],
			)
		}

		if res != testCase.Output.Response {
			t.Fatalf("setDownLimit: response incorrect got: %v, want: %v", res, testCase.Output.Response)
		}
	}
}

func TestAddLocalFiles(t *testing.T) {
	conf := cfg.NewConfig()
	conf.Production = false

	type Init struct {
		Infos []*fileidx.FileInfo
	}
	type Output struct {
		IndexMap map[string]*fileidx.FileInfo
		Response httputil.MsgRes
	}
	type TestCase struct {
		Desc string
		Init
		Output
	}

	testCases := []TestCase{
		TestCase{
			Desc: "success",
			Init: Init{
				Infos: []*fileidx.FileInfo{
					&fileidx.FileInfo{
						Id:        "",
						DownLimit: 0,
						ModTime:   13,
						PathLocal: "filename1",
						State:     "",
						Uploaded:  13,
					},
				},
			},
			Output: Output{
				IndexMap: map[string]*fileidx.FileInfo{
					mockPublicId: &fileidx.FileInfo{
						Id:        mockPublicId,
						DownLimit: conf.DownLimit,
						ModTime:   13,
						PathLocal: filepath.Join(conf.PathLocal, "filename1"),
						State:     fileidx.StateDone,
						Uploaded:  13,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		srv := initServiceForFileInfoTest(conf, testCase.Output.IndexMap, false, testCase.Init.Infos)
		response := srv.AddLocalFilesImp()
		res := response.(httputil.MsgRes)

		if res.Code != 200 {
			t.Fatalf("addLocalFiles: code not correct")
		}

		if !sameMap(srv.Index.List(), testCase.Output.IndexMap) {
			t.Fatalf(
				"addLocalFiles: indexes not identical got: %v want: %v",
				srv.Index.List(),
				testCase.Output.IndexMap,
			)
		}
	}
}
