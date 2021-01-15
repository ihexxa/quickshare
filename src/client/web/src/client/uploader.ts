import { IFilesClient } from "../client";
import { FilesClient } from "../client/files";
import { Response, UnknownErrResp, UploadStatusResp } from "./";

// TODO: get settings from server
// TODO: move chunk copying to worker
const defaultChunkLen = 1024 * 512;
const speedDownRatio = 0.5;
const speedUpRatio = 1.1;
const retryLimit = 4;

export interface IFileUploader {
  stop: () => void;
  err: () => string | null;
  setClient: (client: IFilesClient) => void;
  create: (filePath: string, fileSize: number) => Promise<Response>;
  start: () => Promise<boolean>;
  upload: () => Promise<boolean>;
  uploadChunk: (
    filePath: string,
    base64Chunk: string,
    offset: number
  ) => Promise<Response<UploadStatusResp>>;

  uploadStatus: (filePath: string) => Promise<Response<UploadStatusResp>>;
}

export class FileUploader {
  private reader = new FileReader();
  private client: IFilesClient = new FilesClient("");
  private chunkLen: number = defaultChunkLen;
  private file: File;
  private offset: number;
  private filePath: string;
  private errMsg: string | null;
  private shouldStop: boolean;
  private progressCb: (filePath: string, uploaded: number) => void;

  constructor(
    file: File,
    filePath: string,
    progressCb?: (filePath: string, uploaded: number) => void
  ) {
    this.file = file;
    this.filePath = filePath;
    this.progressCb = progressCb;
    this.offset = 0;
    this.shouldStop = false;
  }

  stop = () => {
    this.shouldStop = true;
  };

  err = (): string | null => {
    return this.errMsg;
  };

  setClient = (client: IFilesClient) => {
    this.client = client;
  };

  create = async (filePath: string, fileSize: number): Promise<Response> => {
    return this.client.create(filePath, fileSize);
  };

  uploadChunk = async (
    filePath: string,
    base64Chunk: string,
    offset: number
  ): Promise<Response<UploadStatusResp>> => {
    return this.client.uploadChunk(filePath, base64Chunk, offset);
  };

  uploadStatus = async (
    filePath: string
  ): Promise<Response<UploadStatusResp>> => {
    return this.client.uploadStatus(filePath);
  };

  start = async (): Promise<boolean> => {
    let resp: Response;
    for (let i = 0; i < retryLimit; i++) {
      resp = await this.create(this.filePath, this.file.size);
      if (resp.status === 200) {
        return await this.upload();
      }
    }

    this.errMsg = `failed to create ${this.filePath}: status=${resp.statusText}`;
    return false;
  };

  upload = async (): Promise<boolean> => {
    while (
      this.chunkLen > 0 &&
      this.offset >= 0 &&
      this.offset < this.file.size &&
      !this.shouldStop
    ) {
      const uploadPromise = new Promise<Response<UploadStatusResp>>(
        (resolve: (resp: Response<UploadStatusResp>) => void) => {
          this.reader.onerror = (ev: ProgressEvent<FileReader>) => {
            resolve(UnknownErrResp(this.reader.error.toString()));
          };

          this.reader.onloadend = (ev: ProgressEvent<FileReader>) => {
            const dataURL = ev.target.result as string; // readAsDataURL
            const base64Chunk = dataURL.slice(dataURL.indexOf(",") + 1);
            this.uploadChunk(this.filePath, base64Chunk, this.offset)
              .then((resp: Response<UploadStatusResp>) => {
                resolve(resp);
              })
              .catch((e) => {
                resolve(UnknownErrResp(e.toString()));
              });
          };
        }
      );

      const chunkRightPos =
        this.offset + this.chunkLen > this.file.size
          ? this.file.size
          : this.offset + this.chunkLen;
      const blob = this.file.slice(this.offset, chunkRightPos);
      this.reader.readAsDataURL(blob);

      const uploadResp = await uploadPromise;

      if (uploadResp.status === 200 && uploadResp.data != null) {
        this.offset = uploadResp.data.uploaded;
        this.chunkLen = Math.ceil(this.chunkLen * speedUpRatio);
      } else {
        this.errMsg = uploadResp.statusText;
        this.chunkLen = Math.ceil(this.chunkLen * speedDownRatio);

        let uploadStatusResp: Response<UploadStatusResp> = undefined;
        try {
          uploadStatusResp = await this.uploadStatus(this.filePath);
        } catch (e) {
          if (uploadStatusResp == null) {
            this.errMsg = `${this.errMsg}; unknown error: empty uploadStatus response`;
            break;
          } else if (uploadStatusResp.status === 500) {
            if (
              !uploadStatusResp.statusText.includes("fail to lock the file") &&
              uploadStatusResp.statusText !== ""
            ) {
              this.errMsg = `${this.errMsg}; unknown error: ${uploadStatusResp.statusText}`;
              break;
            }
          } else if (uploadStatusResp.status === 600) {
            this.errMsg = `${this.errMsg}; unknown error: ${uploadStatusResp.statusText}`;
            break;
          } else {
            // ignore error and retry
          }
        }

        if (uploadStatusResp.status === 200) {
          this.offset = uploadStatusResp.data.uploaded;
        }
      }

      if (this.progressCb != null) {
        this.progressCb(this.filePath, this.offset);
      }
    }

    if (this.chunkLen === 0) {
      this.errMsg = "network is poor, please retry later.";
    } else if (this.shouldStop) {
      this.errMsg = "uploading is stopped";
    }
    return this.offset === this.file.size;
  };
}
