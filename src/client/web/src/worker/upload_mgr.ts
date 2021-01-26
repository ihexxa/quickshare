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

const win = self as any;

export interface IWorker {
  onmessage: (event: MessageEvent) => void;
  postMessage: (event: FileWorkerReq) => void;
}

export class UploadMgr {
  private static infos = Map<string, UploadEntry>();
  private static worker: IWorker;
  private static intervalID: NodeJS.Timeout;
  private static cycle: number = 500;
  private static statusCb = (infos: Map<string, UploadEntry>):void => {};

  static _setInfos = (infos: Map<string, UploadEntry>) => {
    UploadMgr.infos = infos;
  };

  static setCycle = (ms: number) => {
    UploadMgr.cycle = ms;
  };

  static getCycle = (): number => {
    return UploadMgr.cycle;
  };

  static setStatusCb = (cb: (infos: Map<string, UploadEntry>) => void) => {
    UploadMgr.statusCb = cb;
  }

  static init = (worker: IWorker) => {
    UploadMgr.worker = worker;
    // TODO: fallback to normal if Web Worker is not available
    UploadMgr.worker.onmessage = UploadMgr.respHandler;

    const syncing = () => {
      UploadMgr.worker.postMessage({
        kind: syncReqKind,
        infos: UploadMgr.infos.valueSeq().toArray(),
      });
    };
    UploadMgr.intervalID = win.setInterval(syncing, UploadMgr.cycle);
  };

  static destory = () => {
    win.clearInterval(UploadMgr.intervalID);
  };

  static add = (file: File, filePath: string) => {
    const entry = UploadMgr.infos.get(filePath);
    if (entry == null) {
      // new uploading
      UploadMgr.infos = UploadMgr.infos.set(filePath, {
        file: file,
        filePath: filePath,
        size: file.size,
        uploaded: 0,
        runnable: true,
        err: "",
      });
    } else {
      // restart the uploading
      UploadMgr.infos = UploadMgr.infos.set(filePath, {
        ...entry,
        runnable: true,
      });
    }
  };

  static stop = (filePath: string) => {
    const entry = UploadMgr.infos.get(filePath);
    if (entry != null) {
      UploadMgr.infos = UploadMgr.infos.set(filePath, {
        ...entry,
        runnable: false,
      });
      console.log("stopped", filePath);
    } else {
      alert(`failed to stop uploading ${filePath}: not found`);
    }
  };

  static delete = (filePath: string) => {
    UploadMgr.stop(filePath);
    UploadMgr.infos = UploadMgr.infos.delete(filePath);
  };

  static list = (): Map<string, UploadEntry> => {
    return UploadMgr.infos;
  };

  static respHandler = (event: MessageEvent) => {
    const resp = event.data as FileWorkerResp;

    switch (resp.kind) {
      case errKind:
        // TODO: refine this
        const errResp = resp as ErrResp;
        console.error(`respHandler: ${errResp}`);
        break;
      case uploadInfoKind:
        const infoResp = resp as UploadInfoResp;
        const entry = UploadMgr.infos.get(infoResp.filePath);

        if (entry != null) {
          if (infoResp.uploaded === entry.size) {
            UploadMgr.infos = UploadMgr.infos.delete(infoResp.filePath);
          } else {
            UploadMgr.infos = UploadMgr.infos.set(infoResp.filePath, {
              ...entry,
              uploaded: infoResp.uploaded,
              runnable: infoResp.runnable,
              err: infoResp.err,
            });
          }

          // call back to update the info
          UploadMgr.statusCb(UploadMgr.infos);
        } else {
          // TODO: refine this
          console.error(
            `respHandler: fail to found: file(${
              infoResp.filePath
            }) infos(${UploadMgr.infos.toObject()})`
          );
        }
        break;
      default:
        console.error(`respHandler: response kind not found: ${resp}`);
    }
  };
}
