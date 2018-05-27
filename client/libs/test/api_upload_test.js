import { FileUploader } from "../api_upload";
import { list, del } from "../api_share";
import { testLogin, testLogout } from "./api_auth_test";

function verify(fileName) {
  return list()
    .then(list => {
      if (list == null) {
        throw new Error("upload: list fail");
      }

      // TODO: should verify file name
      const filterInfo = list.find(info => {
        return info.PathLocal.includes(fileName);
      });

      if (filterInfo == null) {
        console.error(list);
        throw new Error("upload: file name not found");
      } else {
        return filterInfo.Id;
      }
    })
    .then(shareId => {
      console.log("upload api: ok");
      del(shareId);
    })
    .then(testLogout)
    .catch(err => {
      throw err;
    });
}

export function testUpload(file) {
  const onStart = () => true;
  const onProgress = () => true;
  const onFinish = () => true;
  const onError = err => {
    throw new Error(JSON.stringify(err));
  };
  const uploader = new FileUploader(onStart, onProgress, onFinish, onError);

  return uploader.uploadFile(file).catch(err => {
    console.error(err);
  });
}

export function testUploadOneFile(file, fileName) {
  const onStart = () => true;
  const onProgress = () => true;
  const onFinish = () => true;
  const onError = err => {
    throw new Error(JSON.stringify(err));
  };
  const uploader = new FileUploader(onStart, onProgress, onFinish, onError);

  return testLogin()
    .then(() => uploader.uploadFile(file))
    .then(() => verify(fileName))
    .catch(err => {
      console.error(err);
    });
}
