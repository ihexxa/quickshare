import { Set } from "immutable";
// import { mock, instance } from "ts-mockito";

import { ICoreState, mockState } from "../core_state";
import { RootFrame } from "../root_frame";
import { Updater } from "../panes";

describe("RootFrame", () => {
  test("Updater: showSettings", async () => {
    interface TestCase {
      preState: ICoreState;
      postState: ICoreState;
    }

    const mockUpdate = (apply: (prevState: ICoreState) => ICoreState): void => {};
    const tcs: any = [
      {
        preState: {
          displaying: "",
          panes: {
            displaying: "",
            paneNames: Set<string>(["settings", "login"]),
          },
          update: mockUpdate,
        },
        postState: {
          displaying: "settings",
          panes: {
            displaying: "settings",
            paneNames: Set<string>(["settings", "login"]),
          },
          update: mockUpdate,
        },
      },
    ];

    const setState = (patch: any, state: ICoreState): ICoreState => {
      return { ...state, panel: { ...state.panel, ...patch } };
    };

    tcs.forEach((tc: TestCase) => {
      const preState = setState(tc.preState, mockState());
      const postState = setState(tc.postState, mockState());

      const component = new RootFrame(preState.panel);
      Updater.init(preState.panel.panes);
      
      component.showSettings();
      expect(Updater.props).toEqual(postState.panel.panes);
    });
  });
});
