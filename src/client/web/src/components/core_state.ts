import { List, Set, Map } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fg.worker";

import { Props as PanelProps } from "./root_frame";
import { Item } from "./browser";
import { UploadInfo, User } from "../client";
import { initUploadMgr, IWorker } from "../worker/upload_mgr";


export interface ICoreState {
  panel: PanelProps;
  isVertical: boolean;
}

export function newWithWorker(worker: IWorker): ICoreState {
  initUploadMgr(worker);
  return initState();
}

export function newState(): ICoreState {
  const worker = Worker == null ? new FgWorker() : new BgWorker();
  initUploadMgr(worker);

  return initState();
}

export function initState(): ICoreState {
  return {
    isVertical: isVertical(),
    panel: {
      displaying: "browser",
      authPane: {
        authed: false,
        captchaID: "",
      },
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
        userRole: "",
        displaying: "",
        paneNames: Set<string>(["settings", "login", "admin"]),
        login: {
          authed: false,
          captchaID: "",
        },
        admin: {
          users: Map<string, User>(),
          roles: Set<string>(),
        },
      },
    },
  };
}


export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}