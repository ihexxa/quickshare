import { Response } from "../client";
import { ICoreState } from "../components/core_state";

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

export const mockUpdate = (apply: (prevState: ICoreState) => ICoreState): void => {};
