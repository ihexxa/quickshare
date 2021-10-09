import { mock, instance } from "ts-mockito";

import { Panes } from "../panes";
import { ICoreState, newState } from "../core_state";
import { initUploadMgr } from "../../worker/upload_mgr";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";

describe("Panes", () => {
  test("closePane", async () => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);
    initUploadMgr(mockWorker);

    const coreState = newState();
    const panes = new Panes({
      panes: coreState.panes,
      admin: coreState.admin,
      login: coreState.login,
      ui: coreState.ui,
      msg: coreState.msg,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    updater().init(coreState);

    panes.closePane();

    expect(updater().props.panes).toEqual({
      displaying: "",
      paneNames: coreState.panes.paneNames,
    });
  });
});
