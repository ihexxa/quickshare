import { mock, instance } from "ts-mockito";

import { Panes } from "../panes";
import { ICoreState, newWithWorker } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";

describe("Panes", () => {
  test("closePane", async () => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);

    const coreState = newWithWorker(mockWorker);
    const panes = new Panes({
        panes: coreState.panes,
        admin: coreState.admin,
        login: coreState.login,
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
