import { List, Set, Map } from "immutable";

import { UploadEntry } from "../worker/interface";

import { FilesProps } from "./panel_files";
import { UploadingsProps } from "./panel_uploadings";
import { SharingsProps } from "./panel_sharings";
import { PanesProps } from "./panes";
import { LoginProps } from "./pane_login";
import { AdminProps } from "./pane_admin";
import { MsgPackage } from "../i18n/msger";
import { User, MetadataResp } from "../client";

export interface PanelsProps {
  displaying: string;
}

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
}
export interface ICoreState {
  panels: PanelsProps;
  filesInfo: FilesProps;
  uploadingsInfo: UploadingsProps;
  sharingsInfo: SharingsProps;
  panes: PanesProps;
  login: LoginProps;
  admin: AdminProps;
  ui: UIProps;
  msg: MsgProps;
}

export function newState(): ICoreState {
  return initState();
}

export function initState(): ICoreState {
  return {
    panels: {
      displaying: "item",
    },
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
    panes: {
      // which pane is displaying
      displaying: "",
      // which panes can be displayed
      paneNames: Set<string>([]), // "settings", "login", "admin"
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
    admin: {
      users: Map<string, User>(),
      roles: Set<string>(),
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
    },
  };
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}
