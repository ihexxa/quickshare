import { Map } from "immutable";
import { sha1 } from "object-hash";

import { ILocalStorage, Storage } from "./localstorage";
import { ISettingsClient } from "../client";
import { SettingsClient } from "../client/settings";
import { ICoreState } from "../components/core_state";
import { updater } from "../components/state_updater";

const errorVer = "0.0.1";
const cookieKeyClErrs = "qs_cli_errs";

export interface ClientErrorV001 {
  version: string;
  error: string;
  state: ICoreState;
}

export interface IErrorLogger {
  setClient: (client: ISettingsClient) => void;
  setStorage: (storage: ILocalStorage) => void;
  error: (msg: string) => null | Error;
  report: () => Promise<null | Error>;
  readErrs: () => Map<string, ClientErrorV001>;
}

export class ErrorLog {
  private client: ISettingsClient;
  private storage: ILocalStorage = Storage();

  constructor(client: ISettingsClient) {
    this.client = client;
  }

  setClient(client: ISettingsClient) {
    this.client = client;
  }

  setStorage(storage: ILocalStorage) {
    this.storage = storage;
  }

  private getErrorSign = (errMsg: string): string => {
    return `e:${sha1(errMsg)}`;
  };

  readErrs = (): Map<string, ClientErrorV001> => {
    const errsStr = this.storage.get(cookieKeyClErrs);
    const errsObj = JSON.parse(errsStr);
    return Map(errsObj);
  };

  private writeErrs = (errs: Map<string, ClientErrorV001>) => {
    const errsObj = errs.toObject();
    const errsStr = JSON.stringify(errsObj);
    this.storage.set(cookieKeyClErrs, errsStr);
  };

  error = (msg: string): null | Error => {
    try {
      const sign = this.getErrorSign(msg);
      const clientErr: ClientErrorV001 = {
        version: errorVer,
        error: msg,
        state: updater().props,
      };
      let errs = this.readErrs();
      if (!errs.has(sign)) {
        errs = errs.set(sign, clientErr);
        this.writeErrs(errs);
      }
    } catch (err: any) {
      return Error(`failed to save err log: ${err}`);
    }

    return null;
  };

  report = async (): Promise<null | Error> => {
    try {
      const errs = this.readErrs();

      for (let sign of errs.keySeq().toArray()) {
        const err = errs.get(sign);
        const resp = await this.client.reportError(sign, JSON.stringify(err));
        if (resp.status !== 200) {
          return Error(`failed to report error: ${resp.data}`);
        }
      }

      // truncate errors
      this.writeErrs(Map());
    } catch (e: any) {
      return Error(e);
    }

    return null;
  };
}

const errorLogger = new ErrorLog(new SettingsClient(""));
export const ErrorLogger = (): IErrorLogger => {
  return errorLogger;
};
