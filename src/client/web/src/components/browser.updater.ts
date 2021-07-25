import { List, Map } from "immutable";

import { ICoreState } from "./core_state";
import { Props, getItemPath } from "./browser";
import {
  IUsersClient,
  IFilesClient,
  MetadataResp,
  UploadInfo,
} from "../client";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { UploadEntry } from "../worker/interface";
import { Up } from "../worker/upload_mgr";

export class Updater {
  props: Props;
  private usersClient: IUsersClient = new UsersClient("");
  private filesClient: IFilesClient = new FilesClient("");

  init = (props: Props) => (this.props = { ...props });
  setClients(usersClient: IUsersClient, filesClient: IFilesClient) {
    this.usersClient = usersClient;
    this.filesClient = filesClient;
  }

  addUploads = (fileList: List<File>) => {
    fileList.forEach(file => {
      const filePath = getItemPath(
        this.props.dirPath.join("/"),
        file.name
      );
      // do not wait for the promise
      Up().add(file, filePath);
    })
    this.setUploadings(Up().list());
  };

  deleteUpload = async (filePath: string): Promise<boolean> => {
    Up().delete(filePath);
    const resp = await this.filesClient.deleteUploading(filePath);
    return resp.status === 200;
  };

  setUploadings = (infos: Map<string, UploadEntry>) => {
    this.props.uploadings = List<UploadInfo>(
      infos.valueSeq().map(
        (v: UploadEntry): UploadInfo => {
          return {
            realFilePath: v.filePath,
            size: v.size,
            uploaded: v.uploaded,
          };
        }
      )
    );
  };

  refreshUploadings = async (): Promise<boolean> => {
    const luResp = await this.filesClient.listUploadings();

    this.props.uploadings =
      luResp.status === 200
        ? List<UploadInfo>(luResp.data.uploadInfos)
        : this.props.uploadings;
    return luResp.status === 200;
  };

  stopUploading = (filePath: string) => {
    Up().stop(filePath);
  };

  mkDir = async (dirPath: string): Promise<void> => {
    const resp = await this.filesClient.mkdir(dirPath);
    if (resp.status !== 200) {
      alert(`failed to make dir ${dirPath}`);
    }
  };

  delete = async (
    dirParts: List<string>,
    items: List<MetadataResp>,
    selectedItems: Map<string, boolean>
  ): Promise<void> => {
    const delRequests = items
      .filter((item) => {
        return selectedItems.has(item.name);
      })
      .map(
        async (selectedItem: MetadataResp): Promise<string> => {
          const itemPath = getItemPath(dirParts.join("/"), selectedItem.name);
          const resp = await this.filesClient.delete(itemPath);
          return resp.status === 200 ? "" : selectedItem.name;
        }
      );

    const failedFiles = await Promise.all(delRequests);
    failedFiles.forEach((failedFile) => {
      if (failedFile !== "") {
        alert(`failed to delete ${failedFile}`);
      }
    });
    return this.setItems(dirParts);
  };

  setItems = async (dirParts: List<string>): Promise<void> => {
    const dirPath = dirParts.join("/");
    const listResp = await this.filesClient.list(dirPath);

    this.props.dirPath = dirParts;
    this.props.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.items;
  };

  setHomeItems = async (): Promise<void> => {
    const listResp = await this.filesClient.listHome();

    this.props.dirPath = List<string>(listResp.data.cwd.split("/"));
    this.props.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.items;
  };

  goHome = async (): Promise<void> => {
    const listResp = await this.filesClient.listHome();

    // how to get current dir? to dirPath?
    // this.props.dirPath = dirParts;
    this.props.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.items;
  };

  moveHere = async (
    srcDir: string,
    dstDir: string,
    selectedItems: Map<string, boolean>
  ): Promise<void> => {
    const moveRequests = List<string>(selectedItems.keys()).map(
      async (itemName: string): Promise<string> => {
        const oldPath = getItemPath(srcDir, itemName);
        const newPath = getItemPath(dstDir, itemName);
        const resp = await this.filesClient.move(oldPath, newPath);
        return resp.status === 200 ? "" : itemName;
      }
    );

    const failedFiles = await Promise.all(moveRequests);
    failedFiles.forEach((failedItem) => {
      if (failedItem !== "") {
        alert(`failed to move ${failedItem}`);
      }
    });

    return this.setItems(List<string>(dstDir.split("/")));
  };

  setBrowser = (prevState: ICoreState): ICoreState => {
    prevState.panel.browser = { ...prevState.panel, ...this.props };
    return prevState;
  };
}

export let browserUpdater = new Updater();
export const updater = (): Updater => {
  return browserUpdater;
};
export const setUpdater = (updater: Updater) => {
  browserUpdater = updater;
};
