import { List } from "immutable";

import { BaseClient, Response, userIDParam, Quota } from ".";
import { ClientConfigMsg, ClientErrorReport } from "./";

export class SettingsClient extends BaseClient {
  constructor(url: string) {
    super(url);
  }

  health = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v2/public/settings/health`,
    });
  };

  getClientCfg = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v2/public/settings/client`,
    });
  };

  setClientCfg = (cfg: ClientConfigMsg): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v2/admin/client`,
      data: cfg,
    });
  };

  reportErrors = (reports: List<ClientErrorReport>): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v2/my/errors`,
      data: {
        reports: reports.toArray(),
      },
    });
  };
}
