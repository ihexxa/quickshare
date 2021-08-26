import * as React from "react";
import { Set, Map } from "immutable";

import { updater } from "./state_updater";
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

export interface State {}
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
    if (!this.props.login.authed) {
      // TODO: use constant instead
      // TODO: control this with props
      displaying = "login";
    }

    let panesMap: Map<string, JSX.Element> = Map({
      settings: (
        <PaneSettings
          login={this.props.login}
          msg={this.props.msg}
          update={this.props.update}
        />
      ),
      login: (
        <AuthPane
          login={this.props.login}
          update={this.props.update}
          msg={this.props.msg}
        />
      ),
    });

    if (this.props.login.userRole === "admin") {
      panesMap = panesMap.set(
        "admin",
        <AdminPane
          admin={this.props.admin}
          msg={this.props.msg}
          update={this.props.update}
        />
      );
    }

    const panes = panesMap.keySeq().map((paneName: string): JSX.Element => {
      const isDisplay = displaying === paneName ? "" : "hidden";
      return (
        <div key={paneName} className={`${isDisplay}`}>
          {panesMap.get(paneName)}
        </div>
      );
    });

    const btnClass = displaying === "login" ? "hidden" : "";
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

          {panes}
        </div>
      </div>
    );
  }
}
