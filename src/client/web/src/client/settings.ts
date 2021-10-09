import { BaseClient, Response, userIDParam, Quota } from ".";

import { ClientConfig } from "./";

export class SettingsClient extends BaseClient {
  constructor(url: string) {
    super(url);
  }

  health = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/settings/health`,
    });
  };

  getClientCfg = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/settings/client`,
    });
  };

  setClientCfg = (cfg: ClientConfig): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/settings/client`,
      data: cfg,
    });
  };
}
