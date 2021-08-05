import { ChunkUploader } from "./chunk_uploader";
import {
  FileWorkerReq,
  syncReqKind,
  SyncReq,
  errKind,
  ErrResp,
  uploadInfoKind,
  UploadInfoResp,
  FileWorkerResp,
  UploadStatus,
  UploadState,
  IChunkUploader,
} from "./interface";

export class UploadWorker {
  private uploader: IChunkUploader = new ChunkUploader();
  sendEvent = (resp: FileWorkerResp): void => {
    // TODO: make this abstract
    throw new Error("not implemented");
  };

  constructor() {}

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

  onMsg = (event: MessageEvent) => {
    const req = event.data as FileWorkerReq;

    switch (req.kind) {
      case syncReqKind:
        const syncReq = req as SyncReq;

        if (syncReq.created) {
          this.uploader
            .upload(syncReq.filePath, syncReq.file, syncReq.uploaded)
            .then(this.handleUploadStatus);
        } else {
          this.uploader
            .create(syncReq.filePath, syncReq.file)
            .then(this.handleUploadStatus);
        }

        break;
      default:
        console.error(`unknown worker request(${JSON.stringify(req)})`);
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
