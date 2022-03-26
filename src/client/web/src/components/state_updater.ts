import { List, Map, Set } from "immutable";

import { ICoreState } from "./core_state";
import { getItemPath, sortRows, Row } from "../common/utils";
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
  roleUser,
  roleAdmin,
  visitorID,
  ClientConfigMsg,
  Preferences,
} from "../client";
import { FilesClient, shareIDQuery, shareDirQuery } from "../client/files";
import { UsersClient } from "../client/users";
import { SettingsClient } from "../client/settings";
import { UploadEntry, UploadState } from "../worker/interface";
import { Up } from "../worker/upload_mgr";
import { Env } from "../common/env";
import { controlName as panelTabs } from "./root_frame";
import { errServer } from "../common/errors";
import { ErrorLogger } from "../common/log_error";
import {
  settingsTabsCtrl,
  settingsDialogCtrl,
  sharingCtrl,
  ctrlOn,
  ctrlOff,
  ctrlHidden,
} from "../common/controls";

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

    this.setUploads(Up().list());

    const resp = await this.filesClient.deleteUploading(filePath);
    return resp.status === 200 ? "" : errServer;
  };

  setUploads = (infos: Map<string, UploadEntry>) => {
    this.props.uploadingsInfo.uploadings = List<UploadEntry>(
      infos.valueSeq().map((entry: UploadEntry): UploadEntry => {
        return entry;
      })
    );

    this.sortUploadings(
      this.props.uploadingsInfo.orderBy,
      this.props.uploadingsInfo.order
    );
  };

  addSharing = async (): Promise<string> => {
    const dirPath = this.props.filesInfo.dirPath.join("/");
    const resp = await this.filesClient.addSharing(dirPath);
    return resp.status === 200 ? "" : errServer;
  };

  deleteSharing = async (dirPath: string): Promise<string> => {
    const resp = await this.filesClient.deleteSharing(dirPath);
    return resp.status === 200 ? "" : errServer;
  };

  syncIsSharing = async (dirPath: string): Promise<string> => {
    const resp = await this.filesClient.isSharing(dirPath);
    this.props.filesInfo.isSharing = resp.status === 200;
    if (resp.status !== 200 && resp.status !== 404) {
      return errServer;
    }
    return "";
  };

  setSharing = (shared: boolean) => {
    this.props.filesInfo.isSharing = shared;
  };

  listSharings = async (): Promise<string> => {
    const resp = await this.filesClient.listSharingIDs();
    if (resp.status !== 200) {
      return errServer;
    }
    this.props.sharingsInfo.sharings = Map<string, string>(resp.data.IDs);
    this.sortSharings(
      this.props.sharingsInfo.orderBy,
      this.props.sharingsInfo.order
    );
    return "";
  };

  // this function gets information from server and merge them with local information
  // because some information (error) can only be detected from local
  refreshUploadings = async (): Promise<string> => {
    const luResp = await this.filesClient.listUploadings();
    if (luResp.status !== 200) {
      // this method is called for authed users
      // other status codes are unexpected, including 401
      ErrorLogger().error(
        `refreshUploadings: unexpected response ${luResp.status} ${luResp.data}`
      );
      return errServer;
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
          state: UploadState.Stopped,
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
    this.sortUploadings(
      this.props.uploadingsInfo.orderBy,
      this.props.uploadingsInfo.order
    );
    return "";
  };

  stopUploading = (filePath: string): string => {
    return Up().stop(filePath);
  };

  mkDir = async (dirPath: string): Promise<string> => {
    const resp = await this.filesClient.mkdir(dirPath);
    if (resp.status !== 200) {
      Env().alertMsg(`failed to make dir ${dirPath}`);
      return errServer;
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
      Env().alertMsg(
        `${this.props.msg.pkg.get("delete.fail")}: ${fails.join(",\n")}`
      );
      return errServer;
    }

    return this.setItems(dirParts);
  };

  refreshFiles = async (): Promise<string> => {
    const status = await this.setItems(this.props.filesInfo.dirPath);
    if (status !== "") {
      return status;
    }
  };

  setItems = async (dirParts: List<string>): Promise<string> => {
    const dirPath = dirParts.join("/");
    const listResp = await this.filesClient.list(dirPath);

    if (listResp.status === 200) {
      this.props.filesInfo.dirPath = dirParts;
      this.props.filesInfo.items = List<MetadataResp>(listResp.data.metadatas);
      this.sortFiles(this.props.filesInfo.orderBy, this.props.filesInfo.order);
      return "";
    }
    this.props.filesInfo.dirPath = List<string>([]);
    this.props.filesInfo.items = List<MetadataResp>([]);
    return errServer;
  };

  setHomeItems = async (): Promise<string> => {
    const listResp = await this.filesClient.listHome();

    if (listResp.status === 200) {
      this.props.filesInfo.dirPath = List<string>(listResp.data.cwd.split("/"));
      this.props.filesInfo.items = List<MetadataResp>(listResp.data.metadatas);
      this.sortFiles(this.props.filesInfo.orderBy, this.props.filesInfo.order);
      return "";
    }
    this.props.filesInfo.dirPath = List<string>([]);
    this.props.filesInfo.items = List<MetadataResp>([]);
    return errServer;
  };

  updateItems = (items: List<MetadataResp>) => {
    this.props.filesInfo.items = items;
  };

  updateUploadings = (uploadings: List<UploadEntry>) => {
    this.props.uploadingsInfo.uploadings = uploadings;
  };

  updateSharings = (sharings: Map<string, string>) => {
    this.props.sharingsInfo.sharings = sharings;
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
      Env().alertMsg(`${this.props.msg.pkg.get("move.fail")}: ${fails.join(",\n")}`);
      return errServer;
    }

    return this.setItems(List<string>(dstDir.split("/")));
  };

  initUITree = () => {
    const isAuthed = this.props.login.authed;
    const isSharing =
      this.props.ui.control.controls.get(sharingCtrl) === ctrlOn;
    const leftControls = this.props.ui.control.controls.filter(
      (_: string, key: string): boolean => {
        return (
          key !== panelTabs &&
          key !== settingsDialogCtrl &&
          key !== settingsTabsCtrl &&
          key !== sharingCtrl
        );
      }
    );
    const leftOpts = this.props.ui.control.options.filter(
      (_: Set<string>, key: string): boolean => {
        return (
          key !== panelTabs &&
          key !== settingsDialogCtrl &&
          key !== settingsTabsCtrl &&
          key !== sharingCtrl
        );
      }
    );

    let newControls: Map<string, string> = undefined;
    let newOptions: Map<string, Set<string>> = undefined;
    if (isAuthed) {
      newControls = Map<string, string>({
        [panelTabs]: "filesPanel",
        [settingsDialogCtrl]: ctrlOff,
        [settingsTabsCtrl]: "preferencePane",
        [sharingCtrl]: isSharing ? ctrlOn : ctrlOff,
      });
      newOptions = Map<string, Set<string>>({
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
        newOptions = newOptions.set(
          settingsTabsCtrl,
          Set<string>(["preferencePane", "managementPane"])
        );
      }
    } else {
      if (isSharing) {
        newControls = Map<string, string>({
          [panelTabs]: "filesPanel",
          [settingsDialogCtrl]: ctrlHidden,
          [settingsTabsCtrl]: ctrlHidden,
          [sharingCtrl]: ctrlOn,
        });
        newOptions = Map<string, Set<string>>({
          [panelTabs]: Set<string>(["filesPanel"]),
          [settingsDialogCtrl]: Set<string>([ctrlHidden]),
          [settingsTabsCtrl]: Set<string>([ctrlHidden]),
          [sharingCtrl]: Set<string>([ctrlOn]),
        });
      } else {
        newControls = Map<string, string>({
          [panelTabs]: ctrlHidden,
          [settingsDialogCtrl]: ctrlHidden,
          [settingsTabsCtrl]: ctrlHidden,
          [sharingCtrl]: ctrlOff,
        });
        newOptions = Map<string, Set<string>>({
          [panelTabs]: Set<string>([ctrlHidden]),
          [settingsDialogCtrl]: Set<string>([ctrlHidden]),
          [settingsTabsCtrl]: Set<string>([ctrlHidden]),
          [sharingCtrl]: Set<string>([ctrlOff]),
        });
      }
    }

    this.props.ui.control.controls = newControls.merge(leftControls);
    this.props.ui.control.options = newOptions.merge(leftOpts);
  };

  initClientCfg = async (): Promise<string> => {
    const status = await this.getClientCfg();
    if (status !== "") {
      return status;
    }

    return await this.syncLan();
  };

  initCwd = async (shareDir: string): Promise<string> => {
    if (shareDir !== "") {
      // in sharing mode
      const dirPath = List(shareDir.split("/"));
      this.props.filesInfo.dirPath = dirPath;
    } else {
      this.props.filesInfo.dirPath = List([]);
    }

    return "";
  };

  initFiles = async (shareDir: string): Promise<string> => {
    let status = await this.initCwd(shareDir);
    if (status !== "") {
      return status;
    }

    status = await this.syncCwd();
    if (status !== "") {
      return status;
    }

    return await this.syncIsSharing(this.props.filesInfo.dirPath.join("/"));
  };

  initUploadings = async (): Promise<string> => {
    const status = await this.refreshUploadings();
    if (status !== "") {
      return status;
    }
    this.initUploads();
    return "";
  };

  initSharings = async (): Promise<string> => {
    return await this.listSharings();
  };

  initAdmin = async (): Promise<string> => {
    const statuses = await Promise.all([this.listRoles(), this.listUsers()]);
    if (statuses.join("") !== "") {
      return statuses.join(";");
    }
    return "";
  };

  syncCwd = async (): Promise<string> => {
    if (this.props.filesInfo.dirPath.size !== 0) {
      return this.setItems(this.props.filesInfo.dirPath);
    } else if (this.props.login.authed) {
      return this.setHomeItems();
    }
    // cwd will not be synced if the user is not authned and without sharing mode
    return "";
  };

  initParams = async (
    params: URLSearchParams
  ): Promise<Map<string, string>> => {
    let paramMap = Map<string, string>();
    const paramKeys = [shareIDQuery, shareDirQuery];
    paramKeys.forEach((key) => {
      const val = params.get(key);
      paramMap = paramMap.set(key, val != null ? val : "");
    });

    const shareID = paramMap.get(shareIDQuery);
    if (shareID !== "") {
      const resp = await this.filesClient.getSharingDir(shareID);
      if (resp.status === 200) {
        paramMap = paramMap.set(shareDirQuery, resp.data.sharingDir);
      } else {
        ErrorLogger().error(
          `initParams: unexpected response ${resp.status} ${resp.data}`
        );
      }
    }

    return paramMap;
  };

  initControls = async (params: Map<string, string>) => {
    const shareDir = params.get(shareDirQuery);
    this.props.ui.control.controls = this.props.ui.control.controls.set(
      sharingCtrl,
      shareDir !== "" ? ctrlOn : ctrlOff
    );
  };

  initAuth = async (): Promise<string> => {
    const isAuthedStatus = await this.syncIsAuthed();
    if (isAuthedStatus !== "") {
      return isAuthedStatus;
    }

    const selfStatuses = await this.self();
    if (selfStatuses !== "") {
      return selfStatuses;
    }

    return "";
  };

  initAll = async (params: URLSearchParams): Promise<string> => {
    const paramMap = await this.initParams(params);
    const getCapStatus = await this.getCaptchaID();
    if (getCapStatus !== "") {
      return getCapStatus;
    }

    const authStatus = await this.initAuth();
    if (authStatus !== "") {
      return authStatus;
    }

    const initClientCfgStatus = await this.initClientCfg();
    if (initClientCfgStatus !== "") {
      return initClientCfgStatus;
    }

    this.initControls(paramMap);
    this.initUITree();

    const isInSharingMode = paramMap.get(shareDirQuery) !== "";
    if (
      (this.props.login.userRole === roleVisitor && isInSharingMode) ||
      this.props.login.userRole === roleUser ||
      this.props.login.userRole === roleAdmin
    ) {
      const shareDir = paramMap.get(shareDirQuery);
      const initFilesStatus = await this.initFiles(shareDir);
      if (initFilesStatus !== "") {
        return initFilesStatus;
      }
    }

    if (
      this.props.login.userRole === roleUser ||
      this.props.login.userRole === roleAdmin
    ) {
      const statuses = await Promise.all([
        this.initUploadings(),
        this.initSharings(),
      ]);
      if (statuses.join("") !== "") {
        return statuses.join(";");
      }
    }

    if (this.props.login.userRole === roleAdmin) {
      const status = await this.initAdmin();
      if (status !== "") {
        return status;
      }
    }

    return "";
  };

  resetUser = () => {
    this.props.login.userID = visitorID;
    this.props.login.userName = "visitor";
    this.props.login.userRole = roleVisitor;
    this.props.login.extInfo.usedSpace = "0";
    this.props.login.quota = {
      uploadSpeedLimit: 0,
      downloadSpeedLimit: 0,
      spaceLimit: "0",
    };
    this.props.login.authed = false;
    this.props.login.preferences = {
      bg: {
        url: "",
        repeat: "",
        position: "",
        align: "",
        bgColor: "",
      },
      cssURL: "",
      lanPackURL: "",
      lan: "en_US",
      theme: "light",
      avatar: "",
      email: "",
    };
  };

  self = async (): Promise<string> => {
    const resp = await this.usersClient.self();

    if (resp.status === 200) {
      this.props.login.userID = resp.data.id;
      this.props.login.userName = resp.data.name;
      this.props.login.userRole = resp.data.role;
      this.props.login.extInfo.usedSpace = resp.data.usedSpace;
      this.props.login.quota = resp.data.quota;
      this.props.login.preferences = resp.data.preferences;
      return "";
    } else if (resp.status === 403) {
      this.resetUser();
      return "";
    }

    this.resetUser();
    return errServer;
  };

  addUser = async (user: User): Promise<string> => {
    const resp = await this.usersClient.addUser(user.name, user.pwd, user.role);
    // TODO: should return uid instead
    return resp.status === 200 ? "" : errServer;
  };

  delUser = async (userID: string): Promise<string> => {
    const resp = await this.usersClient.delUser(userID);
    return resp.status === 200 ? "" : errServer;
  };

  setUser = async (
    userID: string,
    role: string,
    quota: Quota
  ): Promise<string> => {
    const resp = await this.usersClient.setUser(userID, role, quota);
    return resp.status === 200 ? "" : errServer;
  };

  setRole = async (userID: string, role: string): Promise<string> => {
    const resp = await this.usersClient.delUser(userID);
    return resp.status === 200 ? "" : errServer;
  };

  forceSetPwd = async (userID: string, pwd: string): Promise<string> => {
    const resp = await this.usersClient.forceSetPwd(userID, pwd);
    return resp.status === 200 ? "" : errServer;
  };

  listUsers = async (): Promise<string> => {
    const resp = await this.usersClient.listUsers();
    if (resp.status !== 200) {
      return errServer;
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
    return resp.status === 200 ? "" : errServer;
  };

  delRole = async (role: string): Promise<string> => {
    const resp = await this.usersClient.delRole(role);
    return resp.status === 200 ? "" : errServer;
  };

  listRoles = async (): Promise<string> => {
    const resp = await this.usersClient.listRoles();
    if (resp.status !== 200) {
      return errServer;
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
    return resp.status === 200 ? "" : errServer;
  };

  logout = async (): Promise<string> => {
    const resp = await this.usersClient.logout();
    this.resetUser();
    return resp.status === 200 ? "" : errServer;
  };

  syncIsAuthed = async (): Promise<string> => {
    const resp = await this.usersClient.isAuthed();
    if (resp.status !== 200) {
      this.resetUser();
      return resp.status === 403 ? "" : errServer;
    }
    this.props.login.authed = true;
    return "";
  };

  getCaptchaID = async (): Promise<string> => {
    const resp = await this.usersClient.getCaptchaID();
    if (resp.status !== 200) {
      return errServer;
    }
    this.props.login.captchaID = resp.data.id;
    return "";
  };

  setPwd = async (oldPwd: string, newPwd: string): Promise<string> => {
    const resp = await this.usersClient.setPwd(oldPwd, newPwd);
    return resp.status === 200 ? "" : errServer;
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
        Env().alertMsg("language package not found");
    }
  };

  setTheme = (theme: string) => {
    this.props.login.preferences.theme = theme;
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
    return resp.status === 200 ? "" : errServer;
  };

  setClientCfgRemote = async (cfg: ClientConfigMsg): Promise<string> => {
    const resp = await this.settingsClient.setClientCfg(cfg);
    return resp.status === 200 ? "" : errServer;
  };

  setClientCfg = (cfg: ClientConfigMsg) => {
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
    return resp.status === 200 ? "" : errServer;
  };

  getClientCfg = async (): Promise<string> => {
    const resp = await this.settingsClient.getClientCfg();
    if (resp.status !== 200) {
      return errServer;
    }
    const clientCfg = resp.data as ClientConfigMsg;
    this.props.ui.siteName = clientCfg.siteName;
    this.props.ui.siteDesc = clientCfg.siteDesc;
    this.props.ui.bg = clientCfg.bg;
    this.props.ui.captchaEnabled = clientCfg.captchaEnabled;
    return "";
  };

  syncLan = async (): Promise<string> => {
    const url = this.props.login.preferences.lanPackURL;
    if (url === "") {
      const lan = this.props.login.preferences.lan;
      if (lan === "en_US" || lan === "zh_CN") {
        // fallback to build-in language pack
        this.props.msg.lan = lan;
        this.props.msg.pkg = MsgPackage.get(lan);
      } else {
        ErrorLogger().error(`syncLan: unexpected lan ${lan}`);
        this.props.msg.lan = "en_US";
        this.props.msg.pkg = MsgPackage.get("en_US");
      }
      return "";
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
      // TODO: should warning here
      return "";
    }
    this.props.msg.lan = resp.data.lan;
    this.props.msg.pkg = Map<string, string>(resp.data);
    return "";
  };

  setFilesOrderBy = (orderBy: string, order: boolean) => {
    this.props.filesInfo.orderBy = orderBy;
    this.props.filesInfo.order = order;
  };

  setUploadingsOrderBy = (orderBy: string, order: boolean) => {
    this.props.uploadingsInfo.orderBy = orderBy;
    this.props.uploadingsInfo.order = order;
  };

  setSharingsOrderBy = (orderBy: string, order: boolean) => {
    this.props.sharingsInfo.orderBy = orderBy;
    this.props.sharingsInfo.order = order;
  };

  sortFiles = (columnName: string, order: boolean) => {
    let orderByKey = 0;
    switch (columnName) {
      case this.props.msg.pkg.get("item.name"):
        orderByKey = 0;
        break;
      case this.props.msg.pkg.get("item.type"):
        orderByKey = 1;
        break;
      default:
        orderByKey = 2;
    }
    const rows = this.props.filesInfo.items.map((item: MetadataResp): Row => {
      return {
        val: item,
        sortVals: List([item.name, item.isDir ? "d" : "f", item.modTime]),
      };
    });

    const sortedFiles = sortRows(rows, orderByKey, order).map(
      (row): MetadataResp => {
        return row.val as MetadataResp;
      }
    );

    this.setFilesOrderBy(columnName, order);
    this.updateItems(sortedFiles);
  };

  sortUploadings = (columnName: string, order: boolean) => {
    const orderByKey = 0;
    const rows = this.props.uploadingsInfo.uploadings.map(
      (uploading: UploadEntry): Row => {
        return {
          val: uploading,
          sortVals: List([uploading.filePath]),
        };
      }
    );

    const sorted = sortRows(rows, orderByKey, order).map((row) => {
      return row.val as UploadEntry;
    });

    this.setUploadingsOrderBy(columnName, order);
    this.updateUploadings(sorted);
  };

  sortSharings = (columnName: string, order: boolean) => {
    const orderByKey = 0;
    const rows = this.props.sharingsInfo.sharings
      .keySeq()
      .map((sharingPath: string): Row => {
        return {
          val: sharingPath,
          sortVals: List([sharingPath]),
        };
      })
      .toList();

    let sorted = Map<string, string>();
    sortRows(rows, orderByKey, order).forEach((row) => {
      const sharingPath = row.val as string;
      sorted = sorted.set(
        sharingPath,
        this.props.sharingsInfo.sharings.get(sharingPath)
      );
    });

    this.setSharingsOrderBy(columnName, order);
    this.updateSharings(sorted);
  };

  resetUsedSpace = async (userID: string): Promise<string> => {
    const resp = await this.usersClient.resetUsedSpace(userID);
    return resp.status == 200 ? "" : errServer;
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
