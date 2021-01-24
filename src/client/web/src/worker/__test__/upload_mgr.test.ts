import { Map } from "immutable";
import { mock, instance, when, anything } from "ts-mockito";

import { FilesClient } from "../../client/files_mock";
import { makePromise } from "../../test/helpers";

import { UploadMgr } from "../upload_mgr";

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
  const fileSize = blob.size;
  const file = new File(content, filePath);
  const makeInfo = (filePath: string, runnable: boolean): UploadEntry => {
    return {
      file: file,
      filePath: filePath,
      size: fileSize,
      uploaded: 0,
      runnable,
      err: "",
    };
  };

  test("test init and respHandler: pick up tasks and remove them after done", async () => {
    interface TestCase {
      inputInfos: Array<UploadEntry>;
      expectedInfos: Array<UploadEntry>;
    }

    class MockWorker {
      constructor() {}
      onmessage = (event: MessageEvent): void => {};
      postMessage = (req: FileWorkerReq): void => {
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
                this.onmessage(
                  new MessageEvent("worker", {
                    data: {
                      kind: uploadInfoKind,
                      filePath: infoArray[i].filePath,
                      uploaded: infoArray[i].size,
                      runnable: true,
                      err: "",
                    },
                  })
                );
                break;
              }
            }
            break;
          default:
            throw Error(
              `unknown worker request ${req.kind} ${req.kind === syncReqKind}`
            );
        }
      };
    }

    const tcs: Array<TestCase> = [
      {
        inputInfos: [
          makeInfo("path1/file1", true),
          makeInfo("path2/file1", true),
        ],
        expectedInfos: [],
      },
      {
        inputInfos: [
          makeInfo("path1/file1", true),
          makeInfo("path2/file1", false),
        ],
        expectedInfos: [makeInfo("path2/file1", false)],
      },
      {
        inputInfos: [
          makeInfo("path1/file1", false),
          makeInfo("path2/file1", true),
        ],
        expectedInfos: [
          makeInfo("path1/file1", false),
        ],
      },
    ];

    const worker = new MockWorker();
    UploadMgr.setCycle(100);

    for (let i = 0; i < tcs.length; i++) {
      const infoMap = arraytoMap(tcs[i].inputInfos);
      UploadMgr._setInfos(infoMap);

      UploadMgr.init(worker);
      // polling needs several rounds to finish all the tasks
      await delay(tcs.length * UploadMgr.getCycle() + 1000);
      // TODO: find a better way to wait
      const gotInfos = UploadMgr.list();

      const expectedInfoMap = arraytoMap(tcs[i].expectedInfos);
      gotInfos.keySeq().forEach((filePath) => {
        expect(gotInfos.get(filePath)).toEqual(expectedInfoMap.get(filePath));
      });
      expectedInfoMap.keySeq().forEach((filePath) => {
        expect(expectedInfoMap.get(filePath)).toEqual(gotInfos.get(filePath));
      });

      UploadMgr.destory();
    }
  });
});
