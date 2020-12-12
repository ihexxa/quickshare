import axios, { AxiosRequestConfig, AxiosResponse } from "axios";

const filePathQuery = "fp";
const listDirQuery =  "dp";

export interface MetadataResp {
  name: string;
  size: number;
  modTime: string;
  isDir: boolean;
}

export interface UploadStatusResp {
  path: string;
  isDir: boolean;
  fileSize: number;
  uploaded: number;
}

export interface ListResp {
  metadatas: MetadataResp[];
}

export class FilesClient {
  private url: string;

  constructor(url: string) {
    this.url = url;
  }

  async do(config: AxiosRequestConfig): Promise<AxiosResponse<any> | null> {
    try {
      return await axios(config);
    } catch (e) {
      return null;
    }
  }

  async create(filePath: string, fileSize: number): Promise<number> {
    const resp = await this.do({
      method: "post",
      url: `${this.url}/v1/fs/files`,
      data: {
        path: filePath,
        fileSize: fileSize,
      },
    });
    return resp != null ? resp.status : 500;
  }

  async delete(filePath: string): Promise<number> {
    const resp = await this.do({
      method: "delete",
      url: `${this.url}/v1/fs/files`,
      params: {
        [filePathQuery]: filePath,
      },
    });
    return resp != null ? resp.status : 500;
  }

  async metadata(filePath: string): Promise<MetadataResp | null> {
    const resp = await this.do({
      method: "get",
      url: `${this.url}/v1/fs/metadata`,
      params: {
        [filePathQuery]: filePath,
      },
    });

    return resp != null ? resp.data : null;
  }

  async mkdir(dirpath: string): Promise<number | null> {
    const resp = await this.do({
      method: "post",
      url: `${this.url}/v1/fs/dirs`,
      data: {
        path: dirpath,
      },
    });

    return resp.status;
  }

  async move(oldPath: string, newPath: string): Promise<number> {
    const resp = await this.do({
      method: "patch",
      url: `${this.url}/v1/fs/files/move`,
      data: {
        oldPath,
        newPath,
      },
    });

    return resp != null ? resp.status : 500;
  }

  async uploadChunk(
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ): Promise<UploadStatusResp | null> {
    const resp = await this.do({
      method: "patch",
      url: `${this.url}/v1/fs/files/chunks`,
      data: {
        path: filePath,
        content,
        offset,
      },
    });

    return resp != null ? resp.data : null;
  }

  async uploadStatus(filePath: string): Promise<UploadStatusResp | null> {
    const resp = await this.do({
      method: "get",
      url: `${this.url}/v1/fs/files/chunks`,
      params: {
        [filePathQuery]: filePath,
      },
    });

    return resp != null ? resp.data : null;
  }

  async list(dirPath: string): Promise<ListResp | null> {
    const resp = await this.do({
      method: "get",
      url: `${this.url}/v1/fs/dirs`,
      params: {
        [listDirQuery]: dirPath,
      },
    });

    return resp != null ? resp.data : null;
  }
}
