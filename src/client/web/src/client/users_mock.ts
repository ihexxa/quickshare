// TODO: replace this with jest mocks
import { Response } from "./";

export class MockUsersClient {
  private url: string;
  private loginMockResp: Promise<Response>;
  private logoutMockResp: Promise<Response>;
  private isAuthedMockResp: Promise<Response>;
  private setPwdMockResp: Promise<Response>;

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

}
