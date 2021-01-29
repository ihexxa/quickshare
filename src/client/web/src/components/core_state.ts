import { List, Set } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fgworker";

import { Props as PanelProps } from "./root_frame";
import { Item } from "./browser";
import { UploadInfo } from "../client";
import { UploadMgr, IWorker } from "../worker/upload_mgr";

export class BaseUpdater {
  public static props: any;
  public static init = (props: any) => (BaseUpdater.props = { ...props });
  public static apply = (prevState: ICoreState): ICoreState => {
    throw Error("apply is not implemented");
  };
}

export interface IContext {
  update: (targetStatePatch: any) => void;
}

export interface ICoreState {
  ctx: IContext;
  panel: PanelProps;
  isVertical: boolean;
}

export function initWithWorker(worker: IWorker): ICoreState {
  UploadMgr.init(worker);
  return initState();
}

export function init(): ICoreState {
  const scripts = Array.from(document.querySelectorAll("script"));
  const worker = Worker == null ? new FgWorker() : new BgWorker();

  UploadMgr.init(worker);
  return initState();
}

export function isVertical(): boolean {
  return window.innerWidth <= window.innerHeight;
}

export function initState(): ICoreState {
  return {
    ctx: null,
    isVertical: isVertical(),
    panel: {
      displaying: "browser",
      authPane: {
        authed: false,
      },
      browser: {
        isVertical: isVertical(),
        dirPath: List<string>(["."]),
        items: List<Item>([]),
        uploadings: List<UploadInfo>([]),
        uploadValue: "",
        uploadFiles: List<File>([]),
      },
      panes: {
        displaying: "",
        paneNames: Set<string>(["settings", "login"]),
        login: {
          authed: false,
        },
      },
    },
  };
}

export function mockState(): ICoreState {
  return {
    ctx: undefined,
    isVertical: false,
    panel: {
      displaying: "browser",
      authPane: {
        authed: false,
      },
      browser: {
        isVertical: false,
        dirPath: List<string>(["."]),
        items: List<Item>([]),
        uploadings: List<UploadInfo>([]),
        uploadValue: "",
        uploadFiles: List<File>([]),
      },
      panes: {
        displaying: "",
        paneNames: Set<string>(["settings", "login"]),
        login: {
          authed: false,
        },
      },
    },
  };
}
