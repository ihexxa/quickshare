import { List, Set, Map } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fg.worker";

// import { Props as PanelProps } from "./root_frame";
import { BrowserProps } from "./browser";
import { PanesProps } from "./panes";
import { LoginProps } from "./pane_login";
import { AdminProps } from "./pane_admin";

import { MsgPackage } from "../i18n/msger";
import { Item } from "./browser";
import { UploadInfo, User } from "../client";
import { initUploadMgr, IWorker } from "../worker/upload_mgr";

export interface MsgProps {
  lan: string;
  pkg: Map<string, string>;
}
export interface ICoreState {
  // panel: PanelProps;
  isVertical: boolean;
  browser: BrowserProps;
  panes: PanesProps;
  login: LoginProps;
  admin: AdminProps;
  msg: MsgProps;
  // settings: SettingsProps;
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
      items: List<Item>([]),
      sharings: List<string>([]),
      isSharing: false,
      uploadings: List<UploadInfo>([]),
      uploadValue: "",
      uploadFiles: List<File>([]),
    },
    panes: {
      displaying: "browser",
      paneNames: Set<string>(["settings", "login", "admin"]),
    },
    login: {
      userRole: "",
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
    }
  };
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}
