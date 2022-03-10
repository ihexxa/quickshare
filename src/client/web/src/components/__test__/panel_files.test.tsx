import { List } from "immutable";
import * as immutable from "immutable";

import { initMockWorker } from "../../test/helpers";
import { FilesPanel } from "../panel_files";
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
import { makePromise } from "../../test/helpers";

describe("FilesPanel", () => {
  const initFilesPanel = (): any => {
    initMockWorker();

    const coreState = newState();
    const usersCl = NewMockUsersClient("");
    const filesCl = NewMockFilesClient("");
    const settingsCl = NewMockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const filesPanel = new FilesPanel({
      filesInfo: coreState.filesInfo,
      msg: coreState.msg,
      login: coreState.login,
      ui: coreState.ui,
      enabled: true,
      update: (updater: (prevState: ICoreState) => ICoreState) => { },
    });

    return {
      filesPanel,
      usersCl,
      filesCl,
    };
  };

  test("chdir", async () => {
    const { filesPanel, usersCl, filesCl } = initFilesPanel();

    const newCwd = List(["newPos", "subFolder"]);

    await filesPanel.chdir(newCwd);

    expect(updater().props.filesInfo.dirPath).toEqual(newCwd);
    expect(updater().props.filesInfo.isSharing).toEqual(true);
    expect(updater().props.filesInfo.items).toEqual(
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
  });

  test("addSharing", async () => {
    const { filesPanel, usersCl, filesCl } = initFilesPanel();

    const newSharingPath = List(["newPos", "subFolder"]);
    const sharingDir = newSharingPath.join("/");
    const newSharings = immutable.Map<string, string>({
      [sharingDir]: "f123456",
    });
    const newSharingsResp = new Map<string, string>();
    newSharingsResp.set(sharingDir, "f123456");

    filesCl.listSharingIDs = jest.fn().mockReturnValueOnce(
      makePromise({
        status: 200,
        statusText: "",
        data: {
          IDs: newSharingsResp,
        },
      })
    );

    await filesPanel.addSharing(newSharingPath);

    expect(updater().props.filesInfo.isSharing).toEqual(true);
    expect(updater().props.sharingsInfo.sharings).toEqual(newSharings);
  });
});
