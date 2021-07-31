import axios, { AxiosRequestConfig } from "axios";

export const defaultTimeout = 10000;
export const userIDParam = "uid";

export interface User {
  id: string;
  name: string;
  pwd: string;
  role: string;
}

export interface ListUsersResp {
  users: Array<User>;
}

export interface ListRolesResp {
  roles: Array<string>;
}

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
  cwd: string;
  metadatas: MetadataResp[];
}

export interface UploadInfo {
  realFilePath: string;
  size: number;
  uploaded: number; // TODO: use string instead
}

export interface ListUploadingsResp {
  uploadInfos: UploadInfo[];
}

export interface IUsersClient {
  login: (user: string, pwd: string) => Promise<Response>;
  logout: () => Promise<Response>;
  isAuthed: () => Promise<Response>;
  self: () => Promise<Response>;
  setPwd: (oldPwd: string, newPwd: string) => Promise<Response>;
  forceSetPwd: (userID: string, newPwd: string) => Promise<Response>;
  addUser: (name: string, pwd: string, role: string) => Promise<Response>;
  delUser: (userID: string) => Promise<Response>;
  listUsers: () => Promise<Response>;
  addRole: (role: string) => Promise<Response>;
  delRole: (role: string) => Promise<Response>;
  listRoles: () => Promise<Response>;
}

export interface IFilesClient {
  create: (filePath: string, fileSize: number) => Promise<Response>;
  delete: (filePath: string) => Promise<Response>;
  metadata: (filePath: string) => Promise<Response>;
  mkdir: (dirpath: string) => Promise<Response>;
  move: (oldPath: string, newPath: string) => Promise<Response>;
  uploadChunk: (
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ) => Promise<Response<UploadStatusResp>>;
  uploadStatus: (filePath: string) => Promise<Response<UploadStatusResp>>;
  list: (dirPath: string) => Promise<Response<ListResp>>;
  listHome: () => Promise<Response<ListResp>>;
  listUploadings: () => Promise<Response<ListUploadingsResp>>;
  deleteUploading: (filePath: string) => Promise<Response>;
}

export interface Response<T = any> {
  status: number;
  statusText: string;
  data: T;
}

export const TimeoutResp: Response<any> = {
  status: 408,
  data: {},
  statusText: "Request Timeout",
};

// 6xx are custom errors for expressing errors which can not be expressed by http status code
export const EmptyBodyResp: Response<any> = {
  status: 601,
  data: {},
  statusText: "Empty Response Body",
};

export const FatalErrResp = (errMsg: string): Response<any> => {
  return {
    status: 600,
    data: {},
    statusText: errMsg,
  };
};

export const isFatalErr = (resp: Response<any>): boolean => {
  return resp.status === 600;
};

export class BaseClient {
  protected url: string;

  constructor(url: string) {
    this.url = url;
  }

  async do(config: AxiosRequestConfig): Promise<Response> {
    let returned = false;
    const src = axios.CancelToken.source();

    return new Promise((resolve: (ret: Response) => void) => {
      setTimeout(() => {
        if (!returned) {
          src.cancel("request timeout");
          resolve(TimeoutResp);
        }
      }, defaultTimeout);

      axios({ ...config, cancelToken: src.token })
        .then((resp) => {
          returned = true;
          resolve(resp);
        })
        .catch((e) => {
          const errMsg = e.toString();

          if (errMsg.includes("ERR_EMPTY")) {
            // this means connection is eliminated by server, it may be caused by timeout.
            resolve(EmptyBodyResp);
          } else if (e.response != null) {
            resolve(e.response);
          } else {
            // TODO: check e.request to get more friendly error message
            resolve(FatalErrResp(errMsg));
          }
        });
    });
  }
}
