import { List, Map, Set } from "immutable";

import { ICoreState } from "./core_state";
import { getItemPath } from "./browser";
import {
  User,
  ListUsersResp,
  ListRolesResp,
  IUsersClient,
  IFilesClient,
  ISettingsClient,
  MetadataResp,
  UploadInfo,
  Quota,
  Response,
  roleVisitor,
  roleAdmin,
  visitorID,
  ClientConfig,
  Preferences,
} from "../client";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { SettingsClient } from "../client/settings";
import { UploadEntry, UploadState } from "../worker/interface";
import { Up } from "../worker/upload_mgr";
import { alertMsg } from "../common/env";
import { LocalStorage } from "../common/localstorage";

import { MsgPackage, isValidLanPack } from "../i18n/msger";

function getCookieLanKey(user: string) {
  return `qs_${user}_lan`;
}

export class Updater {
  props: ICoreState;
  private usersClient: IUsersClient = new UsersClient("");
  private filesClient: IFilesClient = new FilesClient("");
  private settingsClient: ISettingsClient = new SettingsClient("");

  init = (props: ICoreState) => (this.props = { ...props });
  setClients(
    usersClient: IUsersClient,
    filesClient: IFilesClient,
    settingsClient: ISettingsClient
  ) {
    this.usersClient = usersClient;
    this.filesClient = filesClient;
    this.settingsClient = settingsClient;
  }

  initUploads = () => {
    this.props.uploadingsInfo.uploadings.forEach((entry) => {
      Up().addStopped(entry.filePath, entry.uploaded, entry.size);
    });
    // this.setUploadings(Up().list());
  };

