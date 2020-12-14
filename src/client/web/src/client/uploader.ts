import { filesClient } from "../client";
import { UploadStatusResp } from "./files";

const chunkLen = 1024 * 500;

export class FileUploader {
  private reader = new FileReader();
  private client = filesClient;
  private file: File;
  private filePath: string;
  private offset: number;
  private endCb: (err: Error) => void;

  constructor(file: File, filePath: string, endCb: (err: Error) => void) {
    this.file = file;
    this.filePath = filePath;
    this.offset = 0;
    this.endCb = endCb;
  }

  start = async () => {
    console.log("started");
    const status = await this.client.create(this.filePath, this.file.size);
    console.log(status);
    switch (status) {
      case 200:
        await this.upload();
        break;
      default:
      // alert
    }
  };

  upload = async () => {
    // console.log("filesize", chunkRightPos, this.file.size);
    // this.reader.readAsArrayBuffer(this.file.slice(self.offset, chunkRightPos));

    while (this.offset >= 0 && this.offset < this.file.size) {
      let uploadPromise = new Promise<number>((resolve, reject) => {
        this.reader.onerror = (ev: ProgressEvent<FileReader>) => {
          // log error
          reject(-1);
        };

        this.reader.onloadend = (ev: ProgressEvent<FileReader>) => {
          const dataURL = ev.target.result as string; // readAsDataURL
          const base64Chunk = dataURL.slice(dataURL.indexOf(",") + 1);
          this.client
            .uploadChunk(this.filePath, base64Chunk, this.offset)
            .then((resp: UploadStatusResp) => {
              if (resp == null) {
                reject(-1);
              }
              resolve(resp.uploaded);
            })
            .catch((e) => {
              // alert
              reject(-1);
            });
        };
      });

      const chunkRightPos =
        this.offset + chunkLen > this.file.size
          ? this.file.size
          : this.offset + chunkLen;

      console.log(this.offset, chunkRightPos);
      const blob = this.file.slice(this.offset, chunkRightPos);
      this.reader.readAsDataURL(blob);
      this.offset = await uploadPromise;
    }

    this.endCb(null);
  };
}
