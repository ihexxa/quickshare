import * as React from "react";
import { Map } from "immutable";

export interface Hotkey {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  meta?: boolean;
  repeat?: boolean;
  ev?: KeyboardEvent;
}

export type HotkeyCb = (hotKey?: Hotkey) => void;

export class HotkeyHandler {
  private keyMap: Map<string, HotkeyCb>;

  constructor() {
    this.keyMap = Map<string, HotkeyCb>();
  }

  getSign = (hk: Hotkey): string => {
    let sign = hk.key;
    sign = hk.ctrl != null && hk.ctrl ? `${sign}+ctrl` : sign;
    sign = hk.shift != null && hk.shift ? `${sign}+shift` : sign;
    sign = hk.alt != null && hk.alt ? `${sign}+alt` : sign;
    sign = hk.meta != null && hk.meta ? `${sign}+meta` : sign;
    sign = hk.repeat != null && hk.repeat ? `${sign}+repeat` : sign;

    return sign;
  };

  add = (hk: Hotkey, handler: HotkeyCb) => {
    const sign = this.getSign(hk);
    this.keyMap = this.keyMap.set(sign, handler);
  };

  handle = (ev: KeyboardEvent) => {
    const hotKey = {
      key: ev.key,
      ctrl: ev.ctrlKey,
      shift: ev.shiftKey,
      alt: ev.altKey,
      meta: ev.metaKey,
      repeat: ev.repeat,
    };
    const sign = this.getSign(hotKey);

    if (this.keyMap.has(sign)) {
      const handler = this.keyMap.get(sign);
      handler(hotKey);
    }
  };

  printMap = () => {
    console.log(this.keyMap.toMap());
  };
}
