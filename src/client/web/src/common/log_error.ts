import { ILocalStorage, Storage } from "./localstorage";
import { ISettingsClient } from "../client";
import { SettingsClient } from "../client/settings";

export interface IErrorLogger {
  setClient: (client: ISettingsClient) => void;
  setStorage: (storage: ILocalStorage) => void;
  error: (key: string, msg: string) => void;
  report: () => void;
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

  error = (key: string, msg: string) => {
    const existKey = this.storage.get(key);
    if (existKey === "") {
      this.storage.set(key, msg);
    }
  };

  report = () => {
    // TODO:
    // check last submitting, and set submit time
    // report all errors to backend
    // clean storage
  };
}

const errorLogger = new ErrorLog(new SettingsClient(""));
export const ErrorLogger = (): IErrorLogger => {
  return errorLogger;
};
