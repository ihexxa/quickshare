import axios from "axios";
import { config } from "../config";
import { makePostBody } from "./utils";

export function login(serverAddr, adminId, adminPwd, axiosConfig) {
  return axios
    .post(
      `${serverAddr}/login`,
      makePostBody(
        {
          act: "login",
          adminid: adminId,
          adminpwd: adminPwd
        },
        axiosConfig
      )
    )
    .then(response => {
      return response.data.Code === 200;
    });
}

export function logout(serverAddr, axiosConfig) {
  return axios
    .post(`${serverAddr}/login`, makePostBody({ act: "logout" }), axiosConfig)
    .then(response => {
      return response.data.Code === 200;
    });
}
