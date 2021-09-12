import { List, Set, Map } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fg.worker";

import { BrowserProps } from "./browser";
import { PanesProps } from "./panes";
import { LoginProps } from "./pane_login";
import { AdminProps } from "./pane_admin";

import { MsgPackage } from "../i18n/msger";
import { Item } from "./browser";
import { UploadInfo, User, MetadataResp } from "../client";
import { initUploadMgr, IWorker } from "../worker/upload_mgr";

export interface MsgProps {
  lan: string;
  pkg: Map<string, string>;
}

export interface UIProps {
  // background: url("/static/img/textured_paper.png") repeat fixed center;
  wallpaper: string;
  repeat: string;
  position: string;
  align: string;
}
export interface ICoreState {
  isVertical: boolean;
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
    isVertical: isVertical(),
    browser: {
      isVertical: isVertical(),
      dirPath: List<string>(["."]),
      items: List<MetadataResp>([]),
      sharings: List<string>([]),
      isSharing: false,
      uploadings: List<UploadInfo>([]),
      uploadValue: "",
      uploadFiles: List<File>([]),
      tab: "",
    },
    panes: {
      displaying: "browser",
      paneNames: Set<string>(["settings", "login", "admin"]),
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
      wallpaper: "",
      repeat: "",
      position: "",
      align: "",
    },
  };
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}
