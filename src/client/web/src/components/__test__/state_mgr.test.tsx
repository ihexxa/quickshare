import { List, Set, Map } from "immutable";
import { mock, instance } from "ts-mockito";

import { StateMgr } from "../state_mgr";
import { User } from "../../client";
import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
import { ICoreState, newWithWorker } from "../core_state";
import { MockWorker } from "../../worker/interface";

describe("State Manager", () => {
  test("initUpdater", async () => {
    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");

    const mockWorkerClass = mock(MockWorker);
    const mockWorker = instance(mockWorkerClass);

    const mgr = new StateMgr({}); // it will call initUpdater
    mgr.setUsersClient(usersCl);
    mgr.setFilesClient(filesCl);

    // TODO: depress warning
    mgr.update = (apply: (prevState: ICoreState) => ICoreState): void => {
      // no op
    };

    const coreState = newWithWorker(mockWorker);
    await mgr.initUpdater(coreState);

    // browser
    expect(coreState.browser.dirPath.join("/")).toEqual("mock_home/files");
    expect(coreState.browser.isSharing).toEqual(true);
    expect(coreState.browser.sharings).toEqual(
      List(filesResps.listSharingsMockResp.data.sharingDirs)
    );
    expect(coreState.browser.uploadings).toEqual(
      List(filesResps.listUploadingsMockResp.data.uploadInfos)
    );
    expect(coreState.browser.items).toEqual(
      List(filesResps.listHomeMockResp.data.metadatas)
    );

    // panes
    expect(coreState.panes).toEqual({
      displaying: "",
      paneNames: Set(["settings", "login", "admin"]),
    });

    // login
    expect(coreState.login).toEqual({
      userID: "0",
      userName: "mockUser",
      userRole: "admin",
      usedSpace: "256",
      quota: {
        spaceLimit: "7",
        uploadSpeedLimit: 3,
        downloadSpeedLimit: 3,
      },
      authed: true,
      captchaID: "mockCaptchaID",
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
    expect(coreState.admin).toEqual({
      users: usersMap,
      roles: roles,
    });
  });
});
