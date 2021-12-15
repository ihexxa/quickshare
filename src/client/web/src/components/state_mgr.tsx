import * as React from "react";

import { initUploadMgr } from "../worker/upload_mgr";
import BgWorker from "../worker/upload.bg.worker";
import { FgWorker } from "../worker/upload.fg.worker";

import { alertMsg } from "../common/env";
import { getErrMsg } from "../common/utils";
import { updater } from "./state_updater";
import { ICoreState, newState } from "./core_state";
import { RootFrame } from "./root_frame";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { SettingsClient } from "../client/settings";
import { IUsersClient, IFilesClient, ISettingsClient } from "../client";

export interface Props {}
export interface State extends ICoreState {}

export class StateMgr extends React.Component<Props, State, {}> {
  private usersClient: IUsersClient = new UsersClient("");
  private filesClient: IFilesClient = new FilesClient("");
  private settingsClient: ISettingsClient = new SettingsClient("");

  constructor(p: Props) {
    super(p);
    const worker = window.Worker == null ? new FgWorker() : new BgWorker();
    initUploadMgr(worker);
    this.state = newState();

    const query = new URLSearchParams(document.location.search.substring(1));
    this.initUpdater(this.state, query); // don't await
  }

  setUsersClient = (client: IUsersClient) => {
    this.usersClient = client;
  };

  setFilesClient = (client: IFilesClient) => {
    this.filesClient = client;
  };

  setSettingsClient = (client: ISettingsClient) => {
    this.settingsClient = client;
  };

  initUpdater = async (
    state: ICoreState,
    query: URLSearchParams
  ): Promise<void> => {
    updater().init(state);
    if (
      this.usersClient == null ||
      this.filesClient == null ||
      this.settingsClient == null
    ) {
      console.error("updater's clients are not inited");
      return;
    }
    updater().setClients(
      this.usersClient,
      this.filesClient,
      this.settingsClient
    );

    return updater()
      .initAll(query)
      .then((status: string) => {
        if (status !== "") {
          alertMsg(getErrMsg(state.msg.pkg, "op.fail", status.toString()));
        }
        this.update(updater().updateAll);
      });
  };

  update = (update: (prevState: ICoreState) => ICoreState): void => {
    this.setState(update(this.state));
  };

  render() {
    return (
      <RootFrame
        filesInfo={this.state.filesInfo}
        uploadingsInfo={this.state.uploadingsInfo}
        sharingsInfo={this.state.sharingsInfo}
        msg={this.state.msg}
        login={this.state.login}
        admin={this.state.admin}
        ui={this.state.ui}
        update={this.update}
      />
    );
  }
}
