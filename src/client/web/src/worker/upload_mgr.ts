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
  UploadState,
} from "./interface";

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

    const syncing = () => {
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
    this.intervalID = win.setInterval(syncing, this.cycle);
  }

  destory = () => {
    win.clearInterval(this.intervalID);
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

  setStatusCb = (
    cb: (infos: Map<string, UploadEntry>, refresh: boolean) => void
  ) => {
    this.statusCb = cb;
  };

  // addStopped is for initializing uploading list in the UploadMgr
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

  add = (file: File, filePath: string) => {
    const entry = this.infos.get(filePath);
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
      // restart the uploading
      if (
        entry.state === UploadState.Stopped &&
        filePath === entry.filePath &&
        file.size === entry.size
      ) {
        // try to upload a file with same name but actually with different content.
        // it still can not resolve one case: names and sizes are same, but contents are different
        // TODO: showing file SHA will avoid above case
        this.infos = this.infos.set(filePath, {
          ...entry,
          file: file,
          state: UploadState.Ready,
        });
      } else {
        alert(
          `(${filePath}) seems not same file with uploading item, please check.`
        );
      }
    }
    this.statusCb(this.infos.toMap(), false);
  };

  stop = (filePath: string) => {
    const entry = this.infos.get(filePath);
    if (entry != null) {
      this.infos = this.infos.set(filePath, {
        ...entry,
        state: UploadState.Stopped,
      });
    } else {
      alert(`failed to stop uploading ${filePath}: not found`);
    }
    this.statusCb(this.infos.toMap(), false);
  };

  delete = (filePath: string) => {
    this.stop(filePath);
    this.infos = this.infos.delete(filePath);
    this.statusCb(this.infos.toMap(), false);
  };

  list = (): OrderedMap<string, UploadEntry> => {
    return this.infos;
  };

  respHandler = (event: MessageEvent) => {
    const resp = event.data as FileWorkerResp;

    switch (resp.kind) {
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
          // TODO: refine this
          console.error(`uploading ${errResp.filePath} may already be deleted`);
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
          // TODO: refine this
          console.error(
            `respHandler: may already be deleted: file(${
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
