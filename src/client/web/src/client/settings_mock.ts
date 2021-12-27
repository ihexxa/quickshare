// TODO: replace this with jest mocks
import { Response } from ".";

import { ClientConfigMsg } from "./";

export interface SettingsClientResps {
  healthMockResp?: Response;
  setClientCfgMockResp?: Response;
  getClientCfgMockResp?: Response<ClientConfigMsg>;
  reportErrorResp?: Response;
}

export const resps = {
  healthMockResp: { status: 200, statusText: "", data: {} },
  setClientCfgMockResp: { status: 200, statusText: "", data: {} },
  getClientCfgMockResp: {
    status: 200,
    statusText: "",
    data: {
      clientCfg: {
        siteName: "",
        siteDesc: "",
        bg: {
          url: "clientCfg_bg_url",
          repeat: "clientCfg_bg_repeat",
          position: "clientCfg_bg_position",
          align: "clientCfg_bg_align",
        },
      },
    },
  },
  reportErrorResp: {
    status: 200,
    statusText: "",
    data: {},
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

  reportError = (): Promise<Response> => {
    return this.wrapPromise(this.resps.reportErrorResp);
  };
}
