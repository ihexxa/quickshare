import axios from "axios";
import { config } from "../config";
import { makePostBody } from "./utils";

const wait = 5000; // TODO: should tune according to backend
const retryMax = 100000;
const maxUploadLen = 20 * 1024 * 1024;

// TODO: add to react-intl
const msgUploadFailed = "Fail to upload, upload is stopped.";
const msgUploadFailedAndRetry = "Fail to upload, retrying...";
const msgFileExists = "File exists.";
const msgTooBigChunk = "Too big chunk.";
const msgFileNotFound = "File not found, upload stopped.";

function randomWait() {
  return Math.random() * wait;
}

function isKnownErr(res) {
  return res != null && res.Code != null && res.Msg != null;
}

export class FileUploader {
  constructor(onStart, onProgress, onFinish, onError) {
    this.onStart = onStart;
    this.onProgress = onProgress;
    this.onFinish = onFinish;
    this.onError = onError;
    this.retry = retryMax;
    this.reader = new FileReader();

    this.uploadFile = file => {
      return this.startUpload(file);
    };

    this.startUpload = file => {
      return axios
        .post(
          `${config.serverAddr}/startupload`,
          makePostBody({
            fname: file.name
          })
        )
        .then(response => {
          if (
            response.data.ShareId == null ||
            response.data.Start === null ||
            response.data.Length === null
          ) {
            throw response;
          } else {
            this.onStart(response.data.ShareId, file.name);
            return this.upload(
              {
                shareId: response.data.ShareId,
                start: response.data.Start,
                length: response.data.Length
              },
              file
            );
          }
        })
        .catch(response => {
          // TODO: this is not good because error may not be response
          if (isKnownErr(response.data) && response.data.Code === 429) {
            setTimeout(this.startUpload, randomWait(), file);
          } else if (isKnownErr(response.data) && response.data.Code === 412) {
            this.onError(msgFileExists);
          } else if (isKnownErr(response.data) && response.data.Code === 404) {
            this.onError(msgFileNotFound);
          } else if (this.retry > 0) {
            this.retry--;
            this.onError(msgUploadFailedAndRetry);
            console.trace(response);
            setTimeout(this.startUpload, randomWait(), file);
          } else {
            this.onError(msgUploadFailed);
            console.trace(response);
          }
        });
    };

    this.prepareReader = (shareInfo, end, resolve, reject) => {
      this.reader.onerror = err => {
        reject(err);
      };

      this.reader.onloadend = event => {
        const formData = new FormData();
        formData.append("shareid", shareInfo.shareId);
        formData.append("start", shareInfo.start);
        formData.append("len", end - shareInfo.start);
        formData.append("chunk", new Blob([event.target.result]));

        const url = `${config.serverAddr}/upload`;
        const headers = {
          "Content-Type": "multipart/form-data"
        };

        try {
          axios
            .post(url, formData, { headers })
            .then(response => resolve(response))
            .catch(err => {
              reject(err);
            });
        } catch (err) {
          reject(err);
        }
      };
    };

    this.upload = (shareInfo, file) => {
      const uploaded = shareInfo.start + shareInfo.length;
      const end = uploaded < file.size ? uploaded : file.size;

      return new Promise((resolve, reject) => {
        if (
          end == null ||
          shareInfo.start == null ||
          end - shareInfo.start >= maxUploadLen
        ) {
          throw new Error(msgTooBigChunk);
        }

        const chunk = file.slice(shareInfo.start, end);
        this.prepareReader(shareInfo, end, resolve, reject);
        this.reader.readAsArrayBuffer(chunk);
      })
        .then(response => {
          if (
            response.data.ShareId == null ||
            response.data.Start == null ||
            response.data.Length == null ||
            response.data.Start !== end
          ) {
            throw response;
          } else {
            if (end < file.size) {
              this.onProgress(shareInfo.shareId, end / file.size);
              return this.upload(
                {
                  shareId: shareInfo.shareId,
                  start: shareInfo.start + shareInfo.length,
                  length: shareInfo.length
                },
                file
              );
            } else {
              return this.finishUpload(shareInfo);
            }
          }
        })
        .catch(response => {
          // possible error: response.data.Start == null || response.data.Start !== end
          if (isKnownErr(response.data) && response.data.Code === 429) {
            setTimeout(this.upload, randomWait(), shareInfo, file);
          } else if (isKnownErr(response.data) && response.data.Code === 404) {
            this.onError(msgFileNotFound);
          } else if (this.retry > 0) {
            this.retry--;
            setTimeout(this.upload, randomWait(), shareInfo, file);
            this.onError(msgUploadFailedAndRetry);
            console.trace(response);
          } else {
            this.onError(msgUploadFailed);
            console.trace(response);
          }
        });
    };

    this.finishUpload = shareInfo => {
      return axios
        .post(`${config.serverAddr}/finishupload?shareid=${shareInfo.shareId}`)
        .then(response => {
          // TODO: should check Code instead of Url
          if (response.data.ShareId != null && response.data.Start == null) {
            this.onFinish();
            return response.data.ShareId;
          } else {
            throw response;
          }
        })
        .catch(response => {
          if (isKnownErr(response.data) && response.data.Code === 429) {
            setTimeout(this.finishUpload, randomWait(), shareInfo);
          } else if (isKnownErr(response.data) && response.data.Code === 404) {
            this.onError(msgFileNotFound);
          } else if (this.retry > 0) {
            this.retry--;
            setTimeout(this.finishUpload, randomWait(), shareInfo);
            this.onError(msgUploadFailedAndRetry);
            console.trace(response);
          } else {
            this.onError(msgUploadFailed);
            console.trace(response);
          }
        });
    };
  }
}
