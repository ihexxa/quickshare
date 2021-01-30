import * as React from "react";
import { List, Map } from "immutable";
import { mock, instance, anyString, anything, when, verify } from "ts-mockito";

import { ICoreState, initWithWorker, mockState } from "../core_state";
import {
  makePromise,
  makeNumberResponse,
  mockUpdate,
  addMockUpdate,
  mockFileList,
} from "../../test/helpers";
import { Browser } from "../browser";
import { Updater, setUpdater } from "../browser.updater";
import { MockUsersClient } from "../../client/users_mock";
import { UsersClient } from "../../client/users";
import { FilesClient } from "../../client/files";
import { FilesClient as MockFilesClient } from "../../client/files_mock";
import { MetadataResp, UploadInfo } from "../../client";
import { MockWorker, UploadEntry } from "../../worker/interface";
import { UploadMgr, setUploadMgr } from "../../worker/upload_mgr";

describe("Browser", () => {
  const mockWorkerClass = mock(MockWorker);
  const mockWorker = instance(mockWorkerClass);

  test("Updater: addUploads: add each files to UploadMgr", async () => {
    let coreState = mockState();
    const UploadMgrClass = mock(UploadMgr);
    const uploadMgr = instance(UploadMgrClass);
    setUploadMgr(uploadMgr);

    const filePaths = ["./file1", "./file2"];
    const fileList = mockFileList(filePaths);
    const updater = new Updater();
    updater.setUploadings = (infos: Map<string, UploadEntry>) => {};
    updater.init(coreState.panel.browser);

    updater.addUploads(fileList);

    // it seems that new File will do some file path escaping, so just check call time here
    verify(UploadMgrClass.add(anything(), anything())).times(filePaths.length);
    // filePaths.forEach((filePath, i) => {
    //   verify(UploadMgrClass.add(anything(), filePath)).once();
    // });
  });

  test("Updater: deleteUploads: call UploadMgr and api to delete", async () => {
    let coreState = mockState();
    const UploadMgrClass = mock(UploadMgr);
    const uploadMgr = instance(UploadMgrClass);
    setUploadMgr(uploadMgr);

    const updater = new Updater();
    const filesClientClass = mock(FilesClient);
    when(filesClientClass.deleteUploading(anyString())).thenResolve({
      status: 200,
      statusText: "",
      data: "",
    });
    const filesClient = instance(filesClientClass);
    const usersClientClass = mock(UsersClient);
    const usersClient = instance(usersClientClass);
    updater.init(coreState.panel.browser);
    updater.setClients(usersClient, filesClient);

    const filePath = "./path/file";
    updater.deleteUpload(filePath);

    verify(filesClientClass.deleteUploading(filePath)).once();
    verify(UploadMgrClass.delete(filePath)).once();
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
      const updater = new Updater();
      updater.setClients(usersClient, filesClient);
      filesClient.listMock(makePromise(tc.listResp));
      filesClient.deleteMock(makeNumberResponse(200));
      const coreState = initWithWorker(mockWorker);
      updater.init(coreState.panel.browser);

      await updater.delete(
        List<string>(tc.dirPath.split("/")),
        List<MetadataResp>(tc.items),
        Map(tc.selected)
      );

      const newState = updater.setBrowser(coreState);

      // TODO: check inputs of delete
      newState.panel.browser.items.forEach((item, i) => {
        expect(item.name).toEqual(tc.listResp.data.metadatas[i].name);
        expect(item.size).toEqual(tc.listResp.data.metadatas[i].size);
        expect(item.modTime).toEqual(tc.listResp.data.metadatas[i].modTime);
        expect(item.isDir).toEqual(tc.listResp.data.metadatas[i].isDir);
      });
    }
  });

  test("Updater: setItems", async () => {
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
      const updater = new Updater();
      filesClient.listMock(makePromise(tc.listResp));
      updater.setClients(usersClient, filesClient);
      const coreState = initWithWorker(mockWorker);
      updater.init(coreState.panel.browser);

      await updater.setItems(List<string>(tc.filePath.split("/")));
      const newState = updater.setBrowser(coreState);

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
      const updater = new Updater();

      filesClient.listMock(makePromise(tc.listResp));
      filesClient.moveMock(makeNumberResponse(200));
      updater.setClients(usersClient, filesClient);

      const coreState = initWithWorker(mockWorker);
      updater.init(coreState.panel.browser);
      await updater.moveHere(tc.dirPath1, tc.dirPath2, Map(tc.selected));

      const newState = updater.setBrowser(coreState);

      // TODO: check inputs of move
      newState.panel.browser.items.forEach((item, i) => {
        expect(item.name).toEqual(tc.listResp.data.metadatas[i].name);
        expect(item.size).toEqual(tc.listResp.data.metadatas[i].size);
        expect(item.modTime).toEqual(tc.listResp.data.metadatas[i].modTime);
        expect(item.isDir).toEqual(tc.listResp.data.metadatas[i].isDir);
      });
    }
  });

  test("Browser: deleteUpload: tell uploader to deleteUpload and refreshUploadings", async () => {
    let coreState = mockState();
    addMockUpdate(coreState.panel.browser);
    const component = new Browser(coreState.panel.browser);
    const UpdaterClass = mock(Updater);
    const mockUpdater = instance(UpdaterClass);
    setUpdater(mockUpdater);
    when(UpdaterClass.setItems(anything())).thenResolve();
    when(UpdaterClass.deleteUpload(anyString())).thenResolve(true);
    when(UpdaterClass.refreshUploadings()).thenResolve(true);

    const filePath = "filePath";
    await component.deleteUpload(filePath);

    verify(UpdaterClass.deleteUpload(filePath)).once();
    verify(UpdaterClass.refreshUploadings()).once();
  });

  test("Browser: stopUploading: tell updater to stopUploading", async () => {
    let coreState = mockState();
    addMockUpdate(coreState.panel.browser);
    const component = new Browser(coreState.panel.browser);
    const UpdaterClass = mock(Updater);
    const mockUpdater = instance(UpdaterClass);
    setUpdater(mockUpdater);
    when(UpdaterClass.stopUploading(anyString())).thenReturn();

    const filePath = "filePath";
    component.stopUploading(filePath);

    verify(UpdaterClass.stopUploading(filePath)).once();
  });
});
