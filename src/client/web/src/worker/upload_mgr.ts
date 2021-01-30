import { Map } from "immutable";

import {
  FileWorkerReq,
  FileWorkerResp,
  UploadInfoResp,
  ErrResp,
  UploadEntry,
  syncReqKind,
  errKind,
  uploadInfoKind,
} from "./interface";
import { FgWorker } from "./upload.fgworker";

const win: Window = self as any;

export interface IWorker {
  onmessage: (event: MessageEvent) => void;
  postMessage: (event: FileWorkerReq) => void;
}

export class UploadMgr {
  private infos = Map<string, UploadEntry>();
  private worker: IWorker;
  private intervalID: number;
  private cycle: number = 500;
  private statusCb = (infos: Map<string, UploadEntry>): void => {};

  constructor(worker: IWorker) {
    this.worker = worker;
    // TODO: fallback to normal if Web Worker is not available
    this.worker.onmessage = this.respHandler;

    const syncing = () => {
      this.worker.postMessage({
        kind: syncReqKind,
        infos: this.infos.valueSeq().toArray(),
      });
    };
    this.intervalID = win.setInterval(syncing, this.cycle);
  }

  destory = () => {
    win.clearInterval(this.intervalID);
  };

  _setInfos = (infos: Map<string, UploadEntry>) => {
    this.infos = infos;
  };

  setCycle = (ms: number) => {
    this.cycle = ms;
  };

  getCycle = (): number => {
    return this.cycle;
  };

  setStatusCb = (cb: (infos: Map<string, UploadEntry>) => void) => {
    this.statusCb = cb;
  };

  add = (file: File, filePath: string) => {
    const entry = this.infos.get(filePath);
    if (entry == null) {
      // new uploading
      this.infos = this.infos.set(filePath, {
        file: file,
        filePath: filePath,
        size: file.size,
        uploaded: 0,
        runnable: true,
        err: "",
      });
    } else {
      // restart the uploading
      this.infos = this.infos.set(filePath, {
        ...entry,
        runnable: true,
      });
    }
  };

  stop = (filePath: string) => {
    const entry = this.infos.get(filePath);
    if (entry != null) {
      this.infos = this.infos.set(filePath, {
        ...entry,
        runnable: false,
      });
      console.log("stopped", filePath);
    } else {
      alert(`failed to stop uploading ${filePath}: not found`);
    }
  };

  delete = (filePath: string) => {
    this.stop(filePath);
    this.infos = this.infos.delete(filePath);
  };

  list = (): Map<string, UploadEntry> => {
    return this.infos;
  };

  respHandler = (event: MessageEvent) => {
    const resp = event.data as FileWorkerResp;

    switch (resp.kind) {
      case errKind:
        // TODO: refine this
        const errResp = resp as ErrResp;
        console.error(`respHandler: ${errResp}`);
        break;
      case uploadInfoKind:
        const infoResp = resp as UploadInfoResp;
        const entry = this.infos.get(infoResp.filePath);

        if (entry != null) {
          if (infoResp.uploaded === entry.size) {
            this.infos = this.infos.delete(infoResp.filePath);
          } else {
            this.infos = this.infos.set(infoResp.filePath, {
              ...entry,
              uploaded: infoResp.uploaded,
              runnable: infoResp.runnable,
              err: infoResp.err,
            });
          }

          // call back to update the info
          this.statusCb(this.infos);
        } else {
          // TODO: refine this
          console.error(
            `respHandler: fail to found: file(${
              infoResp.filePath
            }) infos(${this.infos.toObject()})`
          );
        }
        break;
      default:
        console.error(`respHandler: response kind not found: ${resp}`);
    }
  };
}

export let uploadMgr = new UploadMgr(new FgWorker());
export const initUploadMgr = (worker: IWorker): UploadMgr => {
  uploadMgr = new UploadMgr(worker);
  return uploadMgr;
};
export const Up = (): UploadMgr => {
  return uploadMgr;
};
export const FgUp = (): UploadMgr => {
  return new UploadMgr(new FgWorker());
};
