import { List, Set, Map } from "immutable";

import { UploadEntry } from "../worker/interface";
import { MsgPackage } from "../i18n/msger";
import { User, MetadataResp } from "../client";
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
} from "../common/controls";
import { LoginProps } from "./pane_login";
import { AdminProps } from "./pane_admin";

export interface MsgProps {
  lan: string;
  pkg: Map<string, string>;
}

export interface UIProps {
  isVertical: boolean;
  siteName: string;
  siteDesc: string;
  bg: {
    url: string;
    repeat: string;
    position: string;
    align: string;
  };
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
  return {
    filesInfo: {
      dirPath: List<string>([]),
      items: List<MetadataResp>([]),
      isSharing: false,
    },
    uploadingsInfo: {
      uploadings: List<UploadEntry>([]),
      uploadFiles: List<File>([]),
    },
    sharingsInfo: {
      sharings: List<string>([]),
    },
    admin: {
      users: Map<string, User>(),
      roles: Set<string>(),
    },
    login: {
      userID: "",
      userName: "",
      userRole: "",
      usedSpace: "0",
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
        },
        cssURL: "",
        lanPackURL: "",
        lan: "en_US",
      },
    },
    msg: {
      lan: "en_US",
      pkg: MsgPackage.get("en_US"),
    },
    ui: {
      isVertical: isVertical(),
      siteName: "",
      siteDesc: "",
      bg: {
        url: "",
        repeat: "",
        position: "",
        align: "",
      },
      control: {
        controls: Map<string, string>({
          [panelTabs]: "filesPanel",
          [settingsDialogCtrl]: ctrlOff,
          [settingsTabsCtrl]: "preferencePane",
          [sharingCtrl]: ctrlOff,
          [filesViewCtrl]: "rows",
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
        }),
      },
    },
  };
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}
