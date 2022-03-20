import { List, Set, Map } from "immutable";

import { initMockWorker, makePromise } from "../../test/helpers";
import { TopBar } from "../topbar";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import {
  NewMockUsersClient,
  resps as usersResps,
} from "../../client/users_mock";
import {
  NewMockFilesClient,
  resps as filesResps,
} from "../../client/files_mock";
import { NewMockSettingsClient } from "../../client/settings_mock";

import { UploadInfo, visitorID, roleVisitor, MetadataResp } from "../../client";
import { UploadEntry, UploadState } from "../../worker/interface";

describe("TopBar", () => {
  initMockWorker();
  // stub confirm
  window.confirm = (message?: string): boolean => {
    return true;
  };

  test("logout as visitor without sharing", async () => {
    const coreState = newState();

    const isSharingMockResp = { status: 404, statusText: "", data: {} };
    const listSharingsMockResp = {
      status: 403,
      statusText: "",
      data: {
        sharingDirs: new Array<string>(),
      },
    };
    const listUploadingsMockResp = {
      status: 403,
      statusText: "",
      data: { uploadInfos: new Array<UploadInfo>() },
    };
    const listHomeMockResp = {
      status: 403,
      statusText: "",
      data: { cwd: "", metadatas: new Array<MetadataResp>() },
    };
    const mockFileResps = {
      ...filesResps,
      listHomeMockResp,
      isSharingMockResp,
      listSharingsMockResp,
      listUploadingsMockResp,
    };

    const selfMockResp = {
      status: 403,
      statusText: "",
      data: {
        id: visitorID,
        name: "visitor",
        role: roleVisitor,
        usedSpace: "0",
        quota: {
          spaceLimit: "0",
          uploadSpeedLimit: 0,
          downloadSpeedLimit: 0,
        },
      },
    };
    const isAuthedMockResp = { status: 403, statusText: "", data: {} };
    const mockUserResps = {
      ...usersResps,
      selfMockResp,
      isAuthedMockResp,
    };

    const filesCl = NewMockFilesClient("");
    filesCl.listHome = jest
      .fn()
      .mockReturnValueOnce(makePromise(listHomeMockResp));
    filesCl.isSharing = jest
      .fn()
      .mockReturnValueOnce(makePromise(isSharingMockResp));
    filesCl.listSharings = jest
      .fn()
      .mockReturnValueOnce(makePromise(listSharingsMockResp));
    filesCl.listUploadings = jest
      .fn()
      .mockReturnValueOnce(makePromise(listUploadingsMockResp));

    const usersCl = NewMockUsersClient("");
    usersCl.self = jest.fn().mockReturnValueOnce(makePromise(selfMockResp));
    usersCl.isAuthed = jest
      .fn()
      .mockReturnValueOnce(makePromise(isAuthedMockResp));

    const settingsCl = NewMockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const topbar = new TopBar({
      ui: coreState.ui,
      login: coreState.login,
      msg: coreState.msg,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    await topbar.logout();

    // files, uploadings, sharings
    expect(coreState.filesInfo.dirPath.join("/")).toEqual(
      mockFileResps.listHomeMockResp.data.cwd
    );
    expect(coreState.filesInfo.isSharing).toEqual(false);
    expect(coreState.filesInfo.items).toEqual(List());
    expect(coreState.sharingsInfo.sharings).toEqual(Map<string, string>());
    expect(coreState.uploadingsInfo.uploadings).toEqual(List<UploadEntry>());

    // login
    expect(coreState.login).toEqual({
      userID: visitorID,
      userName: "visitor",
      userRole: roleVisitor,
      extInfo: {
        usedSpace: "0",
      },
      quota: {
        spaceLimit: "0",
        uploadSpeedLimit: 0,
        downloadSpeedLimit: 0,
      },
      authed: false,
      captchaID: "mockCaptchaID",
      preferences: {
        bg: {
          url: "",
          repeat: "",
          position: "",
          align: "",
          bgColor: "",
        },
        cssURL: "",
        lanPackURL: "",
        lan: "en_US",
        theme: "light",
        avatar: "",
        email: "",
      },
    });

    // admin
    let usersMap = Map({});
    let roles = Set<string>();
    expect(coreState.admin).toEqual({
      users: usersMap,
      roles: roles,
    });
  });
});
