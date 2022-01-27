import { Response, ISettingsClient, ClientConfigMsg } from "./";
import { makePromise } from "../test/helpers";

export interface SettingsClientResps {
  healthMockResp?: Response;
  setClientCfgMockResp?: Response;
  getClientCfgMockResp?: Response<ClientConfigMsg>;
  reportErrorsMockResp?: Response;
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
        url: "clientCfg_bg_url",
        repeat: "clientCfg_bg_repeat",
        position: "clientCfg_bg_position",
        align: "clientCfg_bg_align",
      },
    },
  },
  reportErrorsMockResp: {
    status: 200,
    statusText: "",
    data: {},
  },
};

export class JestSettingsClient {
  url: string = "";
  constructor(url: string) {
    this.url = url;
  }

  health = jest.fn().mockReturnValue(makePromise(resps.healthMockResp));
  getClientCfg = jest
    .fn()
    .mockReturnValue(makePromise(resps.getClientCfgMockResp));
  setClientCfg = jest
    .fn()
    .mockReturnValue(makePromise(resps.setClientCfgMockResp));
  reportErrors = jest
    .fn()
    .mockReturnValue(makePromise(resps.reportErrorsMockResp));
}

export const NewMockSettingsClient = (url: string): ISettingsClient => {
  return new JestSettingsClient(url);
};
