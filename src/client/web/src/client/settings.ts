import { List } from "immutable";

import { BaseClient, Response, userIDParam, Quota } from ".";
import { ClientConfig, ClientErrorReport } from "./";

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
      data: {
        clientCfg: cfg,
      },
    });
  };

  reportErrors = (reports: List<ClientErrorReport>): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/settings/errors`,
      data: {
        reports: reports.toArray(),
      },
    });
  };
}
