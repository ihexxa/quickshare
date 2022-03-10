import { BaseClient, Response, userIDParam, Quota, Preferences } from "./";

export class UsersClient extends BaseClient {
  constructor(url: string) {
    super(url);
  }

  login = (
    user: string,
    pwd: string,
    captchaId: string,
    captchaInput: string
  ): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/users/login`,
      data: {
        user,
        pwd,
        captchaId,
        captchaInput,
      },
    });
  };

  logout = (): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/users/logout`,
    });
  };

  isAuthed = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/users/isauthed`,
    });
  };

  setPwd = (oldPwd: string, newPwd: string): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/users/pwd`,
      data: {
        oldPwd,
        newPwd,
      },
    });
  };

  setUser = (id: string, role: string, quota: Quota): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/users/`,
      data: {
        id,
        role,
        quota,
      },
    });
  };

  forceSetPwd = (userID: string, newPwd: string): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/users/pwd/force-set`,
      data: {
        id: userID,
        newPwd,
      },
    });
  };

  // token cookie is set by browser
  addUser = (name: string, pwd: string, role: string): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/users/`,
      data: {
        name,
        pwd,
        role,
      },
    });
  };

  delUser = (userID: string): Promise<Response> => {
    return this.do({
      method: "delete",
      url: `${this.url}/v1/users/`,
      params: {
        [userIDParam]: userID,
      },
    });
  };

  listUsers = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/users/list`,
      params: {},
    });
  };

  addRole = (role: string): Promise<Response> => {
    return this.do({
      method: "post",
      url: `${this.url}/v1/roles/`,
      data: { role },
    });
  };

  delRole = (role: string): Promise<Response> => {
    return this.do({
      method: "delete",
      url: `${this.url}/v1/roles/`,
      data: { role },
    });
  };

  listRoles = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/roles/list`,
      params: {},
    });
  };

  self = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/users/self`,
      params: {},
    });
  };

  getCaptchaID = (): Promise<Response> => {
    return this.do({
      method: "get",
      url: `${this.url}/v1/captchas/`,
      params: {},
    });
  };

  setPreferences = (prefers: Preferences): Promise<Response> => {
    return this.do({
      method: "patch",
      url: `${this.url}/v1/users/preferences`,
      data: {
        preferences: prefers,
      },
    });
  };

  resetUsedSpace = (userID: string): Promise<Response> => {
    return this.do({
      method: "put",
      url: `${this.url}/v1/users/used-space`,
      data: {
        userID,
      },
    });
  };
}
