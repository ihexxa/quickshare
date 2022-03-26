export class WebEnv {
  constructor() {}

  alertMsg = (msg: string) => {
    if (alert != null) {
      alert(msg);
    } else {
      console.log(msg);
    }
  };

  confirmMsg = (msg: string): boolean => {
    if (confirm != null) {
      return confirm(msg);
    } else {
      console.warn(`${msg}: return yes (confirm is not implemented)`);
      return true;
    }
  };
}

export interface IEnv {
  alertMsg: (msg: string) => void;
  confirmMsg: (msg: string) => boolean;
}

let env = new WebEnv();
export const Env = (): IEnv => env;
export const SetEnv = (expectedEnv: IEnv) => {
  env = expectedEnv;
};
