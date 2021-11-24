// import { mock, instance, verify, when, anything } from "ts-mockito";
// import { List } from "immutable";

// import { Browser } from "../browser";
// import { initUploadMgr } from "../../worker/upload_mgr";
// import { ICoreState, newState } from "../core_state";
// import { updater } from "../state_updater";
// import { MockWorker } from "../../worker/interface";
// import { MockUsersClient, resps as usersResps } from "../../client/users_mock";
// import { MockFilesClient, resps as filesResps } from "../../client/files_mock";
// import { MockSettingsClient } from "../../client/settings_mock";

// describe("Browser", () => {
//   const initBrowser = (): any => {
//     const mockWorkerClass = mock(MockWorker);
//     const mockWorker = instance(mockWorkerClass);
//     initUploadMgr(mockWorker);

//     const coreState = newState();
//     const usersCl = new MockUsersClient("");
//     const filesCl = new MockFilesClient("");
//     const settingsCl = new MockSettingsClient("");

//     updater().init(coreState);
//     updater().setClients(usersCl, filesCl, settingsCl);

//     const browser = new Browser({
//       browser: coreState.browser,
//       msg: coreState.msg,
//       login: coreState.login,
//       ui: coreState.ui,
//       update: (updater: (prevState: ICoreState) => ICoreState) => {},
//     });

//     return {
//       browser,
//       usersCl,
//       filesCl,
//     };
//   };

//   test("addUploads", async () => {
//     const { browser, usersCl, filesCl } = initBrowser();

//     const newSharings = [
//       "mock_sharingfolder1",
//       "mock_sharingfolder2",
//       "newSharing",
//     ];

//     filesCl.setMock({
//       ...filesResps,
//       listSharingsMockResp: {
//         status: 200,
//         statusText: "",
//         data: {
//           sharingDirs: newSharings,
//         },
//       },
//     });

//     await browser.addSharing();

//     // TODO: check addSharing's input
//     expect(updater().props.browser.isSharing).toEqual(true);
//     expect(updater().props.browser.sharings).toEqual(List(newSharings));
//   });

//   test("deleteUploads", async () => {
//     const { browser, usersCl, filesCl } = initBrowser();

//     const newSharings = ["mock_sharingfolder1", "mock_sharingfolder2"];

//     filesCl.setMock({
//       ...filesResps,
//       listSharingsMockResp: {
//         status: 200,
//         statusText: "",
//         data: {
//           sharingDirs: newSharings,
//         },
//       },
//     });

//     await browser.deleteSharing();

//     // TODO: check delSharing's input
//     expect(updater().props.browser.isSharing).toEqual(false);
//     expect(updater().props.browser.sharings).toEqual(List(newSharings));
//   });

//   test("chdir", async () => {
//     const { browser, usersCl, filesCl } = initBrowser();

//     const newSharings = ["mock_sharingfolder1", "mock_sharingfolder2"];
//     const newCwd = List(["newPos", "subFolder"]);

//     filesCl.setMock({
//       ...filesResps,
//       listSharingsMockResp: {
//         status: 200,
//         statusText: "",
//         data: {
//           sharingDirs: newSharings,
//         },
//       },
//     });

//     await browser.chdir(newCwd);

//     expect(updater().props.browser.dirPath).toEqual(newCwd);
//     expect(updater().props.browser.isSharing).toEqual(true);
//     expect(updater().props.browser.sharings).toEqual(
//       List(filesResps.listSharingsMockResp.data.sharingDirs)
//     );
//     expect(updater().props.browser.items).toEqual(
//       List(filesResps.listHomeMockResp.data.metadatas)
//     );
//   });
// });
