import * as React from "react";
import { List } from "immutable";
import { alertMsg, confirmMsg } from "../common/env";

import { ICoreState, MsgProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { PanesProps } from "./panes";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";

export interface State {}
export interface Props {
  login: LoginProps;
  panes: PanesProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class TopBar extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  showSettings = () => {
    updater().displayPane("settings");
    this.props.update(updater().updatePanes);
  };

  showAdmin = async () => {
    return updater()
      .self()
      .then(() => {
        // TODO: remove hardcode role
        if (this.props.login.authed && this.props.login.userRole === "admin") {
          return Promise.all([updater().listRoles(), updater().listUsers()]);
        }
      })
      .then(() => {
        updater().displayPane("admin");
        this.props.update(updater().updateAdmin);
        this.props.update(updater().updatePanes);
      });
  };

  logout = async (): Promise<void> => {
    if (!confirmMsg(this.props.msg.pkg.get("logout.confirm"))) {
      return;
    }

    return updater()
      .logout()
      .then((ok: boolean) => {
        if (ok) {
          const params = new URLSearchParams(
            document.location.search.substring(1)
          );
          return updater().initAll(params);
        } else {
          alertMsg(this.props.msg.pkg.get("login.logout.fail"));
        }
      })
      .then(() => {
        return this.refreshCaptcha();
      })
      .then(() => {
        this.props.update(updater().updateBrowser);
        this.props.update(updater().updateLogin);
        this.props.update(updater().updatePanes);
        this.props.update(updater().updateAdmin);
        this.props.update(updater().updateUI);
        this.props.update(updater().updateMsg);
      });
  };

  refreshCaptcha = async () => {
    return updater()
      .getCaptchaID()
      .then(() => {
        this.props.update(updater().updateLogin);
      });
  };

  render() {
    const showUserInfo = this.props.login.authed ? "" : "hidden";
    const showLogin = this.props.login.authed ? "" : "hidden";
    const showSettings = this.props.panes.paneNames.get("settings")
      ? ""
      : "hidden";
    const showAdmin = this.props.panes.paneNames.get("admin") ? "" : "hidden";

    return (
      <div id="top-bar">
        <Flexbox
          children={List([
            <a
              id="topbar-title"
              href="https://github.com/ihexxa/quickshare"
              target="_blank"
              className="h5"
            >
              Quickshare
            </a>,

            <Flexbox
              children={List([
                <span className={`${showUserInfo}`}>
                  <span id="topbar-user-info">
                    {this.props.login.userName}
                  </span>
                </span>,

                <button
                  onClick={this.showSettings}
                  className={`margin-r-m ${showSettings}`}
                >
                  {this.props.msg.pkg.get("settings")}
                </button>,

                <button
                  onClick={this.showAdmin}
                  className={`margin-r-m ${showAdmin}`}
                >
                  {this.props.msg.pkg.get("admin")}
                </button>,

                <button
                  onClick={this.logout}
                  className={`${showLogin}`}
                >
                  {this.props.msg.pkg.get("login.logout")}
                </button>,
              ])}
              childrenStyles={List([{}, {}, {}, {}])}
            />,
          ])}
          childrenStyles={List([
            {},
            { justifyContent: "flex-end", alignItems: "center" },
          ])}
        />
      </div>
    );
  }
}
