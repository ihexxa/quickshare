// TODO: replace this with jest mocks
import { Response, Quota } from "./";

export interface UsersClientResps {
  loginMockResp: Response;
  logoutMockResp: Response;
  isAuthedMockResp: Response;
  setPwdMockResp: Response;
  setUserMockResp: Response;
  forceSetPwdMockResp: Response;
  addUserMockResp: Response;
  delUserMockResp: Response;
  listUsersMockResp: Response;
  addRoleMockResp: Response;
  delRoleMockResp: Response;
  listRolesMockResp: Response;
  selfMockResp: Response;
  getCaptchaIDMockResp: Response;
}

export const resps = {
  loginMockResp: { status: 200, statusText: "", data: {} },
  logoutMockResp: { status: 200, statusText: "", data: {} },
  isAuthedMockResp: { status: 200, statusText: "", data: {} },
  setPwdMockResp: { status: 200, statusText: "", data: {} },
  setUserMockResp: { status: 200, statusText: "", data: {} },
  forceSetPwdMockResp: { status: 200, statusText: "", data: {} },
  addUserMockResp: { status: 200, statusText: "", data: {} },
  delUserMockResp: { status: 200, statusText: "", data: {} },
  listUsersMockResp: {
    status: 200,
    statusText: "",
    data: {
      users: [
        {
          id: "0",
          name: "mock_username0",
          pwd: "mock_pwd0",
          role: "mock_role0",
          usedSpace: "128",
          quota: {
            spaceLimit: "1024",
            uploadSpeedLimit: 1024,
            downloadSpeedLimit: 1024,
          },
        },
        {
          id: "1",
          name: "mock_username1",
          pwd: "mock_pwd1",
          role: "mock_role1",
          usedSpace: "256",
          quota: {
            spaceLimit: "1024",
            uploadSpeedLimit: 1024,
            downloadSpeedLimit: 1024,
          },
        },
      ],
    },
  },
  addRoleMockResp: { status: 200, statusText: "", data: {} },
  delRoleMockResp: { status: 200, statusText: "", data: {} },
  listRolesMockResp: {
    status: 200,
    statusText: "",
    data: {
      roles: ["admin", "users", "visitor"],
    },
  },
  selfMockResp: {
    status: 200,
    statusText: "",
    data: {
      id: "0",
      name: "mockUser",
      role: "admin",
      usedSpace: "256",
      quota: {
        spaceLimit: "7",
        uploadSpeedLimit: 3,
        downloadSpeedLimit: 3,
      },
    },
  },
  getCaptchaIDMockResp: {
    status: 200,
    statusText: "",
    data: {
      id: "mockCaptchaID",
    },
  },
};
export class MockUsersClient {
  private url: string;
  private resps: UsersClientResps;

  constructor(url: string) {
    this.url = url;
    this.resps = resps;
  }

  setMock = (resps: UsersClientResps) => {
    this.resps = resps;
  };

  wrapPromise = (resp: any): Promise<any> => {
    return new Promise<any>((resolve) => {
      resolve(resp);
    });
  };

  login = (user: string, pwd: string): Promise<Response> => {
    return this.wrapPromise(this.resps.loginMockResp);
  };

  logout = (): Promise<Response> => {
    return this.wrapPromise(this.resps.logoutMockResp);
  };

  isAuthed = (): Promise<Response> => {
    return this.wrapPromise(this.resps.isAuthedMockResp);
  };

  setPwd = (oldPwd: string, newPwd: string): Promise<Response> => {
    return this.wrapPromise(this.resps.setPwdMockResp);
  };

  setUser = (id: string, role: string, quota: Quota): Promise<Response> => {
    return this.wrapPromise(this.resps.setUserMockResp);
  };

  forceSetPwd = (userID: string, newPwd: string): Promise<Response> => {
    return this.wrapPromise(this.resps.forceSetPwdMockResp);
  };

  addUser = (name: string, pwd: string, role: string): Promise<Response> => {
    return this.wrapPromise(this.resps.addUserMockResp);
  };

  delUser = (userID: string): Promise<Response> => {
    return this.wrapPromise(this.resps.delUserMockResp);
  };

  listUsers = (): Promise<Response> => {
    return this.wrapPromise(this.resps.listUsersMockResp);
  };

  addRole = (role: string): Promise<Response> => {
    return this.wrapPromise(this.resps.addRoleMockResp);
  };

  delRole = (role: string): Promise<Response> => {
    return this.wrapPromise(this.resps.delRoleMockResp);
  };

  listRoles = (): Promise<Response> => {
    return this.wrapPromise(this.resps.listRolesMockResp);
  };

  self = (): Promise<Response> => {
    return this.wrapPromise(this.resps.selfMockResp);
  };

  getCaptchaID = (): Promise<Response> => {
    return this.wrapPromise(this.resps.getCaptchaIDMockResp);
  };
}
