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

const ctx: Worker = self as any;

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
  const req = ev.data as FileWorkerReq;

  switch (req.kind) {
    case syncReqKind:
      const syncReq = req as SyncReq;
      
      // find the first qualified task
      const infoArray = syncReq.infos;
      for (let i = 0; i < infoArray.length; i++) {
        if (
          infoArray[i].runnable &&
          infoArray[i].uploaded < infoArray[i].size
        ) {
          if (infoArray[i].filePath !== filePath) {
            // TODO: wait for it is stopped
            if (uploader != null) {
              uploader.stop();
            }
            file = infoArray[i].file;
            filePath = infoArray[i].filePath;
            uploader = new FileUploader(file, filePath, respond);
            uploader.start();
          } else {
            // do nothing, keep running
          }
          break;
        }
      }
      break;
    default:
      console.log(
        `unknown worker request(${JSON.stringify(req)})`
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

export default null as any; 
