import { login, logout } from "../api_auth";
import { config } from "../../config";

const serverAddr = config.serverAddr;
const testId = config.testId;
const testPwd = config.testPwd;

export function testAuth() {
  return testLogin()
    .then(testLogout)
    .catch(err => {
      console.error("auth: fail", err);
    });
}

export function testLogin() {
  return login(serverAddr, testId, testPwd).then(ok => {
    if (ok === true) {
      console.log("login api: ok");
    } else {
      throw new Error("login api: failed");
    }
  });
}

export function testLogout() {
  return logout(serverAddr).then(ok => {
    if (ok === true) {
      console.log("logout api: ok");
    } else {
      throw new Error("logout api: failed");
    }
  });
}
