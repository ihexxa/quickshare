import { UploadWorker } from "./upload.baseworker";
import { FileWorkerResp } from "./interface";

const ctx: Worker = self as any;

class BgWorker extends UploadWorker {
  constructor() {
    super();
  }

  sendEvent = (resp: FileWorkerResp): void => {
    ctx.postMessage(resp);
  };
}

const worker = new BgWorker();
ctx.addEventListener("message", worker.onMsg);
ctx.addEventListener("error", worker.onError);

export default null as any;
