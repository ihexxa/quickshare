import { FileUploader } from "./uploader";

import { FileWorkerReq, FileWorkerResp } from "./interface";

function respond(
  filePath: string,
  uploaded: number,
  done: boolean,
  err: string | null
): void {
  postMessage({
    filePath: filePath,
    uploaded: uploaded,
    done: done,
    err: err,
  });
}

let file: File = undefined;
let filePath: string = undefined;
let uploader: FileUploader = undefined;

addEventListener("message", (req: MessageEvent) => {
  if (req.data.stop) {
    // it is a stop command
    uploader.stop();
    respond(filePath, uploader.getOffset(), true, uploader.err());
  } else {
    file = req.data.file;
    filePath = req.data.filePath;
    uploader = new FileUploader(file, filePath, respond);
    uploader.start().then((_: boolean) => {
      respond(filePath, uploader.getOffset(), true, uploader.err());
    });
  }
});

onerror = (ev: ErrorEvent) => {
  respond(
    filePath,
    -1,
    true,
    `${ev.error.toString()};\nuploader error:${uploader.err()}`
  );
};
