import { FilesClient } from "../client/files";
import { IFilesClient, Response, isFatalErr } from "../client";
import { UploadStatus, UploadState } from "./interface";

// TODO: get settings from server
// TODO: move chunk copying to worker
const defaultChunkLen = 1024 * 1024 * 1;
const speedDownRatio = 0.5;
const speedUpRatio = 1.05;
const createRetryLimit = 2;
const uploadRetryLimit = 1024;
const backoffMax = 2000;

export interface ReaderResult {
  chunk?: string;
  err?: Error;
}

export class ChunkUploader {
  private reader = new FileReader();
  private client: IFilesClient = new FilesClient("");

  private chunkLen: number = defaultChunkLen;

  constructor() {}

  setClient = (client: IFilesClient) => {
    this.client = client;
  };

  backOff = async (): Promise<void> => {
    return new Promise((resolve) => {
      const delay = Math.floor(Math.random() * backoffMax);
      setTimeout(resolve, delay);
    });
  };

  create = async (filePath: string, file: File): Promise<UploadStatus> => {
    let resp: Response;

    for (let i = 0; i < createRetryLimit; i++) {
      try {
        resp = await this.client.create(filePath, file.size);
        if (resp.status === 200 || resp.status === 304) {
          return {
            filePath,
            uploaded: 0,
            state: UploadState.Created,
            err: "",
          };
        }
      } catch (e) {
        await this.backOff();
        console.error(e);
      }
    }

    return {
      filePath,
      uploaded: 0,
      state: UploadState.Error,
      err: `failed to create ${filePath}: status=${resp.statusText}`,
    };
  };

  upload = async (
    filePath: string,
    file: File,
    uploaded: number
  ): Promise<UploadStatus> => {
    if (this.chunkLen === 0) {
      this.chunkLen = 1; // reset it to 1B
    } else if (uploaded > file.size) {
      return {
        filePath,
        uploaded,
        state: UploadState.Error,
        err: "uploaded is greater than file size",
      };
    }

    const readerPromise = new Promise<ReaderResult>(
      (resolve: (result: ReaderResult) => void) => {
        this.reader.onerror = (_: ProgressEvent<FileReader>) => {
          resolve({ err: this.reader.error });
        };

        this.reader.onloadend = (ev: ProgressEvent<FileReader>) => {
          const dataURL = ev.target.result as string; // readAsDataURL
          const base64Chunk = dataURL.slice(dataURL.indexOf(",") + 1);
          resolve({ chunk: base64Chunk });
        };
      }
    );

    const chunkRightPos =
      uploaded + this.chunkLen > file.size
        ? file.size
        : uploaded + this.chunkLen;
    const blob = file.slice(uploaded, chunkRightPos);
    this.reader.readAsDataURL(blob);

    const result = await readerPromise;
    if (result.err != null) {
      return {
        filePath,
        uploaded,
        state: UploadState.Error,
        err: result.err.toString(),
      };
    }

    try {
      const uploadResp = await this.client.uploadChunk(
        filePath,
        result.chunk,
        uploaded
      );

      if (uploadResp.status === 200 && uploadResp.data != null) {
        this.chunkLen = Math.ceil(this.chunkLen * speedUpRatio);
        return {
          filePath,
          uploaded: uploadResp.data.uploaded,
          state: UploadState.Ready,
          err: "",
        };
      } else if (isFatalErr(uploadResp)) {
        return {
          filePath,
          uploaded,
          state: UploadState.Error,
          err: `failed to upload chunk: ${uploadResp.statusText}`,
        };
      }

      this.chunkLen = Math.ceil(this.chunkLen * speedDownRatio);
      await this.backOff();

      const uploadStatusResp = await this.client.uploadStatus(filePath);
      return uploadStatusResp.status === 200
        ? {
            filePath,
            uploaded: uploadStatusResp.data.uploaded,
            state: UploadState.Ready,
            err: "",
          }
        : {
            filePath,
            uploaded: uploaded,
            state: UploadState.Error,
            err: `failed to get upload status: ${uploadStatusResp.statusText}`,
          };
    } catch (e) {
      return {
        filePath,
        uploaded: uploaded,
        state: UploadState.Error,
        err: `[chunk uploader]: ${e.toString()}`,
      };
    }
  };
}