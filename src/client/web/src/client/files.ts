import {
  BaseClient,
  Response,
  UploadStatusResp,
  ListResp,
  ListUploadingsResp,
} from "./";

const filePathQuery = "fp";
const listDirQuery = "dp";
// TODO: get timeout from server

export class FilesClient extends BaseClient {
  constructor(url: string) {
    super(url);
  }

  create = (filePath: string, fileSize: number): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/fs/files`,
      data: {
        path: filePath,
        fileSize: fileSize,
      },
    });
  };

  delete = (filePath: string): Promise<Response> => {
    return this.do({
      method: "delete",
      url: `${this.url}/v1/fs/files`,
      params: {
        [filePathQuery]: filePath,
      },
    });
  };

  metadata = (filePath: string): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/metadata`,
      params: {
        [filePathQuery]: filePath,
      },
    });
  };

  mkdir = (dirpath: string): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/fs/dirs`,
      data: {
        path: dirpath,
      },
    });
  };

  move = (oldPath: string, newPath: string): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/fs/files/move`,
      data: {
        oldPath,
        newPath,
      },
    });
  };

  uploadChunk = (
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ): Promise<Response<UploadStatusResp>> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/fs/files/chunks`,
      data: {
        path: filePath,
        content,
        offset,
      },
    });
  };

  uploadStatus = (filePath: string): Promise<Response<UploadStatusResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/files/chunks`,
      params: {
        [filePathQuery]: filePath,
      },
    });
  };

  list = (dirPath: string): Promise<Response<ListResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/dirs`,
      params: {
        [listDirQuery]: dirPath,
      },
    });
  };

  listUploadings = (): Promise<Response<ListUploadingsResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/uploadings`,
    });
  };

  deleteUploading = (filePath: string): Promise<Response> => {
    return this.do({
      method: "delete",
      url: `${this.url}/v1/fs/uploadings`,
      params: {
        [filePathQuery]: filePath,
      },
    });
  };
}
