import { Set } from "immutable";

import { ICoreState, initState } from "../core_state";
import { RootFrame } from "../root_frame";
// import { Updater } from "../panes";
import { updater } from "../state_updater";

xdescribe("RootFrame", () => {
  test("component: showSettings", async () => {
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
      const preState = setState(tc.preState, initState());
      const postState = setState(tc.postState, initState());

      const component = new RootFrame(preState.panel);
      updater().init(preState);
      
      // component.showSettings();
      expect(updater().props).toEqual(postState);
    });
  });
});
