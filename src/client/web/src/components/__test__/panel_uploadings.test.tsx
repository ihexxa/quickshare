import { mock, instance } from "ts-mockito";

import { UploadingsPanel } from "../panel_uploadings";
import { initUploadMgr } from "../../worker/upload_mgr";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";
import { NewMockUsersClient, resps as usersResps } from "../../client/users_mock";
import { NewMockFilesClient, resps as filesResps } from "../../client/files_mock";
import { NewMockSettingsClient } from "../../client/settings_mock";

describe("UploadingsPanel", () => {
  const initUploadingsPanel = (): any => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);
    initUploadMgr(mockWorker);

    const coreState = newState();
    const usersCl = NewMockUsersClient("");
    const filesCl = NewMockFilesClient("");
    const settingsCl = NewMockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const uploadingsPanel = new UploadingsPanel({
      uploadingsInfo: coreState.uploadingsInfo,
      msg: coreState.msg,
      login: coreState.login,
      ui: coreState.ui,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    return {
      uploadingsPanel,
      usersCl,
      filesCl,
    };
  };

  test("todo", async () => {});
});
