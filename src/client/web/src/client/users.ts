import { BaseClient, Response } from "./";


export class UsersClient extends BaseClient {
  constructor(url: string) {
    super(url);
  }

  login = (user: string, pwd: string): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/users/login`,
      data: {
        user,
        pwd,
      },
    });
  }

  // token cookie is set by browser
  logout = (): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/users/logout`,
    });
  }

  isAuthed = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/users/isauthed`,
    });
  }

  // token cookie is set by browser
  setPwd = (oldPwd: string, newPwd: string): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/users/pwd`,
      data: {
        oldPwd,
        newPwd,
      },
    });
  }
}
