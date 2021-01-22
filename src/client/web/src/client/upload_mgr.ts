import { Map } from "immutable";

// import { UploadEntry } from "./";

import { FileWorkerReq, FileWorkerResp } from "../worker/interface";

// const worker = new UploadWorker();
// worker.postMessage({
//   file: file,
//   filePath: filePath,
//   stop: false,
// });

export interface UploadEntry {
  file: File;
  filePath: string;
  size: number;
  uploaded: number;
  done: boolean;
  working: boolean;
}

async function delay(ms: number): Promise<void> {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, ms);
  });
}
export class UploadMgr {
  private static infos = Map<string, UploadEntry>();
  private static worker: Worker = new Worker("static/worker.bundle.js");
  private static working: string = "";

  static init = () => {
    // TODO: fallback to normal if Web Worker is not available
    UploadMgr.worker.onmessage = UploadMgr.respHandler;
    const polling = () => {
      if (UploadMgr.working === "") {
        const filePaths = Object.keys(UploadMgr.infos);
        for (let i = 0; i < filePaths.length; i++) {
          const entry = UploadMgr.infos.get(filePaths[i]);
          if (entry.working) {
            UploadMgr.startWorker(entry.file, entry.filePath, entry.file.size);
            break;
          }
        }
      }
    };
    setInterval(polling, 500);
  };

  static stopWorker = (filePath: string) => {
    const entry = UploadMgr.infos.get(filePath);
    if (entry != null) {
      UploadMgr.working = "";
      UploadMgr.worker.postMessage({ stop: true });
      UploadMgr.infos.set(filePath, { ...entry, working: false });
    } else {
      alert(`failed to stop uploading ${filePath}`);
    }
  };

  static startWorker = (file: File, filePath: string, size: number) => {
    UploadMgr.working = "filePath";

    const entry = UploadMgr.infos.get(filePath);
    if (entry == null) {
      UploadMgr.infos.set(filePath, {
        file: file,
        filePath: filePath,
        size: size,
        uploaded: 0,
        done: false,
        working: true,
      });
    } else {
      UploadMgr.infos.set(filePath, { ...entry, working: true });
    }

    UploadMgr.worker.postMessage({
      file: file,
      filePath: filePath,
      stop: false,
    });
  };

  static start = (file: File, filePath: string) => {
    if (UploadMgr.working !== "") {
      // one file is in uploading
      if (UploadMgr.working !== filePath) {
        UploadMgr.stopWorker(UploadMgr.working);
        UploadMgr.startWorker(file, filePath, file.size);
      }
    } else {
      // no file is in uploading
      UploadMgr.startWorker(file, filePath, file.size);
    }
  };

  static delete = (filePath: string) => {
    UploadMgr.stopWorker(filePath);
  };

  static list = (): Map<string, UploadEntry> => {
    return UploadMgr.infos;
  };

  static respHandler = (event: MessageEvent) => {
    const resp = event.data as FileWorkerResp;

    const info = UploadMgr.infos.get(resp.filePath);
    if (info != null) {
      if (info.size === resp.uploaded) {
        UploadMgr.infos = UploadMgr.infos.delete(resp.filePath);
        UploadMgr.working = "";
      } else if (resp.done) {
        UploadMgr.infos = UploadMgr.infos.set(resp.filePath, {
          ...info,
          done: resp.done,
          uploaded: resp.uploaded,
        });
        UploadMgr.working = "";
      } else {
        UploadMgr.infos = UploadMgr.infos.set(resp.filePath, {
          ...info,
          uploaded: resp.uploaded,
        });
      }
    } else {
      alert(`file ${resp.filePath} is not found in the manager`);
    }
  };
}
