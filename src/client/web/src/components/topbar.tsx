import * as React from "react";
import { List, Set } from "immutable";
import { alertMsg } from "../common/env";

import { ICoreState, MsgProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { PanesProps } from "./panes";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";

export interface State { }
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
    return updater()
      .logout()
      .then((ok: boolean) => {
        if (ok) {
          const params = new URLSearchParams(document.location.search.substring(1));
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

        updater().initLan();
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
    const showSettings = this.props.panes.paneNames.get("settings") ? "" : "hidden";
    const showAdmin = this.props.panes.paneNames.get("admin") ? "" : "hidden";

    return (
      <div
        id="top-bar"
        className="top-bar cyan1-font padding-t-m padding-b-m padding-l-l padding-r-l"
      >
        <Flexbox
          children={List([
            <a
              href="https://github.com/ihexxa/quickshare"
              target="_blank"
              className="h5"
            >
              Quickshare
            </a>,

            <Flexbox
              children={List([
                <span className={`${showUserInfo}`}>
                  <span className="grey3-font font-s">
                    {this.props.login.userName}
                  </span>
                  &nbsp;-&nbsp;
                  <span className="grey0-font font-s margin-r-m">
                    {this.props.login.userRole}
                  </span>
                </span>,

                <button
                  onClick={this.showSettings}
                  className={`grey3-bg grey4-font margin-r-m ${showSettings}`}
                  style={{ minWidth: "7rem" }}
                >
                  {this.props.msg.pkg.get("settings")}
                </button>,

                <button
                  onClick={this.showAdmin}
                  className={`grey3-bg grey4-font margin-r-m ${showAdmin}`}
                  style={{ minWidth: "7rem" }}
                >
                  {this.props.msg.pkg.get("admin")}
                </button>,

                <button onClick={this.logout} className={`${showLogin}`}>
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
