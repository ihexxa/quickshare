import { Set } from "immutable";

import { ICoreState, initState } from "../core_state";
import { Panes } from "../panes";
import { mockUpdate } from "../../test/helpers";
import { updater } from "../state_updater";

describe("Panes", () => {
  test("Panes: closePane", async () => {
    interface TestCase {
      preState: ICoreState;
      postState: ICoreState;
    }

    const tcs: any = [
      {
        preState: {
          panes: {
            displaying: "settings",
            paneNames: Set<string>(["settings", "login"]),
            update: mockUpdate,
          },
        },
        postState: {
          panes: {
            displaying: "",
            paneNames: Set<string>(["settings", "login"]),
            update: mockUpdate,
          },
        },
      },
    ];

    const setState = (patch: any, state: ICoreState): ICoreState => {
      state.panes = patch.panes;
      return state;
    };

    tcs.forEach((tc: TestCase) => {
      const preState = setState(tc.preState, initState());
      const postState = setState(tc.postState, initState());

      const component = new Panes({
        panes: preState.panes,
        admin: preState.admin,
        login: preState.login,
        update: mockUpdate,
      });
      updater().init(preState);

      component.closePane();
      expect(updater().props).toEqual(postState);
    });
  });
});
