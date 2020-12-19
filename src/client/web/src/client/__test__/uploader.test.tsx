import * as React from "react";

import { FileUploader } from "../uploader";
import { FilesClient } from "../files_mock";

describe("Updater", () => {
  const content = ["123456"];
  const filePath = "mock/file";
  const blob = new Blob(content);
  const fileSize = blob.size;
  const file = new File(content, filePath);

  const makePromise = (ret: any): Promise<any> => {
    return new Promise<any>((resolve) => {
      resolve(ret);
    });
  };
  const makeCreateResp = (status: number): Promise<any> => {
    return makePromise({
      status: status,
      statusText: "",
      data: {
        path: filePath,
        fileSize: fileSize,
      },
    });
  };
  const makeStatusResp = (status: number, uploaded: number): Promise<any> => {
    return makePromise({
      status: status,
      statusText: "",
      data: {
        path: filePath,
        isDir: false,
        fileSize: fileSize,
        uploaded: uploaded,
      },
    });
  };

  interface statusResp {
    status: number;
    uploaded: number;
  }
  interface TestCase {
    createResps: Array<number>;
    uploadChunkResps: Array<any>;
    uploadStatusResps: Array<any>;
    result: boolean;
  }
  test("Updater: start ok", async () => {
    const testCases: Array<TestCase> = [
      {
        // fail to create file 4 times
        createResps: [500, 500, 500, 500],
        uploadChunkResps: [],
        uploadStatusResps: [],
        result: false,
      },
      {
        // fail to get status
        createResps: [200],
        uploadChunkResps: [{ status: 500, uploaded: 0 }],
        uploadStatusResps: [{ status: 600, uploaded: 0 }],
        result: false,
      },
      {
        // upload ok
        createResps: [200],
        uploadChunkResps: [
          { status: 200, uploaded: 0 },
          { status: 200, uploaded: 1 },
          { status: 200, uploaded: fileSize },
        ],
        uploadStatusResps: [],
        result: true,
      },
      {
        // fail once
        createResps: [200],
        uploadChunkResps: [
          { status: 200, uploaded: 0 },
          { status: 500, uploaded: 1 },
          { status: 200, uploaded: fileSize },
        ],
        uploadStatusResps: [{ status: 200, uploaded: 1 }],
        result: true,
      },
      {
        // fail twice
        createResps: [500, 500, 500, 200],
        uploadChunkResps: [
          { status: 200, uploaded: 0 },
          { status: 500, uploaded: 1 },
          { status: 500, uploaded: 1 },
          { status: 200, uploaded: fileSize },
        ],
        uploadStatusResps: [
          { status: 500, uploaded: 1 },
          { status: 500, uploaded: 1 },
        ],
        result: true,
      },
      {
        // other errors
        createResps: [500, 200],
        uploadChunkResps: [
          { status: 601, uploaded: 0 },
          { status: 408, uploaded: fileSize },
          { status: 200, uploaded: 1 },
          { status: 200, uploaded: fileSize },
        ],
        uploadStatusResps: [
          { status: 500, uploaded: 1 },
          { status: 500, uploaded: 1 },
        ],
        result: true,
      },
    ];

    for (let i = 0; i < testCases.length; i++) {
      const tc = testCases[i];
      const uploader = new FileUploader(file, filePath);
      const mockClient = new FilesClient("");

      const createResps = tc.createResps.map((resp) => makeCreateResp(resp));
      mockClient.createMock(createResps);
      const uploadChunkResps = tc.uploadChunkResps.map((resp) =>
        makeStatusResp(resp.status, resp.uploaded)
      );
      mockClient.uploadChunkMock(uploadChunkResps);
      const uploadStatusResps = tc.uploadStatusResps.map((resp) =>
        makeStatusResp(resp.status, resp.uploaded)
      );
      mockClient.uploadStatusMock(uploadStatusResps);
      uploader.setClient(mockClient);

      const ret = await uploader.start();
      expect(ret).toEqual(tc.result);
    }
  });
});
