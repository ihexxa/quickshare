import { Response } from "../client";
import { ICoreState, mockState } from "../components/core_state";
import { List } from "immutable";

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
  apply(mockState());
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
