import * as React from "react";
import { List } from "immutable";

import { updater } from "./state_updater";
import { ICoreState, newState } from "./core_state";
import { RootFrame } from "./root_frame";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { IUsersClient, IFilesClient } from "../client";
import { alertMsg } from "../common/env";

export interface Props {}
export interface State extends ICoreState {}

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
        this.update(updater().updateLogin);
      })
      .then(() => {
        if (updater().props.login.authed) {
          updater().displayPane("");
        } else {
          updater().displayPane("login");
        }
      })
      .then(() => {
        return updater().getCaptchaID();
      })
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.state.msg.pkg.get("stateMgr.cap.fail"));
        } else {
          this.update(updater().updateLogin);
        }
      })
      .then(() => {
        return updater().refreshUploadings();
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
        if (updater().props.login.userRole === "admin") {
          // TODO: remove hardcode
          return updater().listRoles();
        }
      })
      .then(() => {
        if (updater().props.login.userRole === "admin") {
          // TODO: remove hardcode
          return updater().listUsers();
        }
      })
      .then(() => {
        this.update(updater().updatePanes);
        this.update(updater().updateAdmin);
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
        update={this.update}
      />
    );
  }
}
