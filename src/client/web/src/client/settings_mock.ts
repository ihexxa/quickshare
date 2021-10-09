// TODO: replace this with jest mocks
import { Response, Quota } from ".";

import { ClientConfig } from "./";

export interface SettingsClientResps {
  healthMockResp?: Response;
  setClientCfgMockResp?: Response;
  getClientCfgMockResp?: Response<ClientConfig>;
}

export const resps = {
  healthMockResp: { status: 200, statusText: "", data: {} },
  setClientCfgMockResp: { status: 200, statusText: "", data: {} },
  getClientCfgMockResp: {
    status: 200,
    statusText: "",
    data: {
      siteName: "",
      siteDesc: "",
      bg: {
        url: "",
        repeat: "",
        position: "",
        align: "",
      },
    },
  },
};
export class MockSettingsClient {
  private url: string;
  private resps: SettingsClientResps;

  constructor(url: string) {
    this.url = url;
    this.resps = resps;
  }

  setMock = (resps: SettingsClientResps) => {
    this.resps = resps;
  };

  wrapPromise = (resp: any): Promise<any> => {
    return new Promise<any>((resolve) => {
      resolve(resp);
    });
  };

  health = (): Promise<Response> => {
    return this.wrapPromise(this.resps.healthMockResp);
  };

  setClientCfg = (): Promise<Response> => {
    return this.wrapPromise(this.resps.setClientCfgMockResp);
  };

  getClientCfg = (): Promise<Response> => {
    return this.wrapPromise(this.resps.getClientCfgMockResp);
  };
}
