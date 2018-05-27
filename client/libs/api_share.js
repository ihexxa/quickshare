import axios from "axios";
import { config } from "../config";

export const del = shareId => {
  return axios
    .delete(`${config.serverAddr}/fileinfo?shareid=${shareId}`)
    .then(response => response.data.Code === 200);
};

export const list = () => {
  return axios.get(`${config.serverAddr}/fileinfo`).then(response => {
    // TODO check status code
    return response.data.List;
  });
};

export const shadowId = shareId => {
  const act = "shadowid";
  return axios
    .patch(`${config.serverAddr}/fileinfo?act=${act}&shareid=${shareId}`)
    .then(response => {
      return response.data.ShareId;
    });
};

export const publishId = shareId => {
  const act = "publishid";
  return axios
    .patch(`${config.serverAddr}/fileinfo?act=${act}&shareid=${shareId}`)
    .then(response => {
      return response.data.ShareId;
    });
};

export const setDownLimit = (shareId, downLimit) => {
  const act = "setdownlimit";
  return axios
    .patch(
      `${
        config.serverAddr
      }/fileinfo?act=${act}&shareid=${shareId}&downlimit=${downLimit}`
    )
    .then(response => response.data.Code === 200);
};

export const addLocalFiles = () => {
  const act = "addlocalfiles";
  return axios
    .patch(`${config.serverAddr}/fileinfo?act=${act}`)
    .then(response => response.data.Code === 200);
};
