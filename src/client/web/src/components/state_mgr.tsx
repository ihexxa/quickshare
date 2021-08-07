import * as React from "react";

import { updater as BrowserUpdater } from "./browser.updater";
import { Updater as PanesUpdater } from "./panes";
import { ICoreState, init } from "./core_state";
import { RootFrame } from "./root_frame";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { Updater as LoginPaneUpdater } from "./pane_login";

export interface Props {}
export interface State extends ICoreState {}

export class StateMgr extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = init();
    this.initUpdaters(this.state);
  }

  initUpdaters = (state: ICoreState) => {
    BrowserUpdater().init(state.panel.browser);
    BrowserUpdater().setClients(new UsersClient(""), new FilesClient(""));

    LoginPaneUpdater.init(state.panel.authPane);
    LoginPaneUpdater.setClient(new UsersClient(""));
    LoginPaneUpdater.getCaptchaID()
      .then((ok: boolean) => {
        if (!ok) {
          alert("failed to get captcha id");
        } else {
          this.update(LoginPaneUpdater.setAuthPane);
          console.log(LoginPaneUpdater)
        }
      });

    BrowserUpdater()
      .setHomeItems()
      .then(() => {
        return BrowserUpdater().refreshUploadings();
      })
      .then((_: boolean) => {
        BrowserUpdater().initUploads();
      })
      .then(() => {
        this.update(BrowserUpdater().setBrowser);
      })
      .then(() => {
        return PanesUpdater.self();
      })
      .then(() => {
        return PanesUpdater.listRoles();
      })
      .then((_: boolean) => {
        return PanesUpdater.listUsers();
      })
      .then((_: boolean) => {
        this.update(PanesUpdater.updateState);
      });
  };

  update = (apply: (prevState: ICoreState) => ICoreState): void => {
    this.setState(apply(this.state));
  };

  render() {
    return (
      <RootFrame
        authPane={this.state.panel.authPane}
        displaying={this.state.panel.displaying}
        update={this.update}
        browser={this.state.panel.browser}
        panes={this.state.panel.panes}
      />
    );
  }
}
