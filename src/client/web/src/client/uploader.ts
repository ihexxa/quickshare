import { List } from "immutable";

import { Response, UnknownErrResp, UploadStatusResp } from "./";
import { FilesClient } from "../client/files";

const defaultChunkLen = 1024 * 1024 * 30; // 15MB/s
const speedDownRatio = 0.5;
const speedUpRatio = 1.1;

export class FileUploader {
  private reader = new FileReader();
  private client = new FilesClient("");
  private file: File;
  private filePath: string;
  private offset: number;
  private chunkLen: number = defaultChunkLen;
  private progressCb: (filePath: string, progress: number) => void;
  private errMsg: string | null;

  constructor(
    file: File,
    filePath: string,
    progressCb: (filePath: string, progress: number) => void
  ) {
    this.file = file;
    this.filePath = filePath;
    this.progressCb = progressCb;
    this.offset = 0;
  }

  err = (): string | null => {
    return this.errMsg;
  }

  start = async (): Promise<boolean> => {
    const resp = await this.client.create(this.filePath, this.file.size);
    switch (resp.status) {
      case 200:
        return await this.upload();
      default:
        this.errMsg = `failed to create ${this.filePath}: status=${resp.statusText}`;
        return false;
    }
  };

  upload = async (): Promise<boolean> => {
    while (
      this.chunkLen > 0 &&
      this.offset >= 0 &&
      this.offset < this.file.size
    ) {
      let uploadPromise = new Promise<Response<UploadStatusResp>>(
        (resolve: (resp: Response<UploadStatusResp>) => void) => {
          this.reader.onerror = (ev: ProgressEvent<FileReader>) => {
            resolve(UnknownErrResp(this.reader.error.toString()));
          };

          this.reader.onloadend = (ev: ProgressEvent<FileReader>) => {
            const dataURL = ev.target.result as string; // readAsDataURL
            const base64Chunk = dataURL.slice(dataURL.indexOf(",") + 1);
            this.client
              .uploadChunk(this.filePath, base64Chunk, this.offset)
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
        const uploadStatusResp = await this.client.uploadStatus(this.filePath);

        console.log(uploadStatusResp.status);

        if (uploadStatusResp.status === 200) {
          this.offset = uploadStatusResp.data.uploaded;
        } else if (uploadStatusResp.status === 600) {
          this.errMsg = "unknown error";
          break
        } else {
          // do nothing and retry
        }
      }

      this.progressCb(this.filePath, Math.ceil(this.offset / this.file.size));
    }

    if (this.chunkLen === 0) {
      this.errMsg = "network is bad, please retry later.";
    }
    return this.offset === this.file.size;
  };
}
