import { mock, instance } from "ts-mockito";

import { initWithWorker } from "../core_state";
import { Updater } from "../pane_login";
import { MockUsersClient } from "../../client/users_mock";
import { Response } from "../../client";
import { MockWorker } from "../../worker/interface";

describe("AuthPane", () => {
  const mockWorkerClass = mock(MockWorker);
  const mockWorker = instance(mockWorkerClass);

  const makePromise = (ret: any): Promise<any> => {
    return new Promise<any>((resolve) => {
      resolve(ret);
    });
  };
  const makeNumberResponse = (status: number): Promise<Response> => {
    return makePromise({
      status: status,
      statusText: "",
      data: {},
    });
  };

  test("Updater-initIsAuthed", async () => {
    const tests = [
      {
        loginStatus: 200,
        logoutStatus: 200,
        isAuthedStatus: 200,
        setPwdStatus: 200,
        isAuthed: true,
      },
      {
        loginStatus: 200,
        logoutStatus: 200,
        isAuthedStatus: 500,
        setPwdStatus: 200,
        isAuthed: false,
      },
    ];

    const client = new MockUsersClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      client.loginMock(makeNumberResponse(tc.loginStatus));
      client.logoutMock(makeNumberResponse(tc.logoutStatus));
      client.isAuthedMock(makeNumberResponse(tc.isAuthedStatus));
      client.setPwdMock(makeNumberResponse(tc.setPwdStatus));

      const coreState = initWithWorker(mockWorker);
      Updater.setClient(client);
      Updater.init(coreState.panel.authPane);
      await Updater.initIsAuthed();
      const newState = Updater.setAuthPane(coreState);

      expect(newState.panel.authPane.authed).toEqual(tc.isAuthed);
    }
  });
});