  addUploads = (fileList: List<File>) => {
    fileList.forEach((file) => {
      const filePath = getItemPath(
        this.props.filesInfo.dirPath.join("/"),
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
    this.props.uploadingsInfo.uploadings = List<UploadEntry>(
      infos.valueSeq().map((entry: UploadEntry): UploadEntry => {
        return entry;
      })
    );
  };

  addSharing = async (): Promise<boolean> => {
    const dirPath = this.props.filesInfo.dirPath.join("/");
    const resp = await this.filesClient.addSharing(dirPath);
    return resp.status === 200;
  };

  deleteSharing = async (dirPath: string): Promise<boolean> => {
    const resp = await this.filesClient.deleteSharing(dirPath);
    return resp.status === 200;
  };

  isSharing = async (dirPath: string): Promise<boolean> => {
    const resp = await this.filesClient.isSharing(dirPath);
    this.props.filesInfo.isSharing = resp.status === 200;
    return resp.status === 200; // TODO: differentiate 404 and error
  };

  setSharing = (shared: boolean) => {
    this.props.filesInfo.isSharing = shared;
  };

  listSharings = async (): Promise<boolean> => {
    const resp = await this.filesClient.listSharings();
    this.props.sharingsInfo.sharings =
      resp.status === 200
        ? List<string>(resp.data.sharingDirs)
        : this.props.sharingsInfo.sharings;
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
    this.props.uploadingsInfo.uploadings.forEach((entry: UploadEntry) => {
      localUploads = localUploads.set(entry.filePath, entry);
    });

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

    this.props.uploadingsInfo.uploadings = updatedUploads;
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
  ): Promise<boolean> => {
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

  setItems = async (dirParts: List<string>): Promise<boolean> => {
    const dirPath = dirParts.join("/");
    const listResp = await this.filesClient.list(dirPath);

    if (listResp.status === 200) {
      this.props.filesInfo.dirPath = dirParts;
      this.props.filesInfo.items = List<MetadataResp>(listResp.data.metadatas);
      return true;
    }
    this.props.filesInfo.dirPath = List<string>([]);
    this.props.filesInfo.items = List<MetadataResp>([]);
    return false;
  };

  setHomeItems = async (): Promise<boolean> => {
    const listResp = await this.filesClient.listHome();

    if (listResp.status === 200) {
      this.props.filesInfo.dirPath = List<string>(listResp.data.cwd.split("/"));
      this.props.filesInfo.items = List<MetadataResp>(listResp.data.metadatas);
      return true;
    }
    this.props.filesInfo.dirPath = List<string>([]);
    this.props.filesInfo.items = List<MetadataResp>([]);
    return false;
  };

  moveHere = async (
    srcDir: string,
    dstDir: string,
    selectedItems: Map<string, boolean>
  ): Promise<boolean> => {
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
        alertMsg(`pane (${paneName}) not found`);
      }
    }
  };

  setPanes = (paneNames: Set<string>) => {
    this.props.panes.paneNames = paneNames;
  };

  initPanes = async (): Promise<Array<any>> => {
    // init browser content
    if (this.props.login.userRole === roleVisitor) {
      if (this.props.filesInfo.isSharing) {
        // sharing with visitor
        this.setPanes(Set<string>(["login"]));
        this.displayPane("");
        return Promise.all([]);
      }

      // redirect to login
      this.setPanes(Set<string>(["login"]));
      this.displayPane("login");
      return Promise.all([this.getCaptchaID()]);
    }

    if (this.props.login.userRole === roleAdmin) {
      this.setPanes(Set<string>(["login", "settings", "admin"]));
    } else {
      this.setPanes(Set<string>(["login", "settings"]));
    }
    this.displayPane("");

    return Promise.all([
      this.refreshUploadings(),
      this.initUploads(),
      this.listSharings(),
    ]);
  };

  initAll = async (params: URLSearchParams): Promise<any> => {
    return this.initIsAuthed()
      .then(() => {
        return this.self();
      })
      .then(() => {
        const dir = params.get("dir");
        if (dir != null && dir !== "") {
          const dirPath = List(dir.split("/"));
          return this.setItems(dirPath);
        } else {
          return this.setHomeItems();
        }
      })
      .then(() => {
        return this.isSharing(this.props.filesInfo.dirPath.join("/"));
      })
      .then(() => {
        // init settings
        return this.getClientCfg();
      })
      .then(() => {
        // init panes
        return this.initPanes();
      })
      .then(() => {
        // init i18n
        // TOOD: status is ignored, should return alert
        return this.fetchLanPack();
      })
      .then(() => {
        // init admin content
        if (this.props.login.userRole === roleAdmin) {
          return Promise.all([this.listRoles(), this.listUsers()]);
        }
        return;
      });
  };

  resetUser = () => {
    this.props.login.userID = visitorID;
    this.props.login.userName = "visitor";
    this.props.login.userRole = roleVisitor;
    this.props.login.usedSpace = "0";
    this.props.login.quota = {
      uploadSpeedLimit: 0,
      downloadSpeedLimit: 0,
      spaceLimit: "0",
    };
    this.props.login.preferences = {
      bg: {
        url: "",
        repeat: "",
        position: "",
        align: "",
      },
      cssURL: "",
      lanPackURL: "",
      lan: "en_US",
    };
  };

  self = async (): Promise<boolean> => {
    const resp = await this.usersClient.self();
    if (resp.status === 200) {
      this.props.login.userID = resp.data.id;
      this.props.login.userName = resp.data.name;
      this.props.login.userRole = resp.data.role;
      this.props.login.usedSpace = resp.data.usedSpace;
      this.props.login.quota = resp.data.quota;
      this.props.login.preferences = resp.data.preferences;
      return true;
    }
    this.resetUser();
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
    this.resetUser();
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

  setLan = (lan: string) => {
    switch (lan) {
      case "en_US":
        this.props.msg.lan = "en_US";
        this.props.msg.pkg = MsgPackage.get(lan);
        this.props.login.preferences.lan = "en_US";
        break;
      case "zh_CN":
        this.props.msg.lan = "zh_CN";
        this.props.msg.pkg = MsgPackage.get(lan);
        this.props.login.preferences.lan = "zh_CN";
        break;
      default:
        alertMsg("language package not found");
    }
  };

  setTab = (tabName: string) => {
    // switch (tabName) {
    //   case "item":
    //     this.props.panels.displaying = tabName;
    //     break;
    //   case "uploading":
    //     this.props.panels.displaying = tabName;
    //     break;
    //   case "sharing":
    //     this.props.panels.displaying = tabName;
    //     break;
    //   default:
    //     this.props.panels.displaying = "item";
    //     break;
    // }
  };

  setControlOption = (controlName: string, option: string): boolean => {
    const controlExists = this.props.ui.control.controls.has(controlName);
    const options = this.props.ui.control.options.get(controlName);
    if (!controlExists || !options.has(option)) {
      return false;
    }

    this.props.ui.control.controls = this.props.ui.control.controls.set(
      controlName,
      option
    );
    return true;
  };

  generateHash = async (filePath: string): Promise<boolean> => {
    const resp = await this.filesClient.generateHash(filePath);
    return resp.status === 200;
  };

  setClientCfgRemote = async (cfg: ClientConfig): Promise<number> => {
    const resp = await this.settingsClient.setClientCfg(cfg);
    return resp.status;
  };

  setClientCfg = async (cfg: ClientConfig): Promise<void> => {
    this.props.ui = {
      ...this.props.ui,
      siteName: cfg.siteName,
      siteDesc: cfg.siteDesc,
      bg: cfg.bg,
    };
  };

  setPreferences = (prefer: Preferences) => {
    this.props.login.preferences = { ...prefer };
  };

  syncPreferences = async (): Promise<number> => {
    const resp = await this.usersClient.setPreferences(
      this.props.login.preferences
    );
    return resp.status;
  };

  getClientCfg = async (): Promise<number> => {
    const resp = await this.settingsClient.getClientCfg();
    if (resp.status === 200) {
      const clientCfg = resp.data.clientCfg as ClientConfig;
      this.props.ui.siteName = clientCfg.siteName;
      this.props.ui.siteDesc = clientCfg.siteDesc;
      this.props.ui.bg = clientCfg.bg;
    }

    return resp.status;
  };

  fetchLanPack = async (): Promise<number> => {
    const url = this.props.login.preferences.lanPackURL;
    if (url === "") {
      const lan = this.props.login.preferences.lan;
      if (lan == "en_US" || lan == "zh_CN") {
        // fallback to build-in language pack
        this.props.msg.lan = lan;
        this.props.msg.pkg = MsgPackage.get(lan);
      } else {
        // fallback to english
        this.props.msg.lan = "en_US";
        this.props.msg.pkg = MsgPackage.get("en_US");
      }
      return 404;
    }

    const resp = await this.filesClient.download(url);
    let isValid = true;
    if (resp == null || resp.data == null) {
      isValid = false;
    } else if (!isValidLanPack(resp.data)) {
      isValid = false;
    }

    if (!isValid) {
      this.props.msg.lan = "en_US";
      this.props.msg.pkg = MsgPackage.get("en_US");
      return 400;
    }
    this.props.msg.lan = resp.data.lan;
    this.props.msg.pkg = Map<string, string>(resp.data);
    return resp.status;
  };

  updateFilesInfo = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      filesInfo: { ...prevState.filesInfo, ...this.props.filesInfo },
    };
  };

  updateUploadingsInfo = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      uploadingsInfo: {
        ...prevState.uploadingsInfo,
        ...this.props.uploadingsInfo,
      },
    };
  };

  updateSharingsInfo = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      sharingsInfo: { ...prevState.sharingsInfo, ...this.props.sharingsInfo },
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

  updateUI = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      ui: { ...prevState.ui, ...this.props.ui },
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
