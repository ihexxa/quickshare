import { ICoreState } from "./core_state";
import { updater, Updater } from "./state_updater";
import { UploadEntry } from "../worker/interface";
import { MetadataResp } from "../client";

export class QuickshareAPI {
  private updater: Updater;
  constructor() {
    this.updater = updater();
  }
  initAll = async (params: URLSearchParams): Promise<string> => {
    return this.updater.initAll(params);
  };

  addUploadArray = (fileArray: Array<File>): string => {
    return this.updater.addUploadArray(fileArray);
  };

  deleteUpload = async (filePath: string): Promise<string> => {
    return await this.updater.deleteUpload(filePath);
  };

  listUploadArray = async (): Promise<Array<UploadEntry>> => {
    return await this.updater.listUploadArray();
  };

  stopUploading = (filePath: string): string => {
    return this.updater.stopUploading(filePath);
  };

  self = async (): Promise<string> => {
    return await this.updater.self();
  };

  getProps = (): ICoreState => {
    return this.updater.props;
  };

  deleteInArray = async(itemsToDel: Array<string>): Promise<string> => {
    return await this.updater.deleteInArray(itemsToDel);
  }
}

const api = new QuickshareAPI();
export const API = (): QuickshareAPI => {
  return api;
};
