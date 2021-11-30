import { List, Set, Map } from "immutable";

import { initMockWorker } from "../../test/helpers";
import { TopBar } from "../topbar";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockSettingsClient } from "../../client/settings_mock";

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
      status: 401,
      statusText: "",
      data: {
        sharingDirs: new Array<string>(),
      },
    };
    const listUploadingsMockResp = {
      status: 401,
      statusText: "",
      data: { uploadInfos: new Array<UploadInfo>() },
    };
    const listHomeMockResp = {
      status: 401,
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
      status: 401,
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
    const isAuthedMockResp = { status: 401, statusText: "", data: {} };
    const mockUserResps = {
      ...usersResps,
      selfMockResp,
      isAuthedMockResp,
    };

    const filesCl = new MockFilesClient("");
    filesCl.setMock(mockFileResps);
    const usersCl = new MockUsersClient("");
    usersCl.setMock(mockUserResps);
    const settingsCl = new MockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const topbar = new TopBar({
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
    expect(coreState.sharingsInfo.sharings).toEqual(List());
    expect(coreState.uploadingsInfo.uploadings).toEqual(List<UploadEntry>());

    // login
    expect(coreState.login).toEqual({
      userID: visitorID,
      userName: "visitor",
      userRole: roleVisitor,
      usedSpace: "0",
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
        },
        cssURL: "",
        lanPackURL: "",
        lan: "en_US",
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
