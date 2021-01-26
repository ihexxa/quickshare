import { FileUploader } from "./uploader";
import {
  FileWorkerReq,
  syncReqKind,
  SyncReq,
  errKind,
  ErrResp,
  uploadInfoKind,
  UploadInfoResp,
  FileWorkerResp,
} from "./interface";

const ctx: Worker = self as any;

export class UploadWorker {
  private file: File = undefined;
  private filePath: string = undefined;
  private uploader: FileUploader = undefined;
  public sendEvent: (resp: FileWorkerResp) => void;
  public makeUploader: (file: File, filePath: string) => FileUploader;
  public startUploader: (file: File, filePath: string) => void;
  public stopUploader: () => void;

  constructor() {
    this.sendEvent = (resp: FileWorkerResp) => {
      ctx.postMessage(resp);
    };
    this.makeUploader = (file: File, filePath: string): FileUploader => {
      return new FileUploader(file, filePath, this.onCb);
    };
    this.startUploader = (file: File, filePath: string) => {
      this.file = file;
      this.filePath = filePath;
      this.uploader = this.makeUploader(file, filePath);
      this.uploader.start();
    };
    this.stopUploader = () => {
      if (this.uploader != null) {
        this.uploader.stop();
      }
    };
  }

  getFilePath = (): string => {
    return this.filePath;
  };

  setFilePath = (fp: string)=> {
    this.filePath = fp;
  };

  onMsg = (ev: MessageEvent) => {
    const req = ev.data as FileWorkerReq;

    switch (req.kind) {
      case syncReqKind:
        // find the first qualified task
        const syncReq = req as SyncReq;
        const infoArray = syncReq.infos;
        for (let i = 0; i < infoArray.length; i++) {
          if (
            infoArray[i].runnable &&
            infoArray[i].uploaded < infoArray[i].size
          ) {
            if (infoArray[i].filePath !== this.filePath) {
              this.stopUploader();
              this.startUploader(infoArray[i].file, infoArray[i].filePath);
            }
            break;
          }
        }
        break;
      default:
        console.log(`unknown worker request(${JSON.stringify(req)})`);
    }
  };

  onError = (ev: ErrorEvent) => {
    const errResp: ErrResp = {
      kind: errKind,
      err: ev.error,
    };
    this.sendEvent(errResp);
  };

  onCb = (
    filePath: string,
    uploaded: number,
    runnable: boolean,
    err: string
  ): void => {
    const uploadInfoResp: UploadInfoResp = {
      kind: uploadInfoKind,
      filePath,
      uploaded,
      runnable,
      err,
    };
    this.sendEvent(uploadInfoResp);
  };
}

const uploadWorker = new UploadWorker();
ctx.addEventListener("message", uploadWorker.onMsg);
ctx.addEventListener("error", uploadWorker.onError);

export default null as any;
