import {
  Response,
  UploadStatusResp,
  ListResp,
  ListUploadingsResp,
} from "./";

export class FilesClient {
  private url: string;

  private createMockRespID: number = 0;
  private createMockResps: Array<Promise<Response>>;
  private deleteMockResp: Promise<Response>;
  private metadataMockResp: Promise<Response>;
  private mkdirMockResp: Promise<Response>;
  private moveMockResp: Promise<Response>;
  private uploadChunkMockResps: Array<Promise<Response<UploadStatusResp>>>;
  private uploadChunkMockRespID: number = 0;
  private uploadStatusMockResps: Array<Promise<Response<UploadStatusResp>>>;
  private uploadStatusMockRespID: number = 0;
  private listMockResp: Promise<Response<ListResp>>;
  private listUploadingsMockResp: Promise<Response<ListUploadingsResp>>;
  private deleteUploadingMockResp: Promise<Response>;

  constructor(url: string) {
    this.url = url;
  }

  createMock = (resps: Array<Promise<Response>>) => {
    this.createMockResps = resps;
  };

  deleteMock = (resp: Promise<Response>) => {
    this.deleteMockResp = resp;
  };

  metadataMock = (resp: Promise<Response>) => {
    this.metadataMockResp = resp;
  };

  mkdirMock = (resp: Promise<Response>) => {
    this.mkdirMockResp = resp;
  };

  moveMock = (resp: Promise<Response>) => {
    this.moveMockResp = resp;
  };

  uploadChunkMock = (resps: Array<Promise<Response<UploadStatusResp>>>) => {
    this.uploadChunkMockResps = resps;
  };

  uploadStatusMock = (resps: Array<Promise<Response<UploadStatusResp>>>) => {
    this.uploadStatusMockResps = resps;
  };

  listMock = (resp: Promise<Response<ListResp>>) => {
    this.listMockResp = resp;
  };

  listUploadingsMock = (resp: Promise<Response<ListUploadingsResp>>) => {
    this.listUploadingsMockResp = resp;
  }

  deleteUploadingMock = (resp: Promise<Response>) => {
    this.deleteUploadingMockResp = resp;
  }

  create = (filePath: string, fileSize: number): Promise<Response> => {
    if (this.createMockRespID < this.createMockResps.length) {
      return this.createMockResps[this.createMockRespID++];
    }
    throw new Error(`this.createMockRespID (${this.createMockRespID}) out of bound: ${this.createMockResps.length}`);
  };

  delete = (filePath: string): Promise<Response> => {
    return this.deleteMockResp;
  };

  metadata = (filePath: string): Promise<Response> => {
    return this.metadataMockResp;
  };

  mkdir = (dirpath: string): Promise<Response> => {
    return this.mkdirMockResp;
  };

  move = (oldPath: string, newPath: string): Promise<Response> => {
    return this.moveMockResp;
  };

  uploadChunk = (
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ): Promise<Response<UploadStatusResp>> => {
    if (this.uploadChunkMockRespID < this.uploadChunkMockResps.length) {
      return this.uploadChunkMockResps[this.uploadChunkMockRespID++];
    }
    throw new Error(`this.uploadChunkMockRespID (${this.uploadChunkMockRespID}) out of bound: ${this.uploadChunkMockResps.length}`);
  };

  uploadStatus = (filePath: string): Promise<Response<UploadStatusResp>> => {
    if (this.uploadStatusMockRespID < this.uploadStatusMockResps.length) {
      return this.uploadStatusMockResps[this.uploadStatusMockRespID++];
    }
    throw new Error(`this.uploadStatusMockRespID (${this.uploadStatusMockRespID}) out of bound: ${this.uploadStatusMockResps.length}`);
  };

  list = (dirPath: string): Promise<Response<ListResp>> => {
    return this.listMockResp;
  };

  listUploadings = (): Promise<Response<ListUploadingsResp>> => {
    return this.listUploadingsMockResp;
  };

  deleteUploading = (filePath:string): Promise<Response<Response>> => {
    return this.deleteUploadingMockResp;
  };
}
