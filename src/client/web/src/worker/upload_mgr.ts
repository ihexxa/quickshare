import { Map, OrderedMap } from "immutable";

import {
  FileWorkerReq,
  FileWorkerResp,
  UploadInfoResp,
  ErrResp,
  UploadEntry,
  syncReqKind,
  errKind,
  uploadInfoKind,
  imIdleKind,
  UploadState,
} from "./interface";
import { errUploadMgr } from "../common/errors";
import { ErrorLogger } from "../common/log_error";

const win: Window = self as any;

export interface IWorker {
  onmessage: (event: MessageEvent) => void;
  postMessage: (event: FileWorkerReq) => void;
}

export class UploadMgr {
  private idx = 0;
  private cycle: number = 500;
  private intervalID: number;
  private worker: IWorker;
  private infos = OrderedMap<string, UploadEntry>();
  private statusCb = (
    infos: Map<string, UploadEntry>,
    refresh: boolean
  ): void => {};

  constructor(worker: IWorker) {
    this.worker = worker;
    this.worker.onmessage = this.respHandler;
  }

  syncing = () => {
    if (this.infos.size === 0) {
      return;
    }
    if (this.idx > 10000) {
      this.idx = 0;
    }

    const start = this.idx % this.infos.size;
    const infos = this.infos.valueSeq().toArray();
    for (let i = 0; i < this.infos.size; i++) {
      const pos = (start + i) % this.infos.size;
      const info = infos[pos];

      if (
        info.state === UploadState.Ready ||
        info.state === UploadState.Created
      ) {
        this.infos = this.infos.set(info.filePath, {
          ...info,
          state: UploadState.Uploading,
        });

        this.worker.postMessage({
          kind: syncReqKind,
          file: info.file,
          filePath: info.filePath,
          size: info.size,
          uploaded: info.uploaded,
          created: info.uploaded > 0 || info.state === UploadState.Created,
        });
        break;
      }
    }

    this.idx++;
  };

  _setInfos = (infos: OrderedMap<string, UploadEntry>) => {
    this.infos = infos;
  };

  setCycle = (ms: number) => {
    this.cycle = ms;
  };

  getCycle = (): number => {
    return this.cycle;
  };

  // TODO: change it to observer pattern
  // so that it can be observed by multiple components
  setStatusCb = (
    cb: (infos: Map<string, UploadEntry>, refresh: boolean) => void
  ) => {
    this.statusCb = cb;
  };

  // addStopped is for initializing uploading list in the UploadMgr, when page is loaded
  // notice even uploading list are shown in the UI, it may not inited in the UploadMgr
  addStopped = (filePath: string, uploaded: number, fileSize: number) => {
    this.infos = this.infos.set(filePath, {
      file: new File([""], filePath), // create a dumb file
      filePath,
      size: fileSize,
      uploaded,
      state: UploadState.Stopped,
      err: "",
    });
  };

  add = (file: File, filePath: string): string => {
    const entry = this.infos.get(filePath);
    let status = "";

    if (entry == null) {
      // new uploading
      this.infos = this.infos.set(filePath, {
        file: file,
        filePath: filePath,
        size: file.size,
        uploaded: 0,
        state: UploadState.Ready,
        err: "",
      });
    } else {
      // the uploading task exists, restart the uploading
      if (
        entry.state === UploadState.Stopped &&
        filePath === entry.filePath &&
        file.size === entry.size
      ) {
        // try to upload a file with same name
        // the file may be a totally different file.
        // TODO: checking file SHA will avoid above case
        this.infos = this.infos.set(filePath, {
          ...entry,
          file: file,
          state: UploadState.Ready,
        });
      } else {
        status = errUploadMgr;
      }
    }

    this.statusCb(this.infos.toMap(), false);
    return status;
  };

  stop = (filePath: string): string => {
    const entry = this.infos.get(filePath);
    let status = "";

    if (entry != null) {
      this.infos = this.infos.set(filePath, {
        ...entry,
        state: UploadState.Stopped,
      });
    } else {
      status = errUploadMgr;
    }

    this.statusCb(this.infos.toMap(), false);
    return status;
  };

  delete = (filePath: string): string => {
    const status = this.stop(filePath);
    if (status !== "") {
      return status;
    }

    this.infos = this.infos.delete(filePath);
    this.statusCb(this.infos.toMap(), false);
    return "";
  };

  list = (): OrderedMap<string, UploadEntry> => {
    return this.infos;
  };

  respHandler = (event: MessageEvent) => {
    const resp = event.data as FileWorkerResp;

    switch (resp.kind) {
      case imIdleKind:
        this.syncing();
        break;
      case errKind:
        const errResp = resp as ErrResp;
        const errEntry = this.infos.get(errResp.filePath);

        if (errEntry != null) {
          this.infos = this.infos.set(errResp.filePath, {
            ...errEntry,
            state: UploadState.Error,
            err: errResp.err,
          });
        } else {
          ErrorLogger().error(`respHandler: entry not found ${errResp.err}`);
        }

        this.statusCb(this.infos.toMap(), false);
        break;
      case uploadInfoKind:
        const infoResp = resp as UploadInfoResp;
        const entry = this.infos.get(infoResp.filePath);

        if (entry != null) {
          if (infoResp.uploaded === entry.size) {
            this.infos = this.infos.delete(infoResp.filePath);
            this.statusCb(this.infos.toMap(), true);
          } else {
            this.infos = this.infos.set(infoResp.filePath, {
              ...entry,
              uploaded: infoResp.uploaded,
              state:
                // this avoids overwriting Stopped/Error state
                entry.state === UploadState.Stopped ||
                entry.state === UploadState.Error
                  ? UploadState.Stopped
                  : infoResp.state,
            });
            this.statusCb(this.infos.toMap(), false);
          }
        } else {
          ErrorLogger().error(
            `respHandler: entry(uploadInfoKind) not found ${infoResp.err}`
          );
        }

        break;
      default:
        ErrorLogger().error(
          `respHandler: unknown kind: ${JSON.stringify(resp)}`
        );
    }
  };
}

export let uploadMgr: UploadMgr = undefined;
export const initUploadMgr = (worker: IWorker): UploadMgr => {
  uploadMgr = new UploadMgr(worker);
  return uploadMgr;
};
export const Up = (): UploadMgr => {
  return uploadMgr;
};
export const setUploadMgr = (up: UploadMgr) => {
  uploadMgr = up;
};
