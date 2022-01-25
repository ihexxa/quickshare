import { ChunkUploader } from "./chunk_uploader";
import {
  FileWorkerReq,
  syncReqKind,
  SyncReq,
  errKind,
  ErrResp,
  ImIdleResp,
  uploadInfoKind,
  imIdleKind,
  UploadInfoResp,
  FileWorkerResp,
  UploadStatus,
  UploadState,
  IChunkUploader,
} from "./interface";

const win: Window = self as any;

export class UploadWorker {
  private uploader: IChunkUploader = new ChunkUploader();
  private cycle: number = 100;
  private working: boolean = false;

  sendEvent = (resp: FileWorkerResp): void => {
    // TODO: make this abstract
    throw new Error("not implemented");
  };

  constructor() {
    win.setInterval(this.checkIdle, this.cycle);
  }

  checkIdle = () => {
    if (this.working) {
      return;
    }

    const resp: ImIdleResp = {
      kind: imIdleKind,
    };
    this.sendEvent(resp);
  };

  setUploader = (uploader: IChunkUploader) => {
    this.uploader = uploader;
  };

  handleUploadStatus = (status: UploadStatus) => {
    if (status.state !== UploadState.Error) {
      const resp: UploadInfoResp = {
        kind: uploadInfoKind,
        filePath: status.filePath,
        uploaded: status.uploaded,
        state: status.state,
        err: "",
      };
      this.sendEvent(resp);
    } else {
      const resp: ErrResp = {
        kind: errKind,
        filePath: status.filePath,
        err: status.err,
      };
      this.sendEvent(resp);
    }
  };

  onMsg = async (event: MessageEvent) => {
    try {
      this.working = true;
      const req = event.data as FileWorkerReq;

      switch (req.kind) {
        case syncReqKind:
          const syncReq = req as SyncReq;

          if (syncReq.created) {
            if (syncReq.file.size === 0) {
              const resp: UploadInfoResp = {
                kind: uploadInfoKind,
                filePath: syncReq.filePath,
                uploaded: 0,
                state: UploadState.Ready,
                err: "",
              };
              this.sendEvent(resp);
            } else {
              const status = await this.uploader.upload(
                syncReq.filePath,
                syncReq.file,
                syncReq.uploaded
              );
              await this.handleUploadStatus(status);
            }
          } else {
            const status = await this.uploader.create(
              syncReq.filePath,
              syncReq.file
            );
            await this.handleUploadStatus(status);
          }
          break;
        default:
          console.error(`unknown worker request(${JSON.stringify(req)})`);
      }
    } finally {
      this.working = false;
    }
  };

  onError = (ev: ErrorEvent) => {
    const errResp: ErrResp = {
      kind: errKind,
      filePath: "unknown",
      err: ev.error,
    };
    this.sendEvent(errResp);
  };
}
