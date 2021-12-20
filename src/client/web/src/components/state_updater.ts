import { List, Map, Set } from "immutable";

import {
  ICoreState,
  sharingCtrl,
  ctrlOn,
  ctrlOff,
  ctrlHidden,
} from "./core_state";
import { getItemPath } from "../common/utils";
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
import { controlName as panelTabs } from "./root_frame";
import { settingsTabsCtrl } from "./dialog_settings";
import { settingsDialogCtrl } from "./layers";

import { MsgPackage, isValidLanPack } from "../i18n/msger";

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

  initUploads = (): string => {
    this.props.uploadingsInfo.uploadings.forEach((entry) => {
      Up().addStopped(entry.filePath, entry.uploaded, entry.size);
    });

    return "";
  };

  addUploads = (fileList: List<File>): string => {
    fileList.forEach((file) => {
      const filePath = getItemPath(
        this.props.filesInfo.dirPath.join("/"),
        file.name
      );
      const status = Up().add(file, filePath);
      if (status !== "") {
        return status;
      }
    });

    this.setUploads(Up().list());
    return "";
  };

  deleteUpload = async (filePath: string): Promise<string> => {
    const status = Up().delete(filePath);
    if (status !== "") {
      return status;
    }

    const resp = await this.filesClient.deleteUploading(filePath);
    return resp.status === 200 ? "" : "server.fail";
  };

  setUploads = (infos: Map<string, UploadEntry>) => {
    this.props.uploadingsInfo.uploadings = List<UploadEntry>(
      infos.valueSeq().map((entry: UploadEntry): UploadEntry => {
        return entry;
      })
    );
  };

  addSharing = async (): Promise<string> => {
    const dirPath = this.props.filesInfo.dirPath.join("/");
    const resp = await this.filesClient.addSharing(dirPath);
    return resp.status === 200 ? "" : "server.fail";
  };

  deleteSharing = async (dirPath: string): Promise<string> => {
    const resp = await this.filesClient.deleteSharing(dirPath);
    return resp.status === 200 ? "" : "server.fail";
  };

  syncIsSharing = async (dirPath: string): Promise<string> => {
    const resp = await this.filesClient.isSharing(dirPath);
    this.props.filesInfo.isSharing = resp.status === 200;
    if (resp.status !== 200 && resp.status !== 404) {
      return "server.fail";
    }
    return "";
  };

  setSharing = (shared: boolean) => {
    this.props.filesInfo.isSharing = shared;
  };

  listSharings = async (): Promise<string> => {
    const resp = await this.filesClient.listSharings();
    this.props.sharingsInfo.sharings =
      resp.status === 200
        ? List<string>(resp.data.sharingDirs)
        : this.props.sharingsInfo.sharings;
    return resp.status === 200 ? "" : "server.fail";
  };

  // this function gets information from server and merge them with local information
  // because some information (error) can only be detected from local
  refreshUploadings = async (): Promise<string> => {
    const luResp = await this.filesClient.listUploadings();
    if (luResp.status !== 200) {
      // TODO: log error
      console.error(luResp.data);
      return "server.fail";
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
    return "";
  };

  stopUploading = (filePath: string): string => {
    return Up().stop(filePath);
  };

  mkDir = async (dirPath: string): Promise<string> => {
    const resp = await this.filesClient.mkdir(dirPath);
    if (resp.status !== 200) {
      alertMsg(`failed to make dir ${dirPath}`);
      return "server.fail";
    }
    return "";
  };

  delete = async (
    dirParts: List<string>,
    items: List<MetadataResp>,
    selectedItems: Map<string, boolean>
  ): Promise<string> => {
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
      return "server.fail";
    }

    return this.setItems(dirParts);
  };

  setItems = async (dirParts: List<string>): Promise<string> => {
    const dirPath = dirParts.join("/");
    const listResp = await this.filesClient.list(dirPath);

    if (listResp.status === 200) {
      this.props.filesInfo.dirPath = dirParts;
      this.props.filesInfo.items = List<MetadataResp>(listResp.data.metadatas);
      return "";
    }
    this.props.filesInfo.dirPath = List<string>([]);
    this.props.filesInfo.items = List<MetadataResp>([]);
    return "server.fail";
  };

  setHomeItems = async (): Promise<string> => {
    const listResp = await this.filesClient.listHome();

    if (listResp.status === 200) {
      this.props.filesInfo.dirPath = List<string>(listResp.data.cwd.split("/"));
      this.props.filesInfo.items = List<MetadataResp>(listResp.data.metadatas);
      return "";
    }
    this.props.filesInfo.dirPath = List<string>([]);
    this.props.filesInfo.items = List<MetadataResp>([]);
    return "server.fail";
  };

  updateItems = (items: List<MetadataResp>) => {
    this.props.filesInfo.items = items;
  };

  moveHere = async (
    srcDir: string,
    dstDir: string,
    selectedItems: Map<string, boolean>
  ): Promise<string> => {
    const itemsToMove = List<string>(selectedItems.keys()).map(
      (itemName: string): any => {
        const from = getItemPath(srcDir, itemName);
        const to = getItemPath(dstDir, itemName);
        return { from, to };
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
      return "server.fail";
    }

    return this.setItems(List<string>(dstDir.split("/")));
  };

  initUITree = () => {
    const isAuthed = this.props.login.authed;
    const isSharing =
      this.props.ui.control.controls.get(sharingCtrl) === ctrlOn;

    if (isAuthed) {
      this.props.ui.control.controls = Map<string, string>({
        [panelTabs]: "filesPanel",
        [settingsDialogCtrl]: ctrlOff,
        [settingsTabsCtrl]: "preferencePane",
        [sharingCtrl]: isSharing ? ctrlOn : ctrlOff,
      });
      this.props.ui.control.options = Map<string, Set<string>>({
        [panelTabs]: Set<string>([
          "filesPanel",
          "uploadingsPanel",
          "sharingsPanel",
        ]),
        [settingsDialogCtrl]: Set<string>([ctrlOn, ctrlOff]),
        [settingsTabsCtrl]: Set<string>(["preferencePane"]),
        [sharingCtrl]: Set<string>([ctrlOn, ctrlOff]),
      });

      if (this.props.login.userRole == roleAdmin) {
        this.props.ui.control.options = this.props.ui.control.options.set(
          settingsTabsCtrl,
          Set<string>(["preferencePane", "managementPane"])
        );
      }
    } else {
      if (isSharing) {
        this.props.ui.control.controls = Map<string, string>({
          [panelTabs]: "filesPanel",
          [settingsDialogCtrl]: ctrlHidden,
          [settingsTabsCtrl]: ctrlHidden,
          [sharingCtrl]: ctrlOn,
        });
        this.props.ui.control.options = Map<string, Set<string>>({
          [panelTabs]: Set<string>(["filesPanel"]),
          [settingsDialogCtrl]: Set<string>([ctrlHidden]),
          [settingsTabsCtrl]: Set<string>([ctrlHidden]),
          [sharingCtrl]: Set<string>([ctrlOn]),
        });
      } else {
        this.props.ui.control.controls = Map<string, string>({
          [panelTabs]: ctrlHidden,
          [settingsDialogCtrl]: ctrlHidden,
          [settingsTabsCtrl]: ctrlHidden,
          [sharingCtrl]: ctrlOff,
        });
        this.props.ui.control.options = Map<string, Set<string>>({
          [panelTabs]: Set<string>([ctrlHidden]),
          [settingsDialogCtrl]: Set<string>([ctrlHidden]),
          [settingsTabsCtrl]: Set<string>([ctrlHidden]),
          [sharingCtrl]: Set<string>([ctrlOff]),
        });
      }
    }
  };

  initStateForVisitor = async (): Promise<any> => {
    const statuses = await Promise.all([
      this.getClientCfg(),
      this.syncIsSharing(this.props.filesInfo.dirPath.join("/")),
    ]);
    if (statuses.join("") !== "") {
      return statuses.join(";");
    }

    const syncLanStatus = await this.syncLan();
    if (syncLanStatus !== "" && syncLanStatus !== "server.fail.ignore") {
      return syncLanStatus;
    }
    return "";
  };

  initStateForAuthedUser = async (): Promise<string> => {
    const statuses = await Promise.all([
      this.refreshUploadings(),
      this.listSharings(),
      this.initUploads(),
    ]);
    if (statuses.join("") !== "") {
      return statuses.join(";");
    }
    return "";
  };

  initStateForAdmin = async (): Promise<string> => {
    const initVisitorStatus = await this.initStateForVisitor();
    if (initVisitorStatus !== "") {
      return initVisitorStatus;
    }
    const initAuthedUserStatus = await this.initStateForAuthedUser();
    if (initAuthedUserStatus !== "") {
      return initAuthedUserStatus;
    }
    const statuses = await Promise.all([this.listRoles(), this.listUsers()]);
    if (statuses.join("") !== "") {
      return statuses.join(";");
    }
    return "";
  };

  syncCwd = async (): Promise<string> => {
    if (this.props.filesInfo.dirPath.size !== 0) {
      return this.setItems(this.props.filesInfo.dirPath);
    }
    return this.setHomeItems();
  };

  initCwd = async (params: URLSearchParams): Promise<string> => {
    const dir = params.get("dir");

    if (dir != null && dir !== "") {
      // in sharing mode
      const dirPath = List(dir.split("/"));
      this.props.ui.control.controls = this.props.ui.control.controls.set(
        sharingCtrl,
        ctrlOn
      );
      this.props.filesInfo.dirPath = dirPath;
    } else {
      this.props.ui.control.controls = this.props.ui.control.controls.set(
        sharingCtrl,
        ctrlOff
      );
      this.props.filesInfo.dirPath = List([]);
    }

    return "";
  };

  initAll = async (params: URLSearchParams): Promise<string> => {
    const isAuthedStatus = await this.syncIsAuthed();
    if (isAuthedStatus !== "" && isAuthedStatus !== "server.fail.ignore") {
      return isAuthedStatus;
    }

    const statuses = await Promise.all([this.self(), this.initCwd(params)]);
    if (statuses.join("") !== "") {
      return statuses.join(";");
    }

    this.initUITree();
    const syncCwdStatus = await this.syncCwd();
    if (syncCwdStatus !== "") {
      return syncCwdStatus;
    }

    const getCapStatus = await this.getCaptchaID();
    if (getCapStatus !== "") {
      return getCapStatus;
    }

    if (this.props.login.userRole === roleAdmin) {
      return this.initStateForAdmin();
    } else if (this.props.login.userRole === roleVisitor) {
      return this.initStateForVisitor();
    }
    return this.initStateForAuthedUser();
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
    this.props.login.authed = false;
    this.props.login.captchaID = "";
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

  self = async (): Promise<string> => {
    const resp = await this.usersClient.self();
    if (resp.status === 200) {
      this.props.login.userID = resp.data.id;
      this.props.login.userName = resp.data.name;
      this.props.login.userRole = resp.data.role;
      this.props.login.usedSpace = resp.data.usedSpace;
      this.props.login.quota = resp.data.quota;
      this.props.login.preferences = resp.data.preferences;
      return "";
    }
    this.resetUser();
    return "server.fail";
  };

  addUser = async (user: User): Promise<string> => {
    const resp = await this.usersClient.addUser(user.name, user.pwd, user.role);
    // TODO: should return uid instead
    return resp.status === 200 ? "" : "server.fail";
  };

  delUser = async (userID: string): Promise<string> => {
    const resp = await this.usersClient.delUser(userID);
    return resp.status === 200 ? "" : "server.fail";
  };

  setUser = async (
    userID: string,
    role: string,
    quota: Quota
  ): Promise<string> => {
    const resp = await this.usersClient.setUser(userID, role, quota);
    return resp.status === 200 ? "" : "server.fail";
  };

  setRole = async (userID: string, role: string): Promise<string> => {
    const resp = await this.usersClient.delUser(userID);
    return resp.status === 200 ? "" : "server.fail";
  };

  forceSetPwd = async (userID: string, pwd: string): Promise<string> => {
    const resp = await this.usersClient.forceSetPwd(userID, pwd);
    return resp.status === 200 ? "" : "server.fail";
  };

  listUsers = async (): Promise<string> => {
    const resp = await this.usersClient.listUsers();
    if (resp.status !== 200) {
      return "server.fail";
    }

    const lsRes = resp.data as ListUsersResp;
    let users = Map<User>({});
    lsRes.users.forEach((user: User) => {
      users = users.set(user.name, user);
    });
    this.props.admin.users = users;

    return "";
  };

  addRole = async (role: string): Promise<string> => {
    const resp = await this.usersClient.addRole(role);
    // TODO: should return id instead
    return resp.status === 200 ? "" : "server.fail";
  };

  delRole = async (role: string): Promise<string> => {
    const resp = await this.usersClient.delRole(role);
    return resp.status === 200 ? "" : "server.fail";
  };

  listRoles = async (): Promise<string> => {
    const resp = await this.usersClient.listRoles();
    if (resp.status !== 200) {
      return "server.fail";
    }

    const lsRes = resp.data as ListRolesResp;
    let roles = Set<string>();
    Object.keys(lsRes.roles).forEach((role: string) => {
      roles = roles.add(role);
    });
    this.props.admin.roles = roles;

    return "";
  };

  login = async (
    user: string,
    pwd: string,
    captchaID: string,
    captchaInput: string
  ): Promise<string> => {
    const resp = await this.usersClient.login(
      user,
      pwd,
      captchaID,
      captchaInput
    );
    this.props.login.authed = resp.status === 200;
    return resp.status === 200 ? "" : "server.fail";
  };

  logout = async (): Promise<string> => {
    const resp = await this.usersClient.logout();
    this.resetUser();
    return resp.status === 200 ? "" : "server.fail";
  };

  syncIsAuthed = async (): Promise<string> => {
    const resp = await this.usersClient.isAuthed();
    if (resp.status !== 200) {
      this.props.login.authed = false;
      return resp.status === 401 ? "server.fail.ignore" : "server.fail";
    }
    this.props.login.authed = true;
    return "";
  };

  getCaptchaID = async (): Promise<string> => {
    const resp = await this.usersClient.getCaptchaID();
    if (resp.status !== 200) {
      return "server.fail";
    }
    this.props.login.captchaID = resp.data.id;
    return "";
  };

  setPwd = async (oldPwd: string, newPwd: string): Promise<string> => {
    const resp = await this.usersClient.setPwd(oldPwd, newPwd);
    return resp.status === 200 ? "" : "server.fail";
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

  setControlOption = (controlName: string, option: string): boolean => {
    const controlExists = this.props.ui.control.controls.has(controlName);
    const optionsExists = this.props.ui.control.options.has(controlName);
    const options = this.props.ui.control.options.get(controlName);
    if (!controlExists || !optionsExists || !options.has(option)) {
      console.error(
        `control(${controlName}-${controlExists}) or option(${option}-${optionsExists}) not found`
      );
      return false;
    }

    this.props.ui.control.controls = this.props.ui.control.controls.set(
      controlName,
      option
    );
    return true;
  };

  generateHash = async (filePath: string): Promise<string> => {
    const resp = await this.filesClient.generateHash(filePath);
    return resp.status === 200 ? "" : "server.fail";
  };

  setClientCfgRemote = async (cfg: ClientConfig): Promise<string> => {
    const resp = await this.settingsClient.setClientCfg(cfg);
    return resp.status === 200 ? "" : "server.fail";
  };

  setClientCfg = (cfg: ClientConfig) => {
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

  syncPreferences = async (): Promise<string> => {
    const resp = await this.usersClient.setPreferences(
      this.props.login.preferences
    );
    return resp.status === 200 ? "" : "server.fail";
  };

  getClientCfg = async (): Promise<string> => {
    const resp = await this.settingsClient.getClientCfg();
    if (resp.status !== 200) {
      return "server.fail";
    }
    const clientCfg = resp.data.clientCfg as ClientConfig;
    this.props.ui.siteName = clientCfg.siteName;
    this.props.ui.siteDesc = clientCfg.siteDesc;
    this.props.ui.bg = clientCfg.bg;
    return "";
  };

  syncLan = async (): Promise<string> => {
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
      return "fe.fail.ignore";
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
      return "server.fail.ignore";
    }
    this.props.msg.lan = resp.data.lan;
    this.props.msg.pkg = Map<string, string>(resp.data);
    return "";
  };

  updateAll = (prevState: ICoreState): ICoreState => {
    return {
      filesInfo: { ...prevState.filesInfo, ...this.props.filesInfo },
      uploadingsInfo: {
        ...prevState.uploadingsInfo,
        ...this.props.uploadingsInfo,
      },
      sharingsInfo: { ...prevState.sharingsInfo, ...this.props.sharingsInfo },
      login: { ...prevState.login, ...this.props.login },
      admin: { ...prevState.admin, ...this.props.admin },
      msg: { ...prevState.msg, ...this.props.msg },
      ui: { ...prevState.ui, ...this.props.ui },
    };
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
