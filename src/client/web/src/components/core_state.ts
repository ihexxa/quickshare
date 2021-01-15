import { List } from "immutable";

import { Props as PanelProps } from "./panel";
import { Item } from "./browser";
import { UploadInfo } from "../client";


export interface IContext {
  update: (targetStatePatch: any) => void;
}

export interface ICoreState {
  ctx: IContext;
  panel: PanelProps;
}

export function init(): ICoreState {
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
