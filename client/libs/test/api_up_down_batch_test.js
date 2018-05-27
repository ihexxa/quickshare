import axios from "axios";
import md5 from "md5";

import { config } from "../../config";
import { testUpload } from "./api_upload_test";
import { list, del } from "../api_share";
import { testLogin, testLogout } from "./api_auth_test";

export function testUpDownBatch() {
  const fileInfos = [
    {
      fileName: "test_2MB_1",
      content: new Array(1024 * 1024 * 2).join("x")
    },
    {
      fileName: "test_1MB_1",
      content: new Array(1024 * 1024 * 1).join("x")
    },
    {
      fileName: "test_2MB_2",
      content: new Array(1024 * 1024 * 2).join("x")
    },
    {
      fileName: "test_1B",
      content: `${new Array(3).join("o")}${new Array(3).join("x")}`
    }
  ];

  return testLogin()
    .then(() => {
      const promises = fileInfos.map(info => {
        const file = new File([info.content], info.fileName, {
          type: "text/plain"
        });

        return testUpAndDownOneFile(file, info.fileName);
      });

      return Promise.all(promises);
    })
    .then(() => {
      testLogout();
    })
    .catch(err => console.error(err));
}

export function testUpAndDownOneFile(file, fileName) {
  return delTestFile(fileName)
    .then(() => testUpload(file))
    .then(shareId => testDownload(shareId, file))
    .catch(err => console.error(err));
}

function delTestFile(fileName) {
  return list().then(infos => {
    const info = infos.find(info => {
      return info.PathLocal === fileName;
    });

    if (info == null) {
      console.log("up-down: file not found", fileName);
    } else {
      return del(info.Id);
    }
  });
}

function testDownload(shareId, file) {
  return axios
    .get(`${config.serverAddr}/download?shareid=${shareId}`)
    .then(response => {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = event => {
          const upHash = md5(event.target.result);
          const downHash = md5(response.data);
          if (upHash !== downHash) {
            console.error(
              "up&down: hash unmatch",
              file.name,
              upHash,
              downHash,
              upHash.length,
              downHash.length
            );
          } else {
            console.log("up&down: ok: hash match", file.name, upHash, downHash);
            resolve();
          }
        };

        reader.onerror = err => reject(err);

        reader.readAsText(file);
      });
    });
}
