import { Map, Set } from "immutable";

import { msgs as enMsgs } from "./en_US";
import { msgs as cnMsgs } from "./zh_CN";
import { msgs as esMsgs } from "./es_ES";

export class Msger {
  private msgs: Map<string, string>;
  constructor(msgs: Map<string, string>) {
    this.msgs = msgs;
  }
  m(key: string): string {
    return this.msgs.get(key, "");
  }
}

export class MsgPackage {
  static get(key: string): Map<string, string> {
    switch (key) {
      case "en_US":
        return Map(enMsgs);
      case "zh_CN":
        return Map(cnMsgs);
      case "es_ES":
        return Map(esMsgs);
      default:
        return Map(enMsgs);
    }
  }
}

export function isValidLanPack(lanPackObject: any): boolean {
  const topLevelkeys = Set(Object.keys(lanPackObject));
  if (!topLevelkeys.has("lan") && !topLevelkeys.has("pkg")) {
    return false;
  }

  const gotKeys = Set(Object.keys(lanPackObject.pkg));
  let missingKeys = Set<string>();
  lanPackKeys.forEach((key: string) => {
    if (!gotKeys.has(key)) {
      missingKeys = missingKeys.add(key);
    }
  });

  // TODO: provide better error report?
  if (missingKeys.size > 0) {
    console.error(missingKeys);
  }
  return missingKeys.size > 0;
}

export const lanPackKeys = Set<string>([
  "stateMgr.cap.fail",
  "browser.upload.del.fail",
  "browser.folder.add.fail",
  "browser.del.fail",
  "browser.move.fail",
  "browser.share.add.fail",
  "browser.share.del.fail",
  "browser.share.del",
  "browser.share.add",
  "browser.share.title",
  "browser.share.desc",
  "browser.upload.title",
  "browser.upload.desc",
  "browser.folder.name",
  "browser.folder.add",
  "browser.upload",
  "browser.delete",
  "browser.paste",
  "browser.select",
  "browser.deselect",
  "browser.selectAll",
  "browser.stop",
  "browser.location",
  "browser.item.title",
  "browser.used",
  "panes.close",
  "login.logout.fail",
  "login.username",
  "login.captcha",
  "login.pwd",
  "login.login",
  "login.logout",
  "settings.pwd.notSame",
  "settings.pwd.empty",
  "settings.pwd.notChanged",
  "update",
  "settings.pwd.old",
  "settings.pwd.new1",
  "settings.pwd.new2",
  "settings",
  "settings.chooseLan",
  "settings.pwd.update",
  "admin",
  "update.ok",
  "update.fail",
  "delete.fail",
  "delete.ok",
  "delete",
  "spaceLimit",
  "uploadLimit",
  "downloadLimit",
  "add.fail",
  "add.ok",
  "role.delete.warning",
  "user.id",
  "user.add",
  "user.name",
  "user.role",
  "user.password",
  "add",
  "admin.users",
  "role.add",
  "role.name",
  "admin.roles",
  "zhCN",
  "enUS",
  "move.fail",
  "share.404.title",
  "share.404.desc",
  "upload.404.title",
  "upload.404.desc",
  "detail",
  "refresh",
  "refresh-hint",
  "pane.login",
  "pane.admin",
  "pane.settings",
  "logout.confirm",
  "unauthed",
  "err.tooManyUploads",
  "login.role",
  "user.profile",
  "user.downLimit",
  "user.upLimit",
  "user.spaceLimit",
  "cfg.siteName",
  "cfg.siteDesc",
  "cfg.bg",
  "cfg.bg.url",
  "cfg.bg.repeat",
  "cfg.bg.pos",
  "cfg.bg.align",
  "reset",
  "bg.url.alert",
  "bg.pos.alert",
  "bg.align.alert",
  "prefer.theme",
  "prefer.theme.url",
  "settings.customLan",
  "settings.lanPackURL",
]);
