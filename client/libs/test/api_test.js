import { testAuth } from "./api_auth_test";
import { testUploadOneFile } from "./api_upload_test";
import {
  testAddLocalFiles,
  testSetDownLimit,
  testShadowPublishId
} from "./api_share_test";
import { testUpDownBatch } from "./api_up_down_batch_test";

console.log("Test started");

const fileName = `test_filename${Date.now()}`;
const file = new File(["foo"], fileName, {
  type: "text/plain"
});

testAuth()
  .then(testShadowPublishId)
  .then(() => testUploadOneFile(file, fileName))
  .then(testSetDownLimit)
  .then(testAddLocalFiles)
  .then(testUpDownBatch)
  .then(() => {
    console.log("Tests are finished");
  });
