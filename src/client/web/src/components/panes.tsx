import * as React from "react";
import { Set, Map } from "immutable";

import { updater } from "./state_updater";
import { roleAdmin, roleUser, roleVisitor } from "../client";
import { ICoreState, MsgProps } from "./core_state";
import { PaneSettings } from "./pane_settings";
import { AdminPane, AdminProps } from "./pane_admin";
import { AuthPane, LoginProps } from "./pane_login";

export interface PanesProps {
  displaying: string;
  paneNames: Set<string>;
}
export interface Props {
  panes: PanesProps;
  login: LoginProps;
  admin: AdminProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State { }
export class Panes extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  closePane = () => {
    if (this.props.panes.displaying !== "login") {
      updater().displayPane("");
      this.props.update(updater().updatePanes);
    }
  };

  render() {
    let displaying = this.props.panes.displaying;
    const btnClass = displaying === "login" ? "hidden" : "";
    const showSettings = this.props.panes.paneNames.get("settings") && displaying === "settings" ? "" : "hidden";
    const showLogin = this.props.panes.paneNames.get("login") && displaying === "login" ? "" : "hidden";
    const showAdmin = this.props.panes.paneNames.get("admin") && displaying === "admin" ? "" : "hidden";

    return (
      <div id="panes" className={displaying === "" ? "hidden" : ""}>
        <div className="root-container">
          <div className="container">
            <div className="flex-list-container padding-l">
              <h3 className="flex-list-item-l txt-cap">{displaying}</h3>

              <div className="flex-list-item-r">
                <button
                  onClick={this.closePane}
                  className={`red0-bg white-font ${btnClass}`}
                >
                  {this.props.msg.pkg.get("panes.close")}
                </button>
              </div>

            </div>
          </div>

          <div className={`${showSettings}`}>
            <PaneSettings
              login={this.props.login}
              msg={this.props.msg}
              update={this.props.update}
            />
          </div>

          <div className={`${showLogin}`}>
            <AuthPane
              login={this.props.login}
              update={this.props.update}
              msg={this.props.msg}
            />
          </div>

          <div className={`${showAdmin}`}>
            <AdminPane
              admin={this.props.admin}
              msg={this.props.msg}
              update={this.props.update}
            />
          </div>

        </div>
      </div>
    );
  }
}
