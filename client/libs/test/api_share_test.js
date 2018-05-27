import { FileUploader } from "../api_upload";
import {
  del,
  list,
  shadowId,
  publishId,
  setDownLimit,
  addLocalFiles
} from "../api_share";
import { testLogin, testLogout } from "./api_auth_test";

const fileName = "filename";

function upload(fileName) {
  return new Promise(resolve => {
    const onStart = () => true;
    const onProgress = () => true;
    const onFinish = () => resolve();
    const onError = err => {
      throw new Error(JSON.stringify(err));
    };
    const file = new File(["foo"], fileName, {
      type: "text/plain"
    });

    const uploader = new FileUploader(onStart, onProgress, onFinish, onError);
    uploader.uploadFile(file);
  });
}

function getIdFromList(list, fileName) {
  if (list == null) {
    throw new Error("list: list fail");
  }

  // TODO: should verify file name
  const filterInfo = list.find(info => {
    return info.PathLocal.includes(fileName);
  });

  if (filterInfo == null) {
    console.error(list);
    throw new Error("list: file name not found");
  } else {
    return filterInfo.Id;
  }
}

function delWithName(fileName) {
  return list().then(infoList => {
    const infoToDel = infoList.find(info => {
      return info.PathLocal.includes(fileName);
    });

    if (infoToDel == null) {
      console.warn("delWithName: name not found");
    } else {
      return del(infoToDel.Id);
    }
  });
}

export function testShadowPublishId() {
  return testLogin()
    .then(() => upload(fileName))
    .then(list)
    .then(infoList => {
      return getIdFromList(infoList, fileName);
    })
    .then(shareId => {
      return shadowId(shareId).then(secretId => {
        if (shareId === secretId) {
          throw new Error("shadowId: id not changed");
        } else {
          return secretId;
        }
      });
    })
    .then(secretId => {
      return list().then(infoList => {
        const info = infoList.find(info => {
          return info.Id === secretId;
        });

        if (info.PathLocal.includes(fileName)) {
          console.log("shadowId api: ok", secretId);
          return secretId;
        } else {
          throw new Error("shadowId pai: file not found", infoList, fileName);
        }
      });
    })
    .then(secretId => {
      return publishId(secretId).then(publicId => {
        if (publicId === secretId) {
          // TODO: it is not enough to check they are not equal
          throw new Error("publicId: id not changed");
        } else {
          console.log("publishId api: ok", publicId);
          return publicId;
        }
      });
    })
    .then(shareId => del(shareId))
    .then(testLogout)
    .catch(err => {
      console.error(err);
      delWithName(fileName);
    });
}

export function testSetDownLimit() {
  const downLimit = 777;

  return testLogin()
    .then(() => upload(fileName))
    .then(list)
    .then(infoList => {
      return getIdFromList(infoList, fileName);
    })
    .then(shareId => {
      return setDownLimit(shareId, downLimit).then(ok => {
        if (!ok) {
          throw new Error("setDownLimit: failed");
        } else {
          return shareId;
        }
      });
    })
    .then(shareId => {
      return list().then(infoList => {
        const info = infoList.find(info => {
          return info.Id == shareId;
        });

        if (info.DownLimit === downLimit) {
          console.log("setDownLimit api: ok");
          return shareId;
        } else {
          throw new Error("setDownLimit api: limit unchanged");
        }
      });
    })
    .then(shareId => del(shareId))
    .then(testLogout)
    .catch(err => {
      console.error(err);
      delWithName(fileName);
    });
}

// TODO: need to add local file and test
export function testAddLocalFiles() {
  return testLogin()
    .then(() => addLocalFiles())
    .then(ok => {
      if (ok) {
        console.log("addLocalFiles api: ok");
      } else {
        throw new Error("addLocalFiles api: failed");
      }
    })
    .then(() => testLogout())
    .catch(err => {
      console.error(err);
    });
}
