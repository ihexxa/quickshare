import * as React from "react";
import { List, Set } from "immutable";

import { updater } from "./state_updater";
import { ICoreState, newState } from "./core_state";
import { RootFrame } from "./root_frame";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { IUsersClient, IFilesClient, roleAdmin, roleVisitor } from "../client";
import { alertMsg } from "../common/env";

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
      .initIsAuthed()
      .then(() => {
        return updater().self();
      })
      .then(() => {
        const dir = params.get("dir");
        if (dir != null && dir !== "") {
          const dirPath = List(dir.split("/"));
          return updater().setItems(dirPath);
        } else {
          return updater().setHomeItems();
        }
      })
      .then(() => {
        return updater().isSharing(updater().props.browser.dirPath.join("/"));
      })
      .then(() => {
        // init browser content
        if (updater().props.login.userRole === roleVisitor) {
          if (updater().props.browser.isSharing) {
            // sharing with visitor
            updater().setPanes(Set<string>(["login"]));
            updater().displayPane("");
            return Promise.all([]);
          }

          // redirect to login
          updater().setPanes(Set<string>(["login"]));
          updater().displayPane("login");
          return Promise.all([updater().getCaptchaID()]);
        }

        if (updater().props.login.userRole === roleAdmin) {
          updater().setPanes(Set<string>(["login", "settings", "admin"]));
        } else {
          updater().setPanes(Set<string>(["login", "settings"]));
        }
        updater().displayPane("");

        return Promise.all([
          updater().refreshUploadings(),
          updater().initUploads(),
          updater().listSharings(),
        ]);
      })
      .then(() => {
        // init admin content
        if (updater().props.login.userRole === roleAdmin) {
          return Promise.all([updater().listRoles(), updater().listUsers()]);
        }
        return;
      })
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
