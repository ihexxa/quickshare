export class LocalStorage {
  static get(key: string): string {
    const val = window.localStorage.getItem(key);
    return val && val != "undefined" && val != "null" ? val : "";
  }
  static set(key: string, val: string): boolean {
    window.localStorage.setItem(key, val);
    return true;
  }
}
