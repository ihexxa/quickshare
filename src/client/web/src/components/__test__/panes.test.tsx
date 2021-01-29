import { Set } from "immutable";

import { ICoreState, mockState } from "../core_state";
import { Panes, Updater } from "../panes";
import { mockUpdate } from "../../test/helpers";

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
      state.panel.panes = patch.panes;
      return state;
    };

    tcs.forEach((tc: TestCase) => {
      const preState = setState(tc.preState, mockState());
      const postState = setState(tc.postState, mockState());

      const component = new Panes(preState.panel.panes);
      Updater.init(preState.panel.panes);

      component.closePane();
      expect(Updater.props).toEqual(postState.panel.panes);
    });
  });
});
