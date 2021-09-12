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
  Quota,
  Response,
} from "../client";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { UploadEntry, UploadState } from "../worker/interface";
import { Up } from "../worker/upload_mgr";
import { alertMsg } from "../common/env";
import { LocalStorage } from "../common/localstorage";

import { MsgPackage } from "../i18n/msger";

function getCookieLanKey(user: string) {
  return `qs_${user}_lan`;
}

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
    this.props.browser.uploadings.forEach((entry) => {
      Up().addStopped(entry.filePath, entry.uploaded, entry.size);
    });
    // this.setUploadings(Up().list());
  };

  addUploads = (fileList: List<File>) => {
    fileList.forEach((file) => {
      const filePath = getItemPath(
        this.props.browser.dirPath.join("/"),
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
    this.props.browser.uploadings = List<UploadEntry>(
      infos.valueSeq().map((entry: UploadEntry): UploadEntry => {
        return entry;
      })
    );
  };

  addSharing = async (): Promise<boolean> => {
    const dirPath = this.props.browser.dirPath.join("/");
    const resp = await this.filesClient.addSharing(dirPath);
    return resp.status === 200;
  };

  deleteSharing = async (dirPath: string): Promise<boolean> => {
    const resp = await this.filesClient.deleteSharing(dirPath);
    return resp.status === 200;
  };

  isSharing = async (dirPath: string): Promise<boolean> => {
    const resp = await this.filesClient.isSharing(dirPath);
    this.props.browser.isSharing = resp.status === 200;
    return resp.status === 200; // TODO: differentiate 404 and error
  };

  setSharing = (shared: boolean) => {
    this.props.browser.isSharing = shared;
  };

  listSharings = async (): Promise<boolean> => {
    const resp = await this.filesClient.listSharings();
    this.props.browser.sharings =
      resp.status === 200
        ? List<string>(resp.data.sharingDirs)
        : this.props.browser.sharings;
    return resp.status === 200;
  };

  refreshUploadings = async (): Promise<boolean> => {
    // this function get information from server and merge them with local information
    // because some information (error) can only be detected from local
    const luResp = await this.filesClient.listUploadings();
    if (luResp.status !== 200) {
      // TODO: i18n
      console.error(luResp.data);
      return false;
    }

    let localUploads = Map<string, UploadEntry>([]);
    this.props.browser.uploadings.forEach((entry: UploadEntry) => {
      localUploads = localUploads.set(entry.filePath, entry);
    })


    let updatedUploads = List<UploadEntry>([]);
    luResp.data.uploadInfos.forEach((remoteInfo: UploadInfo) => {
      const localEntry = localUploads.get(remoteInfo.realFilePath);
      if (localEntry == null) {
        updatedUploads = updatedUploads.push({
          file: undefined,
          filePath: remoteInfo.realFilePath,
          size: remoteInfo.size,
          uploaded: remoteInfo.uploaded,
          state: UploadState.Ready,
          err: "",
        });
      } else {
        updatedUploads = updatedUploads.push({
          file: localEntry.file,
          filePath: localEntry.filePath,
          size: remoteInfo.size,
          uploaded: remoteInfo.uploaded,
          state: localEntry.state,
          err: localEntry.err,
        });
      }
    });

    this.props.browser.uploadings = updatedUploads;
    return true;
  };

  stopUploading = (filePath: string) => {
    Up().stop(filePath);
  };

  mkDir = async (dirPath: string): Promise<void> => {
    const resp = await this.filesClient.mkdir(dirPath);
    if (resp.status !== 200) {
      alertMsg(`failed to make dir ${dirPath}`);
    }
  };

  delete = async (
    dirParts: List<string>,
    items: List<MetadataResp>,
    selectedItems: Map<string, boolean>
  ): Promise<void> => {
    const pathsToDel = items
      .filter((item) => {
        return selectedItems.has(item.name);
      })
      .map((selectedItem: MetadataResp): string => {
        return getItemPath(dirParts.join("/"), selectedItem.name);
      });

    const batchSize = 3;
    let batch = List<string>();
    let fails = List<string>();

    for (let i = 0; i < pathsToDel.size; i++) {
      batch = batch.push(pathsToDel.get(i));

      if (batch.size >= batchSize || i == pathsToDel.size - 1) {
        let promises = batch.map(async (itemPath): Promise<Response<any>> => {
          return this.filesClient.delete(itemPath);
        });

        const resps = await Promise.all(promises.toSeq());
        resps.forEach((resp: Response<any>, i: number) => {
          if (resp.status !== 200) {
            fails = fails.push(batch.get(i));
          }
        });

        batch = batch.clear();
      }
    }

    if (fails.size > 0) {
      alertMsg(
        `${this.props.msg.pkg.get("delete.fail")}: ${fails.join(",\n")}`
      );
    }

    return this.setItems(dirParts);
  };

  setItems = async (dirParts: List<string>): Promise<void> => {
    const dirPath = dirParts.join("/");
    const listResp = await this.filesClient.list(dirPath);

    this.props.browser.dirPath = dirParts;
    this.props.browser.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.browser.items;
  };

  setHomeItems = async (): Promise<void> => {
    const listResp = await this.filesClient.listHome();

    this.props.browser.dirPath = List<string>(listResp.data.cwd.split("/"));
    this.props.browser.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : this.props.browser.items;
  };

  moveHere = async (
    srcDir: string,
    dstDir: string,
    selectedItems: Map<string, boolean>
  ): Promise<void> => {
    const itemsToMove = List<string>(selectedItems.keys()).map(
      (itemName: string): any => {
        const from = getItemPath(srcDir, itemName);
        const to = getItemPath(dstDir, itemName);
        return { from, to };
        // const resp = await this.filesClient.move(oldPath, newPath);
        // return resp.status === 200 ? "" : itemName;
      }
    );

    const batchSize = 3;
    let batch = List<any>();
    let fails = List<string>();

    for (let i = 0; i < itemsToMove.size; i++) {
      batch = batch.push(itemsToMove.get(i));

      if (batch.size >= batchSize || i == itemsToMove.size - 1) {
        let promises = batch.map(
          async (fromTo: any): Promise<Response<any>> => {
            return this.filesClient.move(fromTo.from, fromTo.to);
          }
        );

        const resps = await Promise.all(promises.toSeq());
        resps.forEach((resp: Response<any>, i: number) => {
          if (resp.status !== 200) {
            fails = fails.push(batch.get(i).from);
          }
        });

        batch = batch.clear();
      }
    }

    if (fails.size > 0) {
      alertMsg(`${this.props.msg.pkg.get("move.fail")}: ${fails.join(",\n")}`);
    }

    return this.setItems(List<string>(dstDir.split("/")));
  };

  displayPane = (paneName: string) => {
    if (paneName === "") {
      // hide all panes
      this.props.panes.displaying = "";
    } else {
      const pane = this.props.panes.paneNames.get(paneName);
      if (pane != null) {
        this.props.panes.displaying = paneName;
      } else {
        alertMsg(`dialgos: pane (${paneName}) not found`);
      }
    }
  };

  self = async (): Promise<boolean> => {
    const resp = await this.usersClient.self();
    if (resp.status === 200) {
      this.props.login.userID = resp.data.id;
      this.props.login.userName = resp.data.name;
      this.props.login.userRole = resp.data.role;
      this.props.login.usedSpace = resp.data.usedSpace;
      this.props.login.quota = resp.data.quota;
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

  setUser = async (
    userID: string,
    role: string,
    quota: Quota
  ): Promise<boolean> => {
    const resp = await this.usersClient.setUser(userID, role, quota);
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
    this.props.admin.users = users;

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
    this.props.admin.roles = roles;

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
    return this.isAuthed().then((isAuthed) => {
      updater().setAuthed(isAuthed);
    });
  };

  setAuthed = (isAuthed: boolean) => {
    this.props.login.authed = isAuthed;
  };

  getCaptchaID = async (): Promise<boolean> => {
    return this.usersClient.getCaptchaID().then((resp) => {
      if (resp.status === 200) {
        this.props.login.captchaID = resp.data.id;
      }
      return resp.status === 200;
    });
  };

  setPwd = async (oldPwd: string, newPwd: string): Promise<boolean> => {
    const resp = await this.usersClient.setPwd(oldPwd, newPwd);
    return resp.status === 200;
  };

  initLan = () => {
    const lanKey = getCookieLanKey(this.props.login.userName);
    const lanSaved = LocalStorage.get(lanKey);
    this.setLan(lanSaved === "" ? "en_US" : lanSaved);
  };

  setLan = (lan: string) => {
    const lanKey = getCookieLanKey(this.props.login.userName);

    switch (lan) {
      case "en_US":
        this.props.msg.lan = "en_US";
        this.props.msg.pkg = MsgPackage.get(lan);
        LocalStorage.set(lanKey, "en_US");
        break;
      case "zh_CN":
        this.props.msg.lan = "zh_CN";
        this.props.msg.pkg = MsgPackage.get(lan);
        LocalStorage.set(lanKey, "zh_CN");
        break;
      default:
        alertMsg("language package not found");
    }
  };

  setTab = (tabName: string) => {
    switch (tabName) {
      case "item":
        this.props.browser.tab = tabName;
        break;
      case "uploading":
        this.props.browser.tab = tabName;
        break;
      case "sharing":
        this.props.browser.tab = tabName;
        break;
      default:
        this.props.browser.tab = "item";
        break;
    }
  };

  updateBrowser = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      browser: { ...prevState.browser, ...this.props.browser },
    };
  };

  updatePanes = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panes: { ...prevState.panes, ...this.props.panes },
    };
  };

  updateLogin = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      login: { ...prevState.login, ...this.props.login },
    };
  };

  updateAdmin = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      admin: { ...prevState.admin, ...this.props.admin },
    };
  };

  updateMsg = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      msg: { ...prevState.msg, ...this.props.msg },
    };
  };
}

export let coreUpdater = new Updater();
export const updater = (): Updater => {
  return coreUpdater;
};
export const setUpdater = (updater: Updater) => {
  coreUpdater = updater;
};
