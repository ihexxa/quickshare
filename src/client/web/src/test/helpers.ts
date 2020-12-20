import { Response } from "../client";

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
