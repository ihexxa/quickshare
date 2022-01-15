import { List, Set, Map } from "immutable";

import { initMockWorker } from "../../test/helpers";
import { User, UploadInfo } from "../../client";
import { AuthPane } from "../pane_login";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { UploadState, UploadEntry } from "../../worker/interface";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockSettingsClient } from "../../client/settings_mock";
import { controlName as panelTabs } from "../root_frame";
import {} from "../layers";
import {} from "../dialog_settings";
import {} from "../panel_files";
import {
  sharingCtrl,
  ctrlOn,
  ctrlOff,
  settingsDialogCtrl,
  settingsTabsCtrl,
  filesViewCtrl,
} from "../../common/controls";

describe("Login", () => {
  initMockWorker();

  test("login as admin without sharing", async () => {
    const coreState = newState();
    const pane = new AuthPane({
      login: coreState.login,
      msg: coreState.msg,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    const settingsCl = new MockSettingsClient("");
    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    await pane.login();

    // TODO: state is not checked

    // files, uploadings, sharings
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
              state: UploadState.Ready,
              err: "",
            };
          }
        )
      )
    );

    // login
    expect(updater().props.login).toEqual({
      userID: "0",
      userName: "mockUser",
      userRole: "admin",
      authed: true,
      usedSpace: "256",
      quota: {
        spaceLimit: "7",
        uploadSpeedLimit: 3,
        downloadSpeedLimit: 3,
      },
      captchaID: "mockCaptchaID",
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

    // ui
    expect(coreState.ui).toEqual({
      isVertical: false,
      siteName: "",
      siteDesc: "",
      bg: {
        url: "clientCfg_bg_url",
        repeat: "clientCfg_bg_repeat",
        position: "clientCfg_bg_position",
        align: "clientCfg_bg_align",
      },
      control: {
        controls: Map<string, string>({
          [panelTabs]: "filesPanel",
          [settingsDialogCtrl]: ctrlOff,
          [settingsTabsCtrl]: "preferencePane",
          [sharingCtrl]: ctrlOff,
          [filesViewCtrl]: "rows",
        }),
        options: Map<string, Set<string>>({
          [panelTabs]: Set<string>([
            "filesPanel",
            "uploadingsPanel",
            "sharingsPanel",
          ]),
          [settingsDialogCtrl]: Set<string>([ctrlOn, ctrlOff]),
          [settingsTabsCtrl]: Set<string>(["preferencePane", "managementPane"]),
          [sharingCtrl]: Set<string>([ctrlOn, ctrlOff]),
          [filesViewCtrl]: Set<string>(["rows", "table"]),
        }),
      },
    });
  });
});
