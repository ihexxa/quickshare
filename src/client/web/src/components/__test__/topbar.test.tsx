import { List, Set, Map } from "immutable";
import { mock, instance } from "ts-mockito";

import { TopBar } from "../topbar";
import { initUploadMgr } from "../../worker/upload_mgr";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockSettingsClient } from "../../client/settings_mock";

import { UploadInfo, visitorID, roleVisitor, MetadataResp } from "../../client";
import { UploadEntry, UploadState } from "../../worker/interface";



describe("TopBar", () => {
  test("logout as visitor without sharing", async () => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);
    initUploadMgr(mockWorker);

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
      data: { cwd: "", metadatas: new Array<MetadataResp>() }
    };
    const mockFileResps = {
      ...filesResps,
      listHomeMockResp,
      isSharingMockResp,
      listSharingsMockResp,
      listUploadingsMockResp,
    }

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
    }
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
      panes: coreState.panes,
      msg: coreState.msg,
      update: (updater: (prevState: ICoreState) => ICoreState) => { },
    });

    await topbar.logout();

    // browser
    expect(coreState.browser.dirPath.join("/")).toEqual(mockFileResps.listHomeMockResp.data.cwd);
    expect(coreState.browser.isSharing).toEqual(false);
    expect(coreState.browser.sharings).toEqual(List());
    expect(coreState.browser.uploadings).toEqual(List<UploadEntry>());
    expect(coreState.browser.items).toEqual(List());

    // panes
    expect(coreState.panes).toEqual({
      displaying: "login",
      paneNames: Set(["login"]),
    });

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
      }
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
