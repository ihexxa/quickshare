import {
  Response,
  UploadStatusResp,
  ListResp,
  ListUploadingsResp,
  ListSharingsResp,
} from "./";

export interface FilesClientResps {
  createMockRespID?: number;
  createMockResp?: Response;
  deleteMockResp?: Response;
  metadataMockResp?: Response;
  mkdirMockResp?: Response;
  moveMockResp?: Response;
  uploadChunkMockResp?: Response<UploadStatusResp>;
  uploadChunkMockRespID?: number;
  uploadStatusMockResp?: Response<UploadStatusResp>;
  uploadStatusMockRespID?: number;
  listMockResp?: Response<ListResp>;
  listHomeMockResp?: Response<ListResp>;
  listUploadingsMockResp?: Response<ListUploadingsResp>;
  deleteUploadingMockResp?: Response;
  addSharingMockResp?: Response;
  deleteSharingMockResp?: Response;
  listSharingsMockResp?: Response<ListSharingsResp>;
  isSharingMockResp?: Response;
}

export const resps = {
  createMockResp: { status: 200, statusText: "", data: {} },
  deleteMockResp: { status: 200, statusText: "", data: {} },
  metadataMockResp: { status: 200, statusText: "", data: {} },
  mkdirMockResp: { status: 200, statusText: "", data: {} },
  moveMockResp: { status: 200, statusText: "", data: {} },
  uploadChunkMockResp: {
    status: 200,
    statusText: "",
    data: {
      path: "mockPath/file",
      isDir: false,
      fileSize: 5,
      uploaded: 3,
    },
  },
  uploadChunkMockRespID: 0,
  uploadStatusMockResp: {
    status: 200,
    statusText: "",
    data: {
      path: "mockPath/file",
      isDir: false,
      fileSize: 5,
      uploaded: 3,
    },
  },
  uploadStatusMockRespID: 0,
  listMockResp: {
    status: 200,
    statusText: "",
    data: {
      cwd: "mock_cwd",
      metadatas: [
        {
          name: "mock_file",
          size: 5,
          modTime: "0",
          isDir: false,
          sha1: "mock_file_sha1",
        },
        {
          name: "mock_dir",
          size: 0,
          modTime: "0",
          isDir: true,
          sha1: "",
        },
      ],
    },
  },
  listHomeMockResp: {
    status: 200,
    statusText: "",
    data: {
      cwd: "mock_home/files",
      metadatas: [
        {
          name: "mock_file",
          size: 5,
          modTime: "0",
          isDir: false,
          sha1: "mock_file_sha1",
        },
        {
          name: "mock_dir",
          size: 0,
          modTime: "0",
          isDir: true,
          sha1: "",
        },
      ],
    },
  },
  listUploadingsMockResp: {
    status: 200,
    statusText: "",
    data: {
      uploadInfos: [
        {
          realFilePath: "mock_ realFilePath1",
          size: 5,
          uploaded: 3,
        },
        {
          realFilePath: "mock_ realFilePath2",
          size: 5,
          uploaded: 3,
        },
      ],
    },
  },
  deleteUploadingMockResp: { status: 200, statusText: "", data: {} },
  addSharingMockResp: { status: 200, statusText: "", data: {} },
  deleteSharingMockResp: { status: 200, statusText: "", data: {} },
  listSharingsMockResp: {
    status: 200,
    statusText: "",
    data: {
      sharingDirs: ["mock_sharingfolder1", "mock_sharingfolder2"],
    },
  },
  isSharingMockResp: { status: 200, statusText: "", data: {} },
};
export class MockFilesClient {
  private url: string;
  private resps: FilesClientResps;

  constructor(url: string) {
    this.url = url;
    this.resps = resps;
  }

  setMock = (resps: FilesClientResps) => {
    this.resps = resps;
  };

  wrapPromise = (resp: any): Promise<any> => {
    return new Promise<any>((resolve) => {
      resolve(resp);
    });
  };

  create = (filePath: string, fileSize: number): Promise<Response> => {
    return this.wrapPromise(this.resps.createMockResp);
  };

  delete = (filePath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.deleteMockResp);
  };

  metadata = (filePath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.metadataMockResp);
  };

  mkdir = (dirpath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.mkdirMockResp);
  };

  move = (oldPath: string, newPath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.moveMockResp);
  };

  uploadChunk = (
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ): Promise<Response<UploadStatusResp>> => {
    return this.wrapPromise(this.resps.uploadChunkMockResp);
  };

  uploadStatus = (filePath: string): Promise<Response<UploadStatusResp>> => {
    return this.wrapPromise(this.resps.uploadStatusMockResp);
  };

  list = (dirPath: string): Promise<Response<ListResp>> => {
    return this.wrapPromise(this.resps.listMockResp);
  };

  listHome = (): Promise<Response<ListResp>> => {
    return this.wrapPromise(this.resps.listHomeMockResp);
  };

  listUploadings = (): Promise<Response<ListUploadingsResp>> => {
    return this.wrapPromise(this.resps.listUploadingsMockResp);
  };

  deleteUploading = (filePath: string): Promise<Response<Response>> => {
    return this.wrapPromise(this.resps.deleteUploadingMockResp);
  };

  addSharing = (dirPath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.addSharingMockResp);
  };

  deleteSharing = (dirPath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.deleteSharingMockResp);
  };

  listSharings = (): Promise<Response<ListSharingsResp>> => {
    return this.wrapPromise(this.resps.listSharingsMockResp);
  };

  isSharing = (dirPath: string): Promise<Response> => {
    return this.wrapPromise(this.resps.isSharingMockResp);
  };
}
