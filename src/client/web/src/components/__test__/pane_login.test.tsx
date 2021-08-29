import { List, Set, Map } from "immutable";
import { mock, instance } from "ts-mockito";

import { User } from "../../client";
import { AuthPane } from "../pane_login";
import { ICoreState, newWithWorker } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker } from "../../worker/interface";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";

describe("Login", () => {
  test("login", async () => {
    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);

    const coreState = newWithWorker(mockWorker);
    const pane = new AuthPane({
      login: coreState.login,
      msg: coreState.msg,
      update: (updater: (prevState: ICoreState) => ICoreState) => {},
    });

    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    updater().init(coreState);
    updater().setClients(usersCl, filesCl);

    await pane.login();

    // TODO: state is not checked

    // login
    expect(updater().props.login).toEqual({
      userID: "0",
      userName: "mockUser",
      userRole: "admin",
      authed: true,
      usedSpace: "256",
      quota: {
        spaceLimit: "7",
        uploadSpeedLimit: 3,
        downloadSpeedLimit: 3,
      },
      captchaID: "",
    });

    // panes
    expect(updater().props.panes).toEqual({
      displaying: "",
      paneNames: Set(["settings", "login", "admin"]),
    });
  });
});
