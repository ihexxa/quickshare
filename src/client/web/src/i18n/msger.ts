import { Map } from "immutable";

import { msgs as enMsgs } from "./en_US";
import { msgs as cnMsgs } from "./zh_CN";

export class Msger {
  private msgs: Map<string, string>;
  constructor(msgs: Map<string, string>) {
    this.msgs = msgs;
  }
  getMsg(key: string): string {
    return this.msgs.get(key, "");
  }
}

export class MsgPackage {
  static getPkg(key: string): Map<string, string> {
    switch (key) {
      case "en-US":
        return Map(enMsgs);
      case "zh-CN":
        return Map(cnMsgs);
      default:
        return Map(enMsgs);
    }
  }
}
