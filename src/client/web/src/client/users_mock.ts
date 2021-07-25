// TODO: replace this with jest mocks
import { Response } from "./";

export class MockUsersClient {
  private url: string;
  private loginMockResp: Promise<Response>;
  private logoutMockResp: Promise<Response>;
  private isAuthedMockResp: Promise<Response>;
  private setPwdMockResp: Promise<Response>;
  private addUserMockResp: Promise<Response>;
  private delUserMockResp: Promise<Response>;
  private listUsersMockResp: Promise<Response>;
  private addRoleMockResp: Promise<Response>;
  private delRoleMockResp: Promise<Response>;
  private listRolesMockResp: Promise<Response>;

  constructor(url: string) {
    this.url = url;
  }

  loginMock = (resp: Promise<Response>) => {
    this.loginMockResp = resp;
  }
  logoutMock = (resp: Promise<Response>) => {
    this.logoutMockResp = resp;
  }
  isAuthedMock = (resp: Promise<Response>) => {
    this.isAuthedMockResp = resp;
  }
  setPwdMock = (resp: Promise<Response>) => {
    this.setPwdMockResp = resp;
  }
  addUserMock = (resp: Promise<Response>) => {
    this.addUserMockResp = resp;
  }
  delUserMock = (resp: Promise<Response>) => {
    this.delUserMockResp = resp;
  }
  listUsersMock = (resp: Promise<Response>) => {
    this.listUsersMockResp = resp;
  }
  addRoleMock = (resp: Promise<Response>) => {
    this.addRoleMockResp = resp;
  }
  delRoleMock = (resp: Promise<Response>) => {
    this.delRoleMockResp = resp;
  }
  listRolesMock = (resp: Promise<Response>) => {
    this.listRolesMockResp = resp;
  }

  login = (user: string, pwd: string): Promise<Response> => {
    return this.loginMockResp;
  }

  logout = (): Promise<Response> => {
    return this.logoutMockResp;
  }

  isAuthed = (): Promise<Response> => {
    return this.isAuthedMockResp;
  }

  setPwd = (oldPwd: string, newPwd: string): Promise<Response> => {
    return this.setPwdMockResp;
  }

  addUser = (name: string, pwd: string, role: string): Promise<Response> => {
    return this.addUserMockResp;
  }

  delUser = (userID: string): Promise<Response> => {
    return this.delUserMockResp;
  }
  
  listUsers = (): Promise<Response> => {
    return this.listUsersMockResp;
  }

  addRole = (role: string): Promise<Response> => {
    return this.addRoleMockResp;
  }

  delRole = (role: string): Promise<Response> => {
    return this.delRoleMockResp;
  }

  listRoles = (): Promise<Response> => {
    return this.listRolesMockResp;
  }
}
