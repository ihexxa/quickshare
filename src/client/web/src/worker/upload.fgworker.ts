import { UploadWorker } from "./upload.baseworker";
import { FileWorkerReq, FileWorkerResp } from "./interface";

export class FgWorker extends UploadWorker {
  constructor() {
    super();
  }

  // provide interfaces for non-worker mode
  onmessage = (event: MessageEvent): void => {};

  sendEvent = (resp: FileWorkerResp) => {
    this.onmessage(
      new MessageEvent("worker", {
        data: resp,
      })
    );
  };

  postMessage = (req: FileWorkerReq): void => {
    this.onMsg(
      new MessageEvent("worker", {
        data: req,
      })
    );
  };
}
