import { mock, instance } from "ts-mockito";

import { PaneSettings } from "../pane_settings";
import { initUploadMgr } from "../../worker/upload_mgr";
import { ICoreState, newState } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";
import {
  NewMockUsersClient,
  resps as usersResps,
} from "../../client/users_mock";
import {
  NewMockFilesClient,
  resps as filesResps,
} from "../../client/files_mock";
import { NewMockSettingsClient } from "../../client/settings_mock";
import { MockWebEnv } from "../../test/helpers";
import { SetEnv } from "../../common/env";

describe("PaneSettings", () => {
  const initPaneSettings = (): any => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);
    initUploadMgr(mockWorker);

    const coreState = newState();
    const usersCl = NewMockUsersClient("");
    const filesCl = NewMockFilesClient("");
    const settingsCl = NewMockSettingsClient("");

    updater().init(coreState);
    updater().setClients(usersCl, filesCl, settingsCl);

    const paneSettings = new PaneSettings({
      msg: coreState.msg,
      login: coreState.login,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    return {
      paneSettings,
      usersCl,
      filesCl,
      settingsCl,
    };
  };

  test("Preferences settings alerts", async () => {
    const { paneSettings, usersCl, filesCl, settingCl } = initPaneSettings();
    const env = new MockWebEnv();
    SetEnv(env);

    await paneSettings.setLan("en_US");
    expect(env.alertMsg.mock.calls.length).toBe(1);

    await paneSettings.setTheme("light");
    expect(env.alertMsg.mock.calls.length).toBe(2);
  });
});
