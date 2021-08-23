import * as React from "react";
import { List } from "immutable";

import { updater } from "./state_updater";
import { ICoreState, newState } from "./core_state";
import { RootFrame } from "./root_frame";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { IUsersClient, IFilesClient } from "../client";

export interface Props {}
export interface State extends ICoreState {}

export class StateMgr extends React.Component<Props, State, {}> {
  private usersClient: IUsersClient = new UsersClient("");
  private filesClient: IFilesClient = new FilesClient("");

  constructor(p: Props) {
    super(p);
    this.state = newState();
    this.initUpdater(this.state);
  }

  setUsersClient = (client: IUsersClient) => {
    this.usersClient = client;
  };

  setFilesClient = (client: IFilesClient) => {
    this.filesClient = client;
  };

  initUpdater = (state: ICoreState) => {
    updater().init(state);
    updater().setClients(new UsersClient(""), new FilesClient(""));

    const params = new URLSearchParams(document.location.search.substring(1));
    updater()
      .getCaptchaID()
      .then((ok: boolean) => {
        if (!ok) {
          alert("failed to get captcha id");
        } else {
          this.update(updater().updateLogin);
        }
      });

    updater()
      .refreshUploadings()
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
        return updater().initUploads();
      })
      .then(() => {
        return updater().isSharing(updater().props.browser.dirPath.join("/"));
      })
      .then(() => {
        return updater().listSharings();
      })
      .then(() => {
        this.update(updater().updateBrowser);
      })
      .then(() => {
        return updater().self();
      })
      .then(() => {
        if (updater().props.panes.userRole === "admin") {
          // TODO: remove hardcode
          return updater().listRoles();
        }
      })
      .then(() => {
        if (updater().props.panes.userRole === "admin") {
          // TODO: remove hardcode
          return updater().listUsers();
        }
      })
      .then(() => {
        this.update(updater().updatePanes);
      });
  };

  update = (apply: (prevState: ICoreState) => ICoreState): void => {
    this.setState(apply(this.state));
  };

  render() {
    return (
      <RootFrame
        browser={this.state.browser}
        panes={{
          panes: this.state.panes,
          login: this.state.login,
          admin: this.state.admin,
        }}
        update={this.update}
      />
    );
  }
}
