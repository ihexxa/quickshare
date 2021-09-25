import * as React from "react";

import { updater } from "./state_updater";
import { ICoreState, newState } from "./core_state";
import { RootFrame } from "./root_frame";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { IUsersClient, IFilesClient } from "../client";

export interface Props { }
export interface State extends ICoreState { }

export class StateMgr extends React.Component<Props, State, {}> {
  private usersClient: IUsersClient = new UsersClient("");
  private filesClient: IFilesClient = new FilesClient("");

  constructor(p: Props) {
    super(p);
    this.state = newState();
    this.initUpdater(this.state); // don't await
  }

  setUsersClient = (client: IUsersClient) => {
    this.usersClient = client;
  };

  setFilesClient = (client: IFilesClient) => {
    this.filesClient = client;
  };

  initUpdater = async (state: ICoreState): Promise<void> => {
    updater().init(state);
    if (this.usersClient == null || this.filesClient == null) {
      console.error("updater's clients are not inited");
      return;
    }
    updater().setClients(this.usersClient, this.filesClient);

    const params = new URLSearchParams(document.location.search.substring(1));
    return updater()
      .initAll(params)
      .then(() => {
        this.update(updater().updateBrowser);
        this.update(updater().updateLogin);
        this.update(updater().updatePanes);
        this.update(updater().updateAdmin);

        updater().initLan();
        this.update(updater().updateMsg);
      });
  };

  update = (apply: (prevState: ICoreState) => ICoreState): void => {
    this.setState(apply(this.state));
  };

  render() {
    return (
      <RootFrame
        browser={this.state.browser}
        msg={this.state.msg}
        panes={this.state.panes}
        login={this.state.login}
        admin={this.state.admin}
        ui={this.state.ui}
        update={this.update}
      />
    );
  }
}
