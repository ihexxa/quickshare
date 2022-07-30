import {
  BaseClient,
  FatalErrResp,
  Response,
  UploadStatusResp,
  ListResp,
  ListUploadingsResp,
  ListSharingsResp,
  ListSharingIDsResp,
  GetSharingDirResp,
} from "./";

export const filePathQuery = "fp";
export const listDirQuery = "dp";
export const shareIDQuery = "shid";
export const shareDirQuery = "shdir";
export const keywordQuery = "k";
// TODO: get timeout from server

function translateResp(resp: Response<any>): Response<any> {
  if (resp.status === 500) {
    // TODO: replace following with error code
    if (
      resp.data == null ||
      resp.data === "" ||
      (resp.data.error != null &&
        !resp.data.error.includes("fail to lock the file") &&
        !resp.data.error.includes("offset != uploaded") &&
        !resp.data.error.includes("i/o timeout") &&
        !resp.data.error.includes("too many opened files"))
    ) {
      return FatalErrResp(resp.statusText);
    }
  } else if (resp.status === 404) {
    return FatalErrResp(resp.statusText);
  }

  return resp;
}

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
    })
      .then((resp) => {
        return translateResp(resp);
      })
      .catch((e) => {
        return FatalErrResp(`unknow uploadStatus error ${e.toString()}`);
      });
  };

  uploadStatus = (filePath: string): Promise<Response<UploadStatusResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/files/chunks`,
      params: {
        [filePathQuery]: filePath,
      },
    })
      .then((resp) => {
        return translateResp(resp);
      })
      .catch((e) => {
        return FatalErrResp(`unknow uploadStatus error ${e.toString()}`);
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

  listHome = (): Promise<Response<ListResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/dirs/home`,
      params: {},
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

  addSharing = (dirPath: string): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/fs/sharings`,
      data: {
        SharingPath: dirPath,
      },
    });
  };

  deleteSharing = (dirPath: string): Promise<Response> => {
    return this.do({
      method: "delete",
      url: `${this.url}/v1/fs/sharings`,
      params: {
        [filePathQuery]: dirPath,
      },
    });
  };

  isSharing = (dirPath: string): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/sharings/exist`,
      params: {
        [filePathQuery]: dirPath,
      },
    });
  };

  listSharings = (): Promise<Response<ListSharingsResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/sharings`,
    });
  };

  listSharingIDs = (): Promise<Response<ListSharingIDsResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/sharings/ids`,
    });
  };

  getSharingDir = (shareID: string): Promise<Response<GetSharingDirResp>> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/sharings/dirs`,
      params: {
        [shareIDQuery]: shareID,
      },
    });
  };

  generateHash = (filePath: string): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/fs/hashes/sha1`,
      data: {
        filePath: filePath,
      },
    });
  };

  download = (url: string): Promise<Response> => {
    return this.do({
      method: "get",
      url,
    });
  };

  search = (keywords: string[]): Promise<Response> => {
    const params = new URLSearchParams();
    keywords.forEach(keyword => {
      params.append(keywordQuery, keyword);
    });

    return this.do({
      method: "get",
      url: `${this.url}/v1/fs/search`,
      params,
    });
  };

  reindex = (): Promise<Response> => {
    return this.do({
      method: "put",
      url: `${this.url}/v1/fs/reindex`,
      data: {},
    });
  };
}
