import { List, Set } from "immutable";

import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fgworker";

import { Props as PanelProps } from "./panel";
import { Item } from "./browser";
import { UploadInfo } from "../client";
import { UploadMgr, IWorker } from "../worker/upload_mgr";

export interface IContext {
  update: (targetStatePatch: any) => void;
}

export interface ICoreState {
  ctx: IContext;
  panel: PanelProps;
}

export function initWithWorker(worker: IWorker): ICoreState {
  UploadMgr.init(worker);
  return initState();
}

export function init(): ICoreState {
  const scripts = Array.from(document.querySelectorAll("script"));
  if (!Worker) {
    alert("web worker is not supported");
  }

  const worker = new BgWorker();
  UploadMgr.init(worker);
  return initState();
}

export function initState(): ICoreState {
  return {
    ctx: null,
    panel: {
      displaying: "browser",
      authPane: {
        authed: false,
      },
      browser: {
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
