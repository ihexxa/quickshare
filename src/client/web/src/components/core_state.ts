import { List } from "immutable";

// import UploadWorker = require("worker-loader!../worker/uploader.worker");
import UploadWorker from "../worker/uploader.worker";


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
  // let workerScriptName = "";
  // for (let i = 0; i < scripts.length; i++) {
  //   if (scripts[i].getAttribute("src").startsWith("static/worker.bundle.js")) {
  //     console.log(scripts[i].src);
  //     workerScriptName = scripts[i].src;
  //     break;
  //   }
  // }
  // if (workerScriptName === "") {
  //   alert("worker script not found");
  // }
  // const worker = new Worker(workerScriptName, { name: "uploader" });

  if (!Worker) {
    alert("web worker is not supported");
  }

  const worker = new UploadWorker();
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
    },
  };
}
