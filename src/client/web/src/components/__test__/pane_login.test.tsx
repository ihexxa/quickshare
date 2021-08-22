import { mock, instance } from "ts-mockito";

import { newWithWorker } from "../core_state";
import { updater } from "../state_updater";
import { MockUsersClient } from "../../client/users_mock";
import { FilesClient } from "../../client/files_mock";
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

    const usersClient = new MockUsersClient("");
    const filesClient = new FilesClient("");
    for (let i = 0; i < tests.length; i++) {
      const tc = tests[i];

      usersClient.loginMock(makeNumberResponse(tc.loginStatus));
      usersClient.logoutMock(makeNumberResponse(tc.logoutStatus));
      usersClient.isAuthedMock(makeNumberResponse(tc.isAuthedStatus));
      usersClient.setPwdMock(makeNumberResponse(tc.setPwdStatus));

      const coreState = newWithWorker(mockWorker);
      updater().setClients(usersClient, filesClient);
      updater().init(coreState);
      await updater().initIsAuthed();
      const newState = updater().updateAuthPane(coreState);

      expect(newState.panel.authPane.authed).toEqual(tc.isAuthed);
    }
  });
});
