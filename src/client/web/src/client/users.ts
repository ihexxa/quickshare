import axios, { AxiosRequestConfig, AxiosResponse } from "axios";

export class UsersClient {
  private url: string;

  constructor(url: string) {
    this.url = url;
  }

  async do(config: AxiosRequestConfig): Promise<AxiosResponse<any> | null> {
    try {
      return await axios(config);
    } catch (e) {
      return null;
    }
  }

  async login(user: string, pwd: string): Promise<number> {
    const resp = await this.do({
      method: "post",
      url: `${this.url}/v1/users/login`,
      data: {
        user,
        pwd,
      },
    });
    return resp != null ? resp.status : 500;
  }

  // token cookie is set by browser
  async logout(): Promise<number> {
    const resp = await this.do({
      method: "post",
      url: `${this.url}/v1/users/logout`,
    });
    return resp != null ? resp.status : 500;
  }

  // token cookie is set by browser
  async setPwd(oldPwd: string, newPwd: string): Promise<number> {
    const resp = await this.do({
      method: "patch",
      url: `${this.url}/v1/users/pwd`,
      data: {
        oldPwd,
        newPwd,
      },
    });
    return resp != null ? resp.status : 500;
  }
}
