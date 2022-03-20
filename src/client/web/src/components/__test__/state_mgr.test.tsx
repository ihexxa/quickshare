import { List, Set, Map } from "immutable";

import { initMockWorker, makePromise } from "../../test/helpers";
import { StateMgr } from "../state_mgr";
import { User, UploadInfo, visitorID, roleVisitor } from "../../client";
import {
  NewMockFilesClient,
  resps as filesResps,
} from "../../client/files_mock";
import { shareDirQuery } from "../../client/files";
import {
  NewMockUsersClient,
  resps as usersResps,
} from "../../client/users_mock";
import {
  NewMockSettingsClient,
  resps as settingsResps,
} from "../../client/settings_mock";
import { ICoreState, newState } from "../core_state";
import { UploadState, UploadEntry } from "../../worker/interface";
import { MsgPackage } from "../../i18n/msger";

describe("State Manager", () => {
  initMockWorker();
  const emptyQuery = new URLSearchParams("");
  // stub alert
  window.alert = (message?: string): void => {
    console.log(message);
  };

  test("initUpdater for admin", async () => {
    const usersCl = NewMockUsersClient("");
    const filesCl = NewMockFilesClient("");
    const settingsCl = NewMockSettingsClient("");

    const mgr = new StateMgr({}); // it will call initUpdater
    mgr.setUsersClient(usersCl);
    mgr.setFilesClient(filesCl);
    mgr.setSettingsClient(settingsCl);

    // TODO: depress warning
    mgr.update = (apply: (prevState: ICoreState) => ICoreState): void => {
      // no op
    };

    const coreState = newState();
    await mgr.initUpdater(coreState, emptyQuery);

    // browser
    expect(coreState.filesInfo.dirPath.join("/")).toEqual("mock_home/files");
    expect(coreState.filesInfo.isSharing).toEqual(true);
    expect(coreState.sharingsInfo.sharings).toEqual(
      Map(filesResps.listSharingIDsMockResp.data.IDs.entries())
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
              state: UploadState.Stopped,
              err: "",
            };
          }
        )
      )
    );

    expect(coreState.filesInfo.items).toEqual(
      List([
        {
          name: "mock_dir",
          size: 0,
          modTime: "0",
          isDir: true,
          sha1: "",
        },
        {
          name: "mock_file",
          size: 5,
          modTime: "0",
          isDir: false,
          sha1: "mock_file_sha1",
        },
      ])
    );

    // login
    expect(coreState.login).toEqual({
      userID: "0",
      userName: "mockUser",
      userRole: "admin",
      extInfo: {
        usedSpace: "256",
      },
      quota: {
        spaceLimit: "7",
        uploadSpeedLimit: 3,
        downloadSpeedLimit: 3,
      },
      authed: true,
      captchaID: "mockCaptchaID",
      preferences: {
        bg: {
          url: "bgUrl",
          repeat: "bgRepeat",
          position: "bgPosition",
          align: "bgAlign",
          bgColor: "bgColor",
        },
        cssURL: "cssURL",
        lanPackURL: "lanPackURL",
        lan: "en_US",
        theme: "light",
        avatar: "avatar",
        email: "email",
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
    expect(coreState.ui.bg).toEqual(settingsResps.getClientCfgMockResp.data.bg);
  });

  test("initUpdater for visitor in sharing mode", async () => {
    const usersCl = NewMockUsersClient("");
    const filesCl = NewMockFilesClient("");
    const settingsCl = NewMockSettingsClient("");

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
    usersCl.isAuthed = jest
      .fn()
      .mockReturnValue(makePromise({ status: 403, statusText: "", data: {} }));
    usersCl.self = jest.fn().mockReturnValue(makePromise(mockSelfResp));

    const coreState = newState();

    const mgr = new StateMgr({}); // it will call initUpdater
    mgr.setUsersClient(usersCl);
    mgr.setFilesClient(filesCl);
    mgr.setSettingsClient(settingsCl);
    // TODO: depress warning
    mgr.update = (apply: (prevState: ICoreState) => ICoreState): void => {
      // no op
    };

    const sharingPath = "sharingPath/files";
    const query = new URLSearchParams(`?${shareDirQuery}=${sharingPath}`);

    await mgr.initUpdater(coreState, query);

    // browser
    expect(coreState.filesInfo.dirPath.join("/")).toEqual(sharingPath);
    expect(coreState.filesInfo.isSharing).toEqual(true);
    expect(coreState.filesInfo.items).toEqual(
      List([
        {
          name: "mock_dir",
          size: 0,
          modTime: "0",
          isDir: true,
          sha1: "",
        },
        {
          name: "mock_file",
          size: 5,
          modTime: "0",
          isDir: false,
          sha1: "mock_file_sha1",
        },
      ])
    );
    expect(coreState.sharingsInfo.sharings).toEqual(Map<string, string>());
    expect(coreState.uploadingsInfo.uploadings).toEqual(List<UploadEntry>([]));

    // login
    expect(coreState.login).toEqual({
      userID: mockSelfResp.data.id,
      userName: mockSelfResp.data.name,
      userRole: mockSelfResp.data.role,
      quota: mockSelfResp.data.quota,
      extInfo: {
        usedSpace: mockSelfResp.data.usedSpace,
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
    expect(coreState.admin).toEqual({
      users: Map({}),
      roles: Set<string>(),
    });

    // msg
    // it is fallback to en_US because language pack url is not valid
    expect(coreState.msg.lan).toEqual("en_US");
    expect(coreState.msg.pkg).toEqual(MsgPackage.get("en_US"));

    // ui
    expect(coreState.ui.bg).toEqual(settingsResps.getClientCfgMockResp.data.bg);
  });

  test("initUpdater for visitor", async () => {
    const usersCl = NewMockUsersClient("");
    const filesCl = NewMockFilesClient("");
    const settingsCl = NewMockSettingsClient("");

    usersCl.isAuthed = jest.fn().mockReturnValue(
      makePromise({
        status: 403,
        statusText: "",
        data: { error: "unauthorized" },
      })
    );
    usersCl.self = jest.fn().mockReturnValue(
      makePromise({
        status: 403,
        statusText: "",
        data: { error: "malformed token" },
      })
    );
    usersCl.getCaptchaID = jest.fn().mockReturnValue(
      makePromise({
        status: 200,
        statusText: "",
        data: { id: "357124" },
      })
    );

    const coreState = newState();

    const mgr = new StateMgr({}); // it will call initUpdater
    mgr.setUsersClient(usersCl);
    mgr.setFilesClient(filesCl);
    mgr.setSettingsClient(settingsCl);
    // TODO: depress warning
    mgr.update = (apply: (prevState: ICoreState) => ICoreState): void => {
      // no op
    };

    const query = new URLSearchParams("");
    await mgr.initUpdater(coreState, query);

    // browser
    expect(coreState.filesInfo.dirPath.join("/")).toEqual("");
    expect(coreState.filesInfo.isSharing).toEqual(false);
    expect(coreState.filesInfo.items).toEqual(List());
    expect(coreState.sharingsInfo.sharings).toEqual(Map<string, string>());
    expect(coreState.uploadingsInfo.uploadings).toEqual(List<UploadEntry>([]));

    // // login
    expect(coreState.login).toEqual({
      userID: visitorID,
      userName: "visitor",
      userRole: roleVisitor,
      quota: {
        uploadSpeedLimit: 0,
        downloadSpeedLimit: 0,
        spaceLimit: "0",
      },
      extInfo: {
        usedSpace: "0",
      },
      authed: false,
      captchaID: "357124",
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
    expect(coreState.admin).toEqual({
      users: Map({}),
      roles: Set<string>(),
    });

    // msg
    // it is fallback to en_US because language pack url is not valid
    expect(coreState.msg.lan).toEqual("en_US");
    expect(coreState.msg.pkg).toEqual(MsgPackage.get("en_US"));

    // ui
    expect(coreState.ui.bg).toEqual({
      align: "clientCfg_bg_align",
      position: "clientCfg_bg_position",
      repeat: "clientCfg_bg_repeat",
      url: "clientCfg_bg_url",
      bgColor: "clientCfg_bg_bg_Color",
    });
  });
});
