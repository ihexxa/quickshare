import axios, { AxiosRequestConfig } from "axios";
import { List } from "immutable";

export const defaultTimeout = 10000;
export const userIDParam = "uid";

export const roleAdmin = "admin";
export const roleUser = "user";
export const roleVisitor = "visitor";
export const visitorID = "1";

export interface Quota {
  spaceLimit: string;
  uploadSpeedLimit: number;
  downloadSpeedLimit: number;
}

export interface BgConfig {
  url: string;
  repeat: string;
  position: string;
  align: string;
  bgColor: string;
}

export interface Preferences {
  bg: BgConfig;
  cssURL: string;
  lanPackURL: string;
  lan: string;
  theme: string;
  avatar: string;
  email: string;
}

export interface User {
  id: string;
  name: string;
  pwd: string;
  role: string;
  quota: Quota;
  usedSpace: string;
  preferences: Preferences | undefined;
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
  sha1: string;
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

export interface ListSharingsResp {
  sharingDirs: string[];
}

export interface ListSharingIDsResp {
  IDs: Map<string, string>;
}

export interface GetSharingDirResp {
  sharingDir: string;
}

export interface SearchItemsResp {
  results: string[];
}

export interface ClientConfig {
  siteName: string;
  siteDesc: string;
  bg: BgConfig;
  allowSetBg: boolean;
  autoTheme: boolean;
}

export interface ClientConfigMsg {
  clientCfg: ClientConfig;
  captchaEnabled?: boolean;
}

export interface ClientErrorReport {
  report: string;
  version: string;
}

export interface IUsersClient {
  login: (
    user: string,
    pwd: string,
    captchaId: string,
    captchaInput: string
  ) => Promise<Response>;
  logout: () => Promise<Response>;
  isAuthed: () => Promise<Response>;
  self: () => Promise<Response>;
  setPwd: (oldPwd: string, newPwd: string) => Promise<Response>;
  setUser: (id: string, role: string, quota: Quota) => Promise<Response>;
  forceSetPwd: (userID: string, newPwd: string) => Promise<Response>;
  addUser: (name: string, pwd: string, role: string) => Promise<Response>;
  delUser: (userID: string) => Promise<Response>;
  listUsers: () => Promise<Response>;
  addRole: (role: string) => Promise<Response>;
  delRole: (role: string) => Promise<Response>;
  listRoles: () => Promise<Response>;
  getCaptchaID: () => Promise<Response>;
  setPreferences: (prefers: Preferences) => Promise<Response>;
  resetUsedSpace: (userID: string) => Promise<Response>;
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
  addSharing: (dirPath: string) => Promise<Response>;
  deleteSharing: (dirPath: string) => Promise<Response>;
  isSharing: (dirPath: string) => Promise<Response>;
  listSharings: () => Promise<Response<ListSharingsResp>>;
  listSharingIDs: () => Promise<Response<ListSharingIDsResp>>;
  getSharingDir: (shareID: string) => Promise<Response<GetSharingDirResp>>;
  generateHash: (filePath: string) => Promise<Response>;
  download: (url: string) => Promise<Response>;
  search: (keywords: string[]) => Promise<Response<SearchItemsResp>>;
  reindex: () => Promise<Response>;
}

export interface ISettingsClient {
  health: () => Promise<Response>;
  getClientCfg: () => Promise<Response>;
  setClientCfg: (cfg: ClientConfigMsg) => Promise<Response>;
  reportErrors: (reports: List<ClientErrorReport>) => Promise<Response>;
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
          returned = true;
          src.cancel("request timeout");
          resolve(TimeoutResp);
        }
      }, defaultTimeout);

      axios({ ...config, cancelToken: src.token })
        .then((resp) => {
          if (!returned) {
            returned = true;
            resolve(resp);
          }
        })
        .catch((e) => {
          returned = true;
          const errMsg = e.toString();

          if (errMsg.includes("ERR_EMPTY")) {
            // TODO: check if this is compatible with all browsers
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
