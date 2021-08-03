import { FilesClient } from "../client/files";
import { IFilesClient, Response, isFatalErr } from "../client";

// TODO: get settings from server
// TODO: move chunk copying to worker
const defaultChunkLen = 1024 * 1024 * 1;
const speedDownRatio = 0.5;
const speedUpRatio = 1.05;
const createRetryLimit = 2;
const uploadRetryLimit = 1024;
const backoffMax = 2000;

export interface IFileUploader {
  stop: () => void;
  err: () => string | null;
  setClient: (client: IFilesClient) => void;
  start: () => Promise<boolean>;
  upload: () => Promise<boolean>;
}

export interface ReaderResult {
  chunk?: string;
  err?: Error;
}

export class FileUploader {
  private reader = new FileReader();
  private client: IFilesClient = new FilesClient("");

  private file: File;
  private filePath: string;
  private chunkLen: number = defaultChunkLen;
  private offset: number = 0;

  private errMsgs: string[] = new Array<string>();
  private isOn: boolean = true;
  private progressCb: (
    filePath: string,
    uploaded: number,
    runnable: boolean,
    err: string
  ) => void;

  constructor(
    file: File,
    filePath: string,
    progressCb: (
      filePath: string,
      uploaded: number,
      runnable: boolean,
      err: string
    ) => void
  ) {
    this.file = file;
    this.filePath = filePath;
    this.progressCb = progressCb;
  }

  getOffset = (): number => {
    return this.offset;
  };

  stop = () => {
    this.isOn = false;
  };

  err = (): string | null => {
    return this.errMsgs.length === 0 ? null : this.errMsgs.reverse().join(";");
  };

  setClient = (client: IFilesClient) => {
    this.client = client;
  };

  backOff = async (): Promise<void> => {
    return new Promise((resolve) => {
      const delay = Math.floor(Math.random() * backoffMax);
      setTimeout(resolve, delay);
    });
  };

  start = async (): Promise<boolean> => {
    let resp: Response;

    for (let i = 0; i < createRetryLimit; i++) {
      try {
        resp = await this.client.create(this.filePath, this.file.size);
        if (resp.status === 200 || resp.status === 304) {
          return await this.upload();
        }
      } catch (e) {
        await this.backOff();
        console.error(e);
      }
    }

    this.errMsgs.push(
      `failed to create ${this.filePath}: status=${resp.statusText}`
    );
    return false;
  };

  upload = async (): Promise<boolean> => {
    while (
      this.chunkLen > 0 &&
      this.offset >= 0 &&
      this.offset < this.file.size &&
      this.isOn
    ) {
      const readerPromise = new Promise<ReaderResult>(
        (resolve: (result: ReaderResult) => void) => {
          this.reader.onerror = (_: ProgressEvent<FileReader>) => {
            resolve({ err: this.reader.error });
          };

          this.reader.onloadend = (ev: ProgressEvent<FileReader>) => {
            const dataURL = ev.target.result as string; // readAsDataURL
            const base64Chunk = dataURL.slice(dataURL.indexOf(",") + 1); // remove prefix
            resolve({ chunk: base64Chunk });
          };
        }
      );

      const chunkRightPos =
        this.offset + this.chunkLen > this.file.size
          ? this.file.size
          : this.offset + this.chunkLen;
      const blob = this.file.slice(this.offset, chunkRightPos);
      this.reader.readAsDataURL(blob);

      const result = await readerPromise;
      if (result.err != null) {
        this.errMsgs.push(result.err.toString());
        break;
      }

      try {
        const uploadResp = await this.client.uploadChunk(
          this.filePath,
          result.chunk,
          this.offset
        );

        if (uploadResp.status === 200 && uploadResp.data != null) {
          this.offset = uploadResp.data.uploaded;
          this.chunkLen = Math.ceil(this.chunkLen * speedUpRatio);
        } else if (isFatalErr(uploadResp)) {
          this.errMsgs.push(`failed to upload chunk: ${uploadResp.statusText}`);
          break;
        } else {
          // this.errMsgs.push(uploadResp.statusText);
          this.chunkLen = Math.ceil(this.chunkLen * speedDownRatio);

          const uploadStatusResp = await this.client.uploadStatus(
            this.filePath
          );
          if (uploadStatusResp.status === 200) {
            this.offset = uploadStatusResp.data.uploaded;
          } else if (isFatalErr(uploadStatusResp)) {
            this.errMsgs.push(
              `failed to get upload status: ${uploadStatusResp.statusText}`
            );
            break;
          }

          await this.backOff();
        }
      } catch (e) {
        this.errMsgs.push(e.toString());
        break;
      }

      this.progressCb(this.filePath, this.offset, true, this.err());
    }

    if (this.chunkLen === 0) {
      this.errMsgs.push(
        "the network condition is poor, please retry later."
      );
    } else if (!this.isOn) {
      this.errMsgs.push("uploading is stopped");
    }

    this.progressCb(this.filePath, this.offset, false, this.err());
    return this.offset === this.file.size;
  };
}
