import { List, Set, Map } from "immutable";
import { mock, instance } from "ts-mockito";

import { User, UploadInfo } from "../../client";
import { AuthPane } from "../pane_login";
import { ICoreState, newWithWorker } from "../core_state";
import { updater } from "../state_updater";
import { MockWorker, UploadState, UploadEntry } from "../../worker/interface";
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
      update: (updater: (prevState: ICoreState) => ICoreState) => { },
    });

    const usersCl = new MockUsersClient("");
    const filesCl = new MockFilesClient("");
    updater().init(coreState);
    updater().setClients(usersCl, filesCl);

    await pane.login();

    // TODO: state is not checked

    // browser
    expect(coreState.browser.dirPath.join("/")).toEqual("mock_home/files");
    expect(coreState.browser.isSharing).toEqual(true);
    expect(coreState.browser.sharings).toEqual(
      List(filesResps.listSharingsMockResp.data.sharingDirs)
    );
    expect(coreState.browser.uploadings).toEqual(
      List<UploadEntry>(
        filesResps.listUploadingsMockResp.data.uploadInfos.map(
          (info: UploadInfo) => {
            return {
              file: undefined,
              filePath: info.realFilePath,
              size: info.size,
              uploaded: info.uploaded,
              state: UploadState.Ready,
              err: "",
            };
          }
        )
      )
    );

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
