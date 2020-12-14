import {
  FilesClient,
  MetadataResp,
  UploadStatusResp,
  ListResp,
} from "../client/files";
import { UsersClient } from "../client/users";

export interface IUsersClient {
  login: (user: string, pwd: string) => Promise<number>;
  logout: () => Promise<number>;
  isAuthed: () => Promise<number>;
  setPwd: (oldPwd: string, newPwd: string) => Promise<number>;
}

export interface IFilesClient {
  create: (filePath: string, fileSize: number) => Promise<number>;
  delete: (filePath: string) => Promise<number>;
  metadata: (filePath: string) => Promise<MetadataResp | null>;
  mkdir: (dirpath: string) => Promise<number | null>;
  move: (oldPath: string, newPath: string) => Promise<number>;
  uploadChunk(
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ): Promise<UploadStatusResp | null>;
  uploadStatus: (filePath: string) => Promise<UploadStatusResp | null>;
  list: (dirPath: string) => Promise<ListResp | null>;
}

export const filesClient = new FilesClient("");
export const usersClient = new UsersClient("");
