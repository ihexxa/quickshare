import { List, Set, Map } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fg.worker";
import { UploadEntry } from "../worker/interface";

import { BrowserProps } from "./browser";
import { PanesProps } from "./panes";
import { LoginProps } from "./pane_login";
import { AdminProps } from "./pane_admin";

import { MsgPackage } from "../i18n/msger";
import { User, MetadataResp } from "../client";
import { initUploadMgr, IWorker } from "../worker/upload_mgr";

export interface MsgProps {
  lan: string;
  pkg: Map<string, string>;
}

export interface UIProps {
  isVertical: boolean;
  bg: {
    url: string;
    repeat: string;
    position: string;
    align: string;
  };
}
export interface ICoreState {
  browser: BrowserProps;
  panes: PanesProps;
  login: LoginProps;
  admin: AdminProps;
  ui: UIProps;
  msg: MsgProps;
}

export function newWithWorker(worker: IWorker): ICoreState {
  initUploadMgr(worker);
  return initState();
}

export function newState(): ICoreState {
  const worker = window.Worker == null ? new FgWorker() : new BgWorker();
  initUploadMgr(worker);

  return initState();
}

export function initState(): ICoreState {
  return {
    browser: {
      dirPath: List<string>(["."]),
      items: List<MetadataResp>([]),
      sharings: List<string>([]),
      isSharing: false,
      uploadings: List<UploadEntry>([]),
      uploadValue: "",
      uploadFiles: List<File>([]),
      tab: "",
    },
    panes: {
      // which pane is displaying
      displaying: "browser",
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
