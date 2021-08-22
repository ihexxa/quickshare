import { List, Map, Set } from "immutable";

import { ICoreState } from "./core_state";
import { getItemPath } from "./browser";
import {
  User,
  ListUsersResp,
  ListRolesResp,
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
  props: ICoreState;
  private usersClient: IUsersClient = new UsersClient("");
  private filesClient: IFilesClient = new FilesClient("");

  init = (props: ICoreState) => (this.props = { ...props });
  setClients(usersClient: IUsersClient, filesClient: IFilesClient) {
    this.usersClient = usersClient;
    this.filesClient = filesClient;
  }

  initUploads = () => {
    this.props.panel.browser.uploadings.forEach((entry) => {
      Up().addStopped(entry.realFilePath, entry.uploaded, entry.size);
    });
    // this.setUploadings(Up().list());
  };

  addUploads = (fileList: List<File>) => {
    fileList.forEach((file) => {
      const filePath = getItemPath(
        this.props.panel.browser.dirPath.join("/"),
        file.name
      );
      // do not wait for the promise
      Up().add(file, filePath);
    });
    this.setUploadings(Up().list());
  };

  deleteUpload = async (filePath: string): Promise<boolean> => {
    Up().delete(filePath);
    const resp = await this.filesClient.deleteUploading(filePath);
    return resp.status === 200;
  };

  setUploadings = (infos: Map<string, UploadEntry>) => {
    this.props.panel.browser.uploadings = List<UploadInfo>(
      infos.valueSeq().map((v: UploadEntry): UploadInfo => {
        return {
          realFilePath: v.filePath,
          size: v.size,
          uploaded: v.uploaded,
        };
      })
    );
  };

  addSharing = async (): Promise<boolean> => {
    const dirPath = this.props.panel.browser.dirPath.join("/");
    const resp = await this.filesClient.addSharing(dirPath);
    return resp.status === 200;
  };

  deleteSharing = async (dirPath: string): Promise<boolean> => {
    const resp = await this.filesClient.deleteSharing(dirPath);
    return resp.status === 200;
  };

  isSharing = async (dirPath: string): Promise<boolean> => {
    const resp = await this.filesClient.isSharing(dirPath);
    this.props.panel.browser.isSharing = resp.status === 200;
    return resp.status === 200; // TODO: differentiate 404 and error
  };

  setSharing = (shared: boolean) => {
    this.props.panel.browser.isSharing = shared;
  };

  listSharings = async (): Promise<boolean> => {
    const resp = await this.filesClient.listSharings();
    this.props.panel.browser.sharings =
      resp.status === 200
        ? List<string>(resp.data.sharingDirs)
        : this.props.panel.browser.sharings;
    return resp.status === 200;
  };

  refreshUploadings = async (): Promise<boolean> => {
    const luResp = await this.filesClient.listUploadings();

    this.props.panel.browser.uploadings =
      luResp.status === 200
        ? List<UploadInfo>(luResp.data.uploadInfos)
        : this.props.panel.browser.uploadings;
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
      .map(async (selectedItem: MetadataResp): Promise<string> => {
        const itemPath = getItemPath(dirParts.join("/"), selectedItem.name);
        const resp = await this.filesClient.delete(itemPath);
        return resp.status === 200 ? "" : selectedItem.name;
      });

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

    this.props.panel.browser.dirPath = dirParts;
    this.props.panel.browser.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.panel.browser.items;
  };

  setHomeItems = async (): Promise<void> => {
    const listResp = await this.filesClient.listHome();

    this.props.panel.browser.dirPath = List<string>(
      listResp.data.cwd.split("/")
    );
    this.props.panel.browser.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.panel.browser.items;
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

  displayPane = (paneName: string) => {
    if (paneName === "") {
      // hide all panes
      this.props.panel.panes.displaying = "";
    } else {
      const pane = this.props.panel.panes.paneNames.get(paneName);
      if (pane != null) {
        this.props.panel.panes.displaying = paneName;
      } else {
        alert(`dialgos: pane (${paneName}) not found`);
      }
    }
  };

  self = async (): Promise<boolean> => {
    const resp = await this.usersClient.self();
    if (resp.status === 200) {
      this.props.panel.panes.userRole = resp.data.role;
      return true;
    }
    return false;
  };

  addUser = async (user: User): Promise<boolean> => {
    const resp = await this.usersClient.addUser(user.name, user.pwd, user.role);
    // TODO: should return uid instead
    return resp.status === 200;
  };

  delUser = async (userID: string): Promise<boolean> => {
    const resp = await this.usersClient.delUser(userID);
    return resp.status === 200;
  };

  setRole = async (userID: string, role: string): Promise<boolean> => {
    const resp = await this.usersClient.delUser(userID);
    return resp.status === 200;
  };

  forceSetPwd = async (userID: string, pwd: string): Promise<boolean> => {
    const resp = await this.usersClient.forceSetPwd(userID, pwd);
    return resp.status === 200;
  };

  listUsers = async (): Promise<boolean> => {
    const resp = await this.usersClient.listUsers();
    if (resp.status !== 200) {
      return false;
    }

    const lsRes = resp.data as ListUsersResp;
    let users = Map<User>({});
    lsRes.users.forEach((user: User) => {
      users = users.set(user.name, user);
    });
    this.props.panel.panes.admin.users = users;

    return true;
  };

  addRole = async (role: string): Promise<boolean> => {
    const resp = await this.usersClient.addRole(role);
    // TODO: should return uid instead
    return resp.status === 200;
  };

  delRole = async (role: string): Promise<boolean> => {
    const resp = await this.usersClient.delRole(role);
    return resp.status === 200;
  };

  listRoles = async (): Promise<boolean> => {
    const resp = await this.usersClient.listRoles();
    if (resp.status !== 200) {
      return false;
    }

    const lsRes = resp.data as ListRolesResp;
    let roles = Set<string>();
    Object.keys(lsRes.roles).forEach((role: string) => {
      roles = roles.add(role);
    });
    this.props.panel.panes.admin.roles = roles;

    return true;
  };

  login = async (
    user: string,
    pwd: string,
    captchaID: string,
    captchaInput: string
  ): Promise<boolean> => {
    const resp = await this.usersClient.login(
      user,
      pwd,
      captchaID,
      captchaInput
    );
    updater().setAuthed(resp.status === 200);
    return resp.status === 200;
  };

  logout = async (): Promise<boolean> => {
    const resp = await this.usersClient.logout();
    updater().setAuthed(false);
    return resp.status === 200;
  };

  isAuthed = async (): Promise<boolean> => {
    const resp = await this.usersClient.isAuthed();
    return resp.status === 200;
  };

  initIsAuthed = async (): Promise<void> => {
    return updater()
      .isAuthed()
      .then((isAuthed) => {
        updater().setAuthed(isAuthed);
      });
  };

  setAuthed = (isAuthed: boolean) => {
    this.props.panel.authPane.authed = isAuthed;
  };

  getCaptchaID = async (): Promise<boolean> => {
    return this.usersClient.getCaptchaID().then((resp) => {
      if (resp.status === 200) {
        this.props.panel.authPane.captchaID = resp.data.id;
      }
      return resp.status === 200;
    });
  };

  setPwd = async (oldPwd: string, newPwd: string): Promise<boolean> => {
    const resp = await this.usersClient.setPwd(oldPwd, newPwd);
    return resp.status === 200;
  };

  updateBrowser = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panel: {
        ...prevState.panel,
        browser: {
          dirPath: this.props.panel.browser.dirPath,
          isSharing: this.props.panel.browser.isSharing,
          items: this.props.panel.browser.items,
          uploadings: this.props.panel.browser.uploadings,
          sharings: this.props.panel.browser.sharings,
          uploadFiles: this.props.panel.browser.uploadFiles,
          uploadValue: this.props.panel.browser.uploadValue,
          isVertical: this.props.panel.browser.isVertical,
        },
      },
    };
  };

  updatePanes = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panel: {
        ...prevState.panel,
        panes: { ...prevState.panel.panes, ...this.props.panel.panes },
      },
    };
  };

  updateAuthPane = (preState: ICoreState): ICoreState => {
    preState.panel.authPane = {
      ...preState.panel.authPane,
      ...this.props.panel.authPane,
    };
    return preState;
  };
}

export let coreUpdater = new Updater();
export const updater = (): Updater => {
  return coreUpdater;
};
export const setUpdater = (updater: Updater) => {
  coreUpdater = updater;
};
