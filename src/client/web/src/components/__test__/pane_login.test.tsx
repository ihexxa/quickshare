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
      authed: true,
      captchaID: "",
    });

    // panes
    expect(updater().props.panes).toEqual({
      userRole: "admin",
      displaying: "",
      paneNames: Set(["settings", "login", "admin"]),
    });

    // admin
    let usersMap = Map({});
    usersResps.listUsersMockResp.data.users.forEach((user: User) => {
      usersMap = usersMap.set(user.name, user);
    });
    let roles = Set<string>();
    Object.keys(usersResps.listRolesMockResp.data.roles).forEach(
      (role: string) => {
        roles = roles.add(role);
      }
    );
    expect(updater().props.admin).toEqual({
      users: usersMap,
      roles: roles,
    });
  });
});
