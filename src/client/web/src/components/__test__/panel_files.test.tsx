import { mock, instance, verify, when, anything } from "ts-mockito";
import { List } from "immutable";

import { FilesPanel } from "../panel_files";
import { initUploadMgr } from "../../worker/upload_mgr";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockSettingsClient } from "../../client/settings_mock";

describe("FilesPanel", () => {
  const initFilesPanel = (): any => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);
    initUploadMgr(mockWorker);

    const coreState = newState();
    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    const settingsCl = new MockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const filesPanel = new FilesPanel({
      filesInfo: coreState.filesInfo,
      msg: coreState.msg,
      login: coreState.login,
      ui: coreState.ui,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    return {
      filesPanel,
      usersCl,
      filesCl,
    };
  };

  test("chdir", async () => {
    const { filesPanel, usersCl, filesCl } = initFilesPanel();

    const newSharings = ["mock_sharingfolder1", "mock_sharingfolder2"];
    const newCwd = List(["newPos", "subFolder"]);

    filesCl.setMock({
      ...filesResps,
      listSharingsMockResp: {
        status: 200,
        statusText: "",
        data: {
          sharingDirs: newSharings,
        },
      },
    });

    await filesPanel.chdir(newCwd);

    expect(updater().props.filesInfo.dirPath).toEqual(newCwd);
    expect(updater().props.filesInfo.isSharing).toEqual(true);
    expect(updater().props.filesInfo.items).toEqual(
      List(filesResps.listHomeMockResp.data.metadatas)
    );
    expect(updater().props.sharingsInfo.sharings).toEqual(
      List(filesResps.listSharingsMockResp.data.sharingDirs)
    );
  });
});
