import { mock, instance, when } from "ts-mockito";

import { UploadWorker } from "../upload.baseworker";
import { FileUploader } from "../uploader";
import { FileWorkerResp, UploadEntry, syncReqKind, UploadState } from "../interface";

describe("upload.worker", () => {
  const content = ["123456"];
  const filePath = "mock/file";
  const blob = new Blob(content);
  const fileSize = blob.size;
  const file = new File(content, filePath);

  const makeEntry = (filePath: string, state: UploadState): UploadEntry => {
    return {
      file: file,
      filePath,
      size: fileSize,
      uploaded: 0,
      state,
      err: "",
    };
  };

  xtest("onMsg:syncReqKind: filter list and start uploading correct file", async () => {});
});
