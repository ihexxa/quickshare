import { mock, instance } from "ts-mockito";
import { List } from "immutable";

import { MockWorker } from "../worker/interface";
import { initUploadMgr } from "../worker/upload_mgr";
import { Response } from "../client";
import { ICoreState, initState } from "../components/core_state";

export const makePromise = (ret: any): Promise<any> => {
  return new Promise<any>((resolve) => {
    resolve(ret);
  });
};

export const makeNumberResponse = (status: number): Promise<Response> => {
  return makePromise({
    status: status,
    statusText: "",
    data: {},
  });
};

export const mockUpdate = (
  apply: (prevState: ICoreState) => ICoreState
): void => {
  apply(initState());
};

export const addMockUpdate = (subState: any) => {
  subState.update = mockUpdate;
};

export function mockRandFile(filePath: string): File {
  const values = new Array<string>(Math.floor(7 * Math.random()));
  const content = [values.join("")];
  return new File(content, filePath);
}

export function mockFileList(filePaths: Array<string>): List<File> {
  const files = filePaths.map(filePath => {
    return mockRandFile(filePath);
  })
  return List<File>(files);
}

export function initMockWorker() {
  const mockWorkerClass = mock(MockWorker);
  const mockWorker = instance(mockWorkerClass);
  initUploadMgr(mockWorker);
}