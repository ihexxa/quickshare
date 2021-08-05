import { Map } from "immutable";

import { FgWorker } from "../upload.fg.worker";

import { Up, initUploadMgr } from "../upload_mgr";
import {
  UploadState,
  UploadStatus,
} from "../interface";

import {
  FileWorkerReq,
  UploadEntry,
  uploadInfoKind,
  syncReqKind,
  SyncReq,
} from "../interface";

function arraytoMap(infos: Array<UploadEntry>): Map<string, UploadEntry> {
  let map = Map<string, UploadEntry>();
  infos.forEach((info) => {
    map = map.set(info.filePath, info);
  });
  return map;
}

const delay = (ms: number): Promise<void> => {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, ms);
  });
};

describe("UploadMgr", () => {
  const newFile = (filePath: string, content: string): File => {
    const contentArray = [content];
    const blob = new Blob(contentArray);
    return new File(contentArray, filePath);
  };

  test("test syncing: pick up item which is ready", async () => {
    interface TestCase {
      inputInfos: Array<UploadEntry>;
      expectedInfo: UploadEntry;
    }

    class MockWorker {
      constructor() {}
      public expectedEntry: UploadEntry;
      onmessage = (event: MessageEvent): void => {};
      postMessage = (req: FileWorkerReq): void => {
        expect(req.filePath).toEqual(this.expectedEntry.filePath);
        expect(req.uploaded).toEqual(this.expectedEntry.uploaded);
        expect(req.size).toEqual(this.expectedEntry.size);
      };
    }

    const tcs: Array<TestCase> = [
      {
        inputInfos: [
          {
            file: undefined,
            filePath: "t1/file1",
            state: UploadState.Stopped,
            uploaded: 3,
            size: 2,
            err: "",
          },
          {
            file: undefined,
            filePath: "t1/file2",
            state: UploadState.Uploading,
            uploaded: 6,
            size: 3,
            err: "",
          },
          {
            file: undefined,
            filePath: "t1/file3",
            state: UploadState.Ready,
            uploaded: 6,
            size: 3,
            err: "",
          },
        ],
        expectedInfo: {
          file: undefined,
          filePath: "t1/file3",
          state: UploadState.Ready,
          uploaded: 6,
          size: 3,
          err: "",
        },
      },
      {
        inputInfos: [
          {
            file: undefined,
            filePath: "path3/file1",
            state: UploadState.Ready,
            uploaded: 6,
            size: 3,
            err: "",
          },
          {
            file: undefined,
            filePath: "path2/file1",
            state: UploadState.Stopped,
            uploaded: 6,
            size: 3,
            err: "",
          },
        ],
        expectedInfo: {
          file: undefined,
          filePath: "path3/file1",
          state: UploadState.Ready,
          uploaded: 6,
          size: 3,
          err: "",
        },
      },
      {
        inputInfos: [
          {
            file: undefined,
            filePath: "path3/file1",
            state: UploadState.Created,
            uploaded: 6,
            size: 3,
            err: "",
          },
          {
            file: undefined,
            filePath: "path2/file1",
            state: UploadState.Stopped,
            uploaded: 6,
            size: 3,
            err: "",
          },
        ],
        expectedInfo: {
          file: undefined,
          filePath: "path3/file1",
          state: UploadState.Created,
          uploaded: 6,
          size: 3,
          err: "",
        },
      },
    ];

    const worker = new MockWorker();
    for (let i = 0; i < tcs.length; i++) {
      initUploadMgr(worker);
      const up = Up();
      up.setCycle(50);

      const infoMap = arraytoMap(tcs[i].inputInfos);
      up._setInfos(infoMap);
      worker.expectedEntry = tcs[i].expectedInfo;

      // TODO: find a better way to wait
      // polling needs several rounds to finish all the tasks
      await delay(tcs.length * up.getCycle() + 1000);
      up.destory();
    }
  });

  test("test e2e: from syncing to respHandler", async () => {
    interface TestCase {
      inputInfos: Array<UploadEntry>;
      expectedInfos: Array<UploadEntry>;
    }

    class MockUploader {
      constructor() {}
      create = (filePath: string, file: File): Promise<UploadStatus> => {
        return new Promise((resolve) =>
          resolve({
            filePath,
            uploaded: 0,
            state: UploadState.Created,
            err: "",
          })
        );
      };
      upload = (
        filePath: string,
        file: File,
        uploaded: number
      ): Promise<UploadStatus> => {
        return new Promise((resolve) =>
          resolve({
            filePath,
            uploaded: file.size,
            state: UploadState.Ready,
            err: "",
          })
        );
      };
    }

    const tcs: Array<TestCase> = [
      {
        inputInfos: [
          {
            file: newFile("t1/file1", "123"),
            filePath: "t1/file1",
            state: UploadState.Ready,
            uploaded: 0,
            size: 3,
            err: "",
          },
          {
            file: newFile("t1/file2", "123"),
            filePath: "t1/file2",
            state: UploadState.Ready,
            uploaded: 0,
            size: 3,
            err: "",
          },
        ],
        expectedInfos: [],
      },
      {
        inputInfos: [
          {
            file: newFile("t1/file1", "123"),
            filePath: "t1/file1",
            state: UploadState.Stopped,
            uploaded: 0,
            size: 3,
            err: "",
          },
          {
            file: newFile("t1/file2", "123"),
            filePath: "t1/file2",
            state: UploadState.Error,
            uploaded: 0,
            size: 3,
            err: "",
          },
          {
            file: newFile("t1/file3", "123"),
            filePath: "t1/file3",
            state: UploadState.Ready,
            uploaded: 0,
            size: 3,
            err: "",
          },
        ],
        expectedInfos: [
          {
            file: newFile("t1/file1", "123"),
            filePath: "t1/file1",
            state: UploadState.Stopped,
            uploaded: 0,
            size: 3,
            err: "",
          },
          {
            file: newFile("t1/file2", "123"),
            filePath: "t1/file2",
            state: UploadState.Error,
            uploaded: 0,
            size: 3,
            err: "",
          },
        ],
      },
    ];

    for (let i = 0; i < tcs.length; i++) {
      const uploader = new MockUploader();
      const worker = new FgWorker();
      worker.setUploader(uploader);
      initUploadMgr(worker);
      const up = Up();
      up.setCycle(1);

      const infoMap = arraytoMap(tcs[i].inputInfos);
      up._setInfos(infoMap);

      // TODO: find a better way to wait, or this test is flanky
      // polling needs several rounds to finish all the tasks
      await delay(tcs[i].inputInfos.length * up.getCycle() * 2 + 5000);

      const infos = up.list();
      expect(infos.size).toEqual(tcs[i].expectedInfos.length);
      if (tcs[i].expectedInfos.length !== 0) {
        tcs[i].expectedInfos.forEach((info) => {
          const expectedInfo = infos.get(info.filePath);
          expect(expectedInfo).toEqual(info);
        });
      }

      up.destory();
    }
  });
});
