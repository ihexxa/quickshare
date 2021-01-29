import { List, Map } from "immutable";
import { mock, instance, anyString, when } from "ts-mockito";

import { ICoreState, initWithWorker, mockState } from "../core_state";
import {
  makePromise,
  makeNumberResponse,
  mockUpdate,
} from "../../test/helpers";
import { Updater, Browser } from "../browser";
import { MockUsersClient } from "../../client/users_mock";
import { UsersClient } from "../../client/users";
import { FilesClient } from "../../client/files";
import { FilesClient as MockFilesClient } from "../../client/files_mock";
import { MetadataResp, UploadInfo } from "../../client";
import { MockWorker, UploadEntry } from "../../worker/interface";
import { UploadMgr } from "../../worker/upload_mgr";

describe("Browser", () => {
  const mockWorkerClass = mock(MockWorker);
  const mockWorker = instance(mockWorkerClass);

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
    const filesClient = new MockFilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      filesClient.listMock(makePromise(tc.listResp));
      Updater.setClients(usersClient, filesClient);

      const coreState = initWithWorker(mockWorker);
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
    const filesClient = new MockFilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      filesClient.listMock(makePromise(tc.listResp));
      filesClient.deleteMock(makeNumberResponse(200));
      Updater.setClients(usersClient, filesClient);

      const coreState = initWithWorker(mockWorker);
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
    const filesClient = new MockFilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      filesClient.listMock(makePromise(tc.listResp));
      filesClient.moveMock(makeNumberResponse(200));
      Updater.setClients(usersClient, filesClient);

      const coreState = initWithWorker(mockWorker);
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

  xtest("Browser: deleteUploading", async () => {
    interface TestCase {
      deleteFile: string;
      preState: ICoreState;
      postState: ICoreState;
    }

    const tcs: any = [
      {
        deleteFile: "./path/file",
        preState: {
          browser: {
            uploadings: List<UploadInfo>([
              {
                realFilePath: "./path/file",
                size: 1,
                uploaded: 0,
              },
            ]),
            update: mockUpdate,
          },
        },
        postState: {
          browser: {
            uploadings: List<UploadInfo>(),
            update: mockUpdate,
          },
        },
      },
    ];

    const setState = (patch: any, state: ICoreState): ICoreState => {
      state.panel.browser = patch.browser;
      return state;
    };
    const mockFilesClientClass = mock(FilesClient);
    when(mockFilesClientClass.deleteUploading(anyString())).thenResolve({
      status: 200,
      statusText: "",
      data: "",
    });
    // TODO: the return should dpends on test case
    when(mockFilesClientClass.listUploadings()).thenResolve({
      status: 200,
      statusText: "",
      data: { uploadInfos: Array<UploadInfo>() },
    });

    const mockUsersClientClass = mock(UsersClient);

    const mockFilesClient = instance(mockFilesClientClass);
    const mockUsersClient = instance(mockUsersClientClass);
    tcs.forEach((tc: TestCase) => {
      const preState = setState(tc.preState, mockState());
      const postState = setState(tc.postState, mockState());
      // const existingFileName = preState.panel.browser.uploadings.get(0).realFilePath;
      const infos:Map<string, UploadEntry> = Map();
      UploadMgr._setInfos(infos);

      const component = new Browser(preState.panel.browser);
      Updater.init(preState.panel.browser);
      Updater.setClients(mockUsersClient, mockFilesClient);

      component.deleteUploading(tc.deleteFile);
      expect(Updater.props).toEqual(postState.panel.browser);
    });
  });
});
