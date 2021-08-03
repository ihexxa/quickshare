import { Map } from "immutable";
import { mock, instance, when, anything } from "ts-mockito";

import { FilesClient } from "../../client/files_mock";
import { makePromise } from "../../test/helpers";

import { Up, initUploadMgr } from "../upload_mgr";
import { 
  UploadState,
  FileWorkerResp,
  UploadInfoResp
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
  const content = ["123456"];
  const filePath = "mock/file";
  const blob = new Blob(content);
  // const fileSize = blob.size;
  // const file = new File(content, filePath);


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
            err: ""
          },
          {
            file: undefined,
            filePath: "t1/file2",
            state: UploadState.Uploading,
            uploaded: 6, 
            size: 3,
            err: ""
          },
          {
            file: undefined,
            filePath: "t1/file3",
            state: UploadState.Ready,
            uploaded: 6, 
            size: 3,
            err: ""
          }
        ],
        expectedInfo: {
          file: undefined,
          filePath: "t1/file3",
          state: UploadState.Ready,
          uploaded: 6, 
          size: 3,
          err: ""
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
            err: ""
          },
          {
            file: undefined,
            filePath: "path2/file1",
            state: UploadState.Stopped,
            uploaded: 6, 
            size: 3,
            err: ""
          },
        ],
        expectedInfo: {
          file: undefined,
          filePath: "path3/file1",
          state: UploadState.Ready,
          uploaded: 6, 
          size: 3,
          err: ""
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
            err: ""
          },
          {
            file: undefined,
            filePath: "path2/file1",
            state: UploadState.Stopped,
            uploaded: 6, 
            size: 3,
            err: ""
          },
        ],
        expectedInfo: {
          file: undefined,
          filePath: "path3/file1",
          state: UploadState.Created,
          uploaded: 6, 
          size: 3,
          err: ""
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
});
