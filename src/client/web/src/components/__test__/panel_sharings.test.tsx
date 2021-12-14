import { mock, instance, verify, when, anything } from "ts-mockito";
import { List } from "immutable";

import { SharingsPanel } from "../panel_sharings";
import { initUploadMgr } from "../../worker/upload_mgr";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockSettingsClient } from "../../client/settings_mock";

xdescribe("SharingsPanel", () => {
  const initSharingsPanel = (): any => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);
    initUploadMgr(mockWorker);

    const coreState = newState();
    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    const settingsCl = new MockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const sharingsPanel = new SharingsPanel({
      sharingsInfo: coreState.sharingsInfo,
      msg: coreState.msg,
      login: coreState.login,
      ui: coreState.ui,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    return {
      sharingsPanel,
      usersCl,
      filesCl,
    };
  };

  test("delete sharing", async () => {
    const { sharingsPanel, usersCl, filesCl } = initSharingsPanel();

    const newSharings = ["mock_sharingfolder1", "mock_sharingfolder2"];

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

    await sharingsPanel.deleteSharing();

    // TODO: check delSharing's input
    expect(updater().props.filesInfo.isSharing).toEqual(false);
    expect(updater().props.sharingsInfo.sharings).toEqual(List(newSharings));
  });
});
