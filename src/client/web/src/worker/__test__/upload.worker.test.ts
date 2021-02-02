import { mock, instance, when } from "ts-mockito";

import { UploadWorker } from "../upload.baseworker";
import { FileUploader } from "../uploader";
import { FileWorkerResp, UploadEntry, syncReqKind } from "../interface";

describe("upload.worker", () => {
  const content = ["123456"];
  const filePath = "mock/file";
  const blob = new Blob(content);
  const fileSize = blob.size;
  const file = new File(content, filePath);

  const makeEntry = (filePath: string, runnable: boolean): UploadEntry => {
    return {
      file: file,
      filePath,
      size: fileSize,
      uploaded: 0,
      runnable,
      err: "",
    };
  };

  test("onMsg:syncReqKind: filter list and start uploading correct file", async () => {
    const mockUploaderClass = mock(FileUploader);
    when(mockUploaderClass.start()).thenCall(
      (): Promise<boolean> => {
        return new Promise((resolve) => resolve(true));
      }
    );
    when(mockUploaderClass.stop()).thenCall(() => {});

    interface TestCase {
      desc: string;
      infos: Array<UploadEntry>;
      expectedUploadingFile: string;
      expectedUploaderStartInput: string;
      currentFilePath: string;
    }

    const tcs: Array<TestCase> = [
      {
        desc: "add new uploadings when worker is in idle",
        infos: [makeEntry("file1", true), makeEntry("file2", true)],
        currentFilePath: "",
        expectedUploadingFile: "file1",
        expectedUploaderStartInput: "file1",
      },
      {
        desc: "add new uploadings when worker is in idle and skip some stopped files",
        infos: [makeEntry("file1", false), makeEntry("file2", true)],
        currentFilePath: "",
        expectedUploadingFile: "file2",
        expectedUploaderStartInput: "file2",
      },
      {
        desc: "current file should be stopped and start new uploading",
        infos: [makeEntry("file0", false), makeEntry("file1", true)],
        currentFilePath: "file0",
        expectedUploadingFile: "file1",
        expectedUploaderStartInput: "file1",
      },
      {
        desc: "uploader should keep uploading if the first uploadable file is not changed",
        infos: [makeEntry("file1", true)],
        expectedUploadingFile: "file1",
        expectedUploaderStartInput: undefined,
        currentFilePath: "file1",
      },
    ];

    for (let i = 0; i < tcs.length; i++) {
      const uploadWorker = new UploadWorker();
      let currentUploader: FileUploader = undefined;
      let uploaderFile: File = undefined;
      let uploaderFilePath: string = undefined;
      let uploaderStopFilePath: string = undefined;
      uploadWorker.sendEvent = (_: FileWorkerResp) => {};
      uploadWorker.makeUploader = (
        file: File,
        filePath: string
      ): FileUploader => {
        uploaderFile = file;
        uploaderFilePath = filePath;
        currentUploader = instance(mockUploaderClass);
        return currentUploader;
      };
      if (tcs[i].currentFilePath !== "") {
        uploadWorker.setFilePath(tcs[i].currentFilePath);
      }
      const req = {
        kind: syncReqKind,
        infos: tcs[i].infos,
      };

      uploadWorker.onMsg(
        new MessageEvent("worker", {
          data: req,
        })
      );

      console.log(tcs[i].desc);
      expect(uploadWorker.getFilePath()).toEqual(tcs[i].expectedUploadingFile);
      expect(uploaderFilePath).toEqual(tcs[i].expectedUploaderStartInput);
    }
  });
});
