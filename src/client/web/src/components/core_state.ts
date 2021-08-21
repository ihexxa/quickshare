import { List, Set, Map } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fg.worker";

import { Props as PanelProps } from "./root_frame";
import { Item } from "./browser";
import { UploadInfo, User } from "../client";
import { initUploadMgr, IWorker } from "../worker/upload_mgr";

export class BaseUpdater {
  public static props: any;
  public static init = (props: any) => (BaseUpdater.props = { ...props });
  public static apply = (prevState: ICoreState): ICoreState => {
    throw Error("apply is not implemented");
  };
}

export interface ICoreState {
  panel: PanelProps;
  isVertical: boolean;
}

export function initWithWorker(worker: IWorker): ICoreState {
  initUploadMgr(worker);
  return initState();
}

export function init(): ICoreState {
  const worker = Worker == null ? new FgWorker() : new BgWorker();
  initUploadMgr(worker);

  return initState();
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
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

export function mockState(): ICoreState {
  return {
    isVertical: false,
    panel: {
      displaying: "browser",
      authPane: {
        authed: false,
        captchaID: "",
      },
      browser: {
        isVertical: false,
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
