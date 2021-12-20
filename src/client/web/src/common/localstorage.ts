export interface ILocalStorage {
  get: (key: string) => string;
  set: (key: string, val: string) => void;
}

export const errNoLocalStorage = "local storage is not supported";

class LocalStorage {
  constructor() {}

  get(key: string): string {
    if (window != null && window.localStorage != null) {
      const val = window.localStorage.getItem(key);
      return val && val != "undefined" && val != "null" ? val : "";
    }

    return "";
  }

  set(key: string, val: string) {
    if (window != null && window.localStorage != null) {
      window.localStorage.setItem(key, val);
    } else {
      console.error(errNoLocalStorage);
    }
  }
}

var localStorage: LocalStorage;
export const Storage = () => {
  if (localStorage == null) {
    localStorage = new LocalStorage();
  }
  return localStorage;
};
