import { List, Map } from "immutable";

import { init } from "../core_state";
import { makePromise, makeNumberResponse } from "../../test/helpers";
import { Updater } from "../browser";
import { MockUsersClient } from "../../client/users_mock";
import { FilesClient } from "../../client/files_mock";
import { MetadataResp } from "../../client";

describe("Browser", () => {
  test("Updater: setPwd", async () => {
    const tests = [
      {
        listResp: {
          status: 200,
          statusText: "",
          data: {
            metadatas: [
              {
                name: "file",
                size: 1,
                modTime: "1-1",
                isDir: false,
              },
              {
                name: "folder",
                size: 0,
                modTime: "1-1",
                isDir: true,
              },
            ],
          },
        },
        filePath: "path/file",
      },
    ];

    const usersClient = new MockUsersClient("");
    const filesClient = new FilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      filesClient.listMock(makePromise(tc.listResp));
      Updater.setClients(usersClient, filesClient);

      const coreState = init();
      Updater.init(coreState.panel.browser);
      await Updater.setItems(List<string>(tc.filePath.split("/")));
      const newState = Updater.setBrowser(coreState);

      newState.panel.browser.items.forEach((item, i) => {
        expect(item.name).toEqual(tc.listResp.data.metadatas[i].name);
        expect(item.size).toEqual(tc.listResp.data.metadatas[i].size);
        expect(item.modTime).toEqual(tc.listResp.data.metadatas[i].modTime);
        expect(item.isDir).toEqual(tc.listResp.data.metadatas[i].isDir);
      });
    }
  });

  test("Updater: delete", async () => {
    const tests = [
      {
        dirPath: "path/path2",
        items: [
          {
            name: "file",
            size: 1,
            modTime: "1-1",
            isDir: false,
          },
          {
            name: "folder",
            size: 0,
            modTime: "1-1",
            isDir: true,
          },
        ],
        selected: {
          file: true,
        },
        listResp: {
          status: 200,
          statusText: "",
          data: {
            metadatas: [
              {
                name: "folder",
                size: 0,
                modTime: "1-1",
                isDir: true,
              },
            ],
          },
        },
        filePath: "path/file",
      },
    ];

    const usersClient = new MockUsersClient("");
    const filesClient = new FilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      filesClient.listMock(makePromise(tc.listResp));
      filesClient.deleteMock(makeNumberResponse(200));
      Updater.setClients(usersClient, filesClient);

      const coreState = init();
      Updater.init(coreState.panel.browser);
      await Updater.delete(
        List<string>(tc.dirPath.split("/")),
        List<MetadataResp>(tc.items),
        Map<boolean>(tc.selected)
      );
      const newState = Updater.setBrowser(coreState);

      // TODO: check inputs of delete

      newState.panel.browser.items.forEach((item, i) => {
        expect(item.name).toEqual(tc.listResp.data.metadatas[i].name);
        expect(item.size).toEqual(tc.listResp.data.metadatas[i].size);
        expect(item.modTime).toEqual(tc.listResp.data.metadatas[i].modTime);
        expect(item.isDir).toEqual(tc.listResp.data.metadatas[i].isDir);
      });
    }
  });

  test("Updater: moveHere", async () => {
    const tests = [
      {
        dirPath1: "path/path1",
        dirPath2: "path/path2",
        selected: {
          file1: true,
          file2: true,
        },
        listResp: {
          status: 200,
          statusText: "",
          data: {
            metadatas: [
              {
                name: "file1",
                size: 1,
                modTime: "1-1",
                isDir: false,
              },
              {
                name: "file2",
                size: 2,
                modTime: "1-1",
                isDir: false,
              },
            ],
          },
        },
      },
    ];

    const usersClient = new MockUsersClient("");
    const filesClient = new FilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      filesClient.listMock(makePromise(tc.listResp));
      filesClient.moveMock(makeNumberResponse(200));
      Updater.setClients(usersClient, filesClient);

      const coreState = init();
      Updater.init(coreState.panel.browser);
      await Updater.moveHere(
        tc.dirPath1,
        tc.dirPath2,
        Map<boolean>(tc.selected)
      );

      // TODO: check inputs of move

      const newState = Updater.setBrowser(coreState);
      newState.panel.browser.items.forEach((item, i) => {
        expect(item.name).toEqual(tc.listResp.data.metadatas[i].name);
        expect(item.size).toEqual(tc.listResp.data.metadatas[i].size);
        expect(item.modTime).toEqual(tc.listResp.data.metadatas[i].modTime);
        expect(item.isDir).toEqual(tc.listResp.data.metadatas[i].isDir);
      });
    }
  });
});
