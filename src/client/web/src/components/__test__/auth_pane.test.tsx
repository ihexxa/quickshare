import * as React from "react";

import { init } from "../core_state";
import { Updater } from "../auth_pane";
import { MockUsersClient } from "../../client/users_mock";

describe("AuthPane", () => {
  test("Updater: initIsAuthed", () => {
    const tests = [
      {
        loginStatus: 200,
        isAuthed: true,
      },
      {
        loginStatus: 500,
        isAuthed: false,
      },
    ];

    const client = new MockUsersClient("foobarurl");
    Updater.setClient(client);
    const coreState = init();

    tests.forEach(async (tc) => {
      client.mockisAuthedResp(tc.loginStatus);
      await Updater.initIsAuthed().then(() => {
        const newState = Updater.setAuthPane(coreState);
        expect(newState.panel.authPane.authed).toEqual(tc.isAuthed);
      });
    });
  });
});
