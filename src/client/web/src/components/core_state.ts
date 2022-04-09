import { List, Set, Map } from "immutable";

import { UploadEntry } from "../worker/interface";
import { MsgPackage } from "../i18n/msger";
import { User, MetadataResp, ClientConfig } from "../client";
import { FilesProps } from "./panel_files";
import { UploadingsProps } from "./panel_uploadings";
import { SharingsProps } from "./panel_sharings";
import { controlName as panelTabs } from "./root_frame";
import {
  filesViewCtrl,
  settingsTabsCtrl,
  settingsDialogCtrl,
  sharingCtrl,
  ctrlOn,
  ctrlOff,
  loadingCtrl,
} from "../common/controls";
import { LoginProps } from "./pane_login";
import { AdminProps } from "./pane_admin";

export interface MsgProps {
  lan: string;
  pkg: Map<string, string>;
}

export interface UIProps {
  clientCfg: ClientConfig;
  captchaEnabled: boolean;
  isVertical: boolean;
  control: {
    controls: Map<string, string>;
    options: Map<string, Set<string>>;
  };
}
export interface ICoreState {
  filesInfo: FilesProps;
  uploadingsInfo: UploadingsProps;
  sharingsInfo: SharingsProps;
  admin: AdminProps;
  login: LoginProps;
  ui: UIProps;
  msg: MsgProps;
}

export function newState(): ICoreState {
  return initState();
}

export function initState(): ICoreState {
  const defaultLanPackage = MsgPackage.get("en_US");
  const filesOrderBy = defaultLanPackage.get("item.name");
  const uploadingsOrderBy = defaultLanPackage.get("item.path");
  const sharingsOrderBy = defaultLanPackage.get("item.path");

  return {
    filesInfo: {
      dirPath: List<string>([]),
      items: List<MetadataResp>([]),
      isSharing: false,
      orderBy: filesOrderBy,
      order: true,
    },
    uploadingsInfo: {
      uploadings: List<UploadEntry>([]),
      uploadFiles: List<File>([]),
      orderBy: uploadingsOrderBy,
      order: true,
    },
    sharingsInfo: {
      sharings: Map<string, string>(),
      orderBy: sharingsOrderBy,
      order: true,
    },
    admin: {
      users: Map<string, User>(),
      roles: Set<string>(),
    },
    login: {
      userID: "",
      userName: "",
      userRole: "",
      extInfo: {
        usedSpace: "0",
      },
      quota: {
        spaceLimit: "0",
        uploadSpeedLimit: 0,
        downloadSpeedLimit: 0,
      },
      authed: false,
      captchaID: "",
      preferences: {
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
      },
    },
    msg: {
      lan: "en_US",
      pkg: defaultLanPackage,
    },
    ui: {
      isVertical: isVertical(),
      clientCfg: {
        siteName: "",
        siteDesc: "",
        bg: {
          url: "",
          repeat: "",
          position: "",
          align: "",
          bgColor: "",
        },
        allowSetBg: false,
        autoTheme: true,
      },
      captchaEnabled: true,
      control: {
        controls: Map<string, string>({
          [panelTabs]: "filesPanel",
          [settingsDialogCtrl]: ctrlOff,
          [settingsTabsCtrl]: "preferencePane",
          [sharingCtrl]: ctrlOff,
          [filesViewCtrl]: "rows",
          [loadingCtrl]: ctrlOff,
        }),
        options: Map<string, Set<string>>({
          [panelTabs]: Set<string>([
            "filesPanel",
            "uploadingsPanel",
            "sharingsPanel",
          ]),
          [settingsDialogCtrl]: Set<string>([ctrlOn, ctrlOff]),
          [settingsTabsCtrl]: Set<string>(["preferencePane", "managementPane"]),
          [sharingCtrl]: Set<string>([ctrlOn, ctrlOff]),
          [filesViewCtrl]: Set<string>(["rows", "table"]),
          [loadingCtrl]: Set<string>([ctrlOn, ctrlOff]),
        }),
      },
    },
  };
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}
