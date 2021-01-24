import { FileUploader } from "./uploader";
import {
  FileWorkerReq,
  syncReqKind,
  SyncReq,
  errKind,
  ErrResp,
  uploadInfoKind,
  UploadInfoResp,
} from "./interface";

// export default null;

const ctx: Worker = self as any;
// const ctx = self;
console.log(ctx);

export function respond(
  filePath: string,
  uploaded: number,
  runnable: boolean,
  err: string
) {
  const uploadInfoResp: UploadInfoResp = {
    kind: uploadInfoKind,
    filePath,
    uploaded,
    runnable,
    err,
  };
  ctx.postMessage(uploadInfoResp);
}

export function respondErr(err: string) {
  const errResp: ErrResp = {
    kind: errKind,
    err,
  };
  ctx.postMessage(errResp);
}

let file: File = undefined;
let filePath: string = undefined;
let uploader: FileUploader = undefined;

function msgHandler(ev: MessageEvent) {
  console.log("receive", ev);
  const req = ev.data as FileWorkerReq;

  switch (req.kind) {
    case syncReqKind:
      const syncReq = req as SyncReq;
      // find the first qualified task
      const infoArray = syncReq.infos.valueSeq().toArray().reverse();
      for (let i = 0; i < infoArray.length; i++) {
        if (
          infoArray[i].runnable &&
          infoArray[i].uploaded < infoArray[i].size
        ) {
          if (infoArray[i].filePath !== filePath) {
            // TODO: wait for it is stopped
            uploader.stop();
            file = infoArray[i].file;
            filePath = infoArray[i].filePath;
            uploader = new FileUploader(file, filePath, respond);
            uploader.start().then((_: boolean) => {
              respond(filePath, uploader.getOffset(), false, uploader.err());
            });
          } else {
            // do nothing, keep running
          }
          break;
        }
      }
      break;
    default:
      console.log(
        `unknown worker request ${req.kind} ${req.kind === "worker.req.sync"}`
      );
  }
}

ctx.addEventListener("message", msgHandler);
ctx.onmessage = msgHandler;

ctx.addEventListener("error", (ev: ErrorEvent) => {
  if (ev.error != null) {
    respondErr(`worker error: ${ev.error.toString()}`);
  }
});
