import { List, Set, Map } from "immutable";
import { mock, instance } from "ts-mockito";

import { initMockWorker } from "../../test/helpers";
import { StateMgr } from "../state_mgr";
import { User, UploadInfo } from "../../client";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import {
  MockSettingsClient,
  resps as settingsResps,
} from "../../client/settings_mock";
import { ICoreState, newState } from "../core_state";
import { UploadState, UploadEntry } from "../../worker/interface";
import { MsgPackage } from "../../i18n/msger";

describe("State Manager", () => {
  initMockWorker();

  test("initUpdater for admin", async () => {
    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    const settingsCl = new MockSettingsClient("");

    const mgr = new StateMgr({}); // it will call initUpdater
    mgr.setUsersClient(usersCl);
    mgr.setFilesClient(filesCl);
    mgr.setSettingsClient(settingsCl);

    // TODO: depress warning
    mgr.update = (apply: (prevState: ICoreState) => ICoreState): void => {
      // no op
    };

    const coreState = newState();
    await mgr.initUpdater(coreState);

    // browser
    expect(coreState.filesInfo.dirPath.join("/")).toEqual("mock_home/files");
    expect(coreState.filesInfo.isSharing).toEqual(true);
    expect(coreState.sharingsInfo.sharings).toEqual(
      List(filesResps.listSharingsMockResp.data.sharingDirs)
    );
    expect(coreState.uploadingsInfo.uploadings).toEqual(
      List<UploadEntry>(
        filesResps.listUploadingsMockResp.data.uploadInfos.map(
          (info: UploadInfo) => {
            return {
              file: undefined,
              filePath: info.realFilePath,
              size: info.size,
              uploaded: info.uploaded,
              state: UploadState.Ready,
              err: "",
            };
          }
        )
      )
    );

    expect(coreState.filesInfo.items).toEqual(
      List(filesResps.listHomeMockResp.data.metadatas)
    );

    // login
    expect(coreState.login).toEqual({
      userID: "0",
      userName: "mockUser",
      userRole: "admin",
      usedSpace: "256",
      quota: {
        spaceLimit: "7",
        uploadSpeedLimit: 3,
        downloadSpeedLimit: 3,
      },
      authed: true,
      captchaID: "",
      preferences: {
        bg: {
          url: "bgUrl",
          repeat: "bgRepeat",
          position: "bgPosition",
          align: "bgAlign",
        },
        cssURL: "cssURL",
        lanPackURL: "lanPackURL",
        lan: "en_US",
      },
    });

    // admin
    let usersMap = Map({});
    usersResps.listUsersMockResp.data.users.forEach((user: User) => {
      usersMap = usersMap.set(user.name, user);
    });
    let roles = Set<string>();
    Object.keys(usersResps.listRolesMockResp.data.roles).forEach(
      (role: string) => {
        roles = roles.add(role);
      }
    );
    expect(coreState.admin).toEqual({
      users: usersMap,
      roles: roles,
    });

    // msg
    // it is fallback to en_US because language pack url is not valid
    expect(coreState.msg.lan).toEqual("en_US");
    expect(coreState.msg.pkg).toEqual(MsgPackage.get("en_US"));

    // ui
    expect(coreState.ui.bg).toEqual(settingsResps.getClientCfgMockResp.data.clientCfg.bg);
  });

  test("initUpdater for visitor in sharing mode", async () => {
    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    const settingsCl = new MockSettingsClient("");
    const mockSelfResp = {
      status: 200,
      statusText: "",
      data: {
        id: "-1",
        name: "visitor",
        role: "visitor",
        usedSpace: "0",
        quota: {
          spaceLimit: "0",
          uploadSpeedLimit: 0,
          downloadSpeedLimit: 0,
        },
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
      },
    };
    const mockIsAuthedResp = { status: 401, statusText: "", data: {} };
    const mockUserResps = {
      ...usersResps,
      isAuthedMockResp: mockIsAuthedResp,
      selfMockResp: mockSelfResp,
    };
    usersCl.setMock(mockUserResps);

    const coreState = newState();

    const mgr = new StateMgr({}); // it will call initUpdater
    mgr.setUsersClient(usersCl);
    mgr.setFilesClient(filesCl);
    mgr.setSettingsClient(settingsCl);
    // TODO: depress warning
    mgr.update = (apply: (prevState: ICoreState) => ICoreState): void => {
      // no op
    };
    await mgr.initUpdater(coreState);

    // browser
    // TODO: mock query to get dir parm
    expect(coreState.filesInfo.dirPath.join("/")).toEqual("mock_home/files");
    expect(coreState.filesInfo.isSharing).toEqual(true);
    expect(coreState.filesInfo.items).toEqual(
      List(filesResps.listHomeMockResp.data.metadatas)
    );
    expect(coreState.sharingsInfo.sharings).toEqual(List([]));
    expect(coreState.uploadingsInfo.uploadings).toEqual(List<UploadEntry>([]));

    // login
    expect(coreState.login).toEqual({
      userID: mockSelfResp.data.id,
      userName: mockSelfResp.data.name,
      userRole: mockSelfResp.data.role,
      quota: mockSelfResp.data.quota,
      usedSpace: mockSelfResp.data.usedSpace,
      authed: false,
      captchaID: "",
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
    expect(coreState.admin).toEqual({
      users: Map({}),
      roles: Set<string>(),
    });

    // msg
    // it is fallback to en_US because language pack url is not valid
    expect(coreState.msg.lan).toEqual("en_US");
    expect(coreState.msg.pkg).toEqual(MsgPackage.get("en_US"));

    // ui
    expect(coreState.ui.bg).toEqual(settingsResps.getClientCfgMockResp.data.clientCfg.bg);
  });
});
