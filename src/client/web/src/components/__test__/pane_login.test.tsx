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
      userRole: coreState.login.userRole,
      authed: coreState.login.authed,
      captchaID: coreState.login.captchaID,
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
      userRole: "admin",
      authed: true,
      captchaID: "",
    });

    // panes
    expect(updater().props.panes).toEqual({
      displaying: "",
      paneNames: Set(["settings", "login", "admin"]),
    });
  });
});
