// import axios, { AxiosRequestConfig, AxiosResponse } from "axios";

// TODO: replace this with jest mocks
export class MockUsersClient {
  private url: string;
  private loginResp: number;
  private logoutResp: number;
  private isAuthedResp: number;
  private setPwdResp: number;

  constructor(url: string) {
    this.url = url;
  }

  mockLoginResp(status: number) {
    this.loginResp = status;
  }
  mocklogoutResp(status: number) {
    this.logoutResp = status;
  }
  mockisAuthedResp(status: number) {
    this.isAuthedResp = status;
  }
  mocksetPwdResp(status: number) {
    this.setPwdResp = status;
  }
  
  async login(user: string, pwd: string): Promise<number> {
    return this.loginResp
  }

  // token cookie is set by browser
  async logout(): Promise<number> {
    return this.logoutResp
  }

  async isAuthed(): Promise<number> {
    return this.isAuthedResp
  }

  // token cookie is set by browser
  async setPwd(oldPwd: string, newPwd: string): Promise<number> {
    return this.setPwdResp
  }
}
