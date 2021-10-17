import * as React from "react";

import { ICoreState, MsgProps, UIProps } from "./core_state";
import { Browser, BrowserProps } from "./browser";
import { LoginProps } from "./pane_login";
import { Panes, PanesProps } from "./panes";
import { AdminProps } from "./pane_admin";
import { TopBar } from "./topbar";
import { roleVisitor } from "../client";

export interface Props {
  browser: BrowserProps;
  panes: PanesProps;
  admin: AdminProps;
  login: LoginProps;
  msg: MsgProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}
export class RootFrame extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  render() {
    let bgStyle = undefined;
    if (this.props.login.preferences != null && this.props.login.preferences.bg.url !== "") {
      bgStyle = {
        background: `url("${this.props.login.preferences.bg.url}") ${this.props.login.preferences.bg.repeat} ${this.props.login.preferences.bg.position} ${this.props.login.preferences.bg.align}`,
      };
    } else if (this.props.ui.bg.url !== "") {
      bgStyle = {
        background: `url("${this.props.ui.bg.url}") ${this.props.ui.bg.repeat} ${this.props.ui.bg.position} ${this.props.ui.bg.align}`,
      };
    } else {
      bgStyle = {};
    }

    const showBrowser =
      this.props.login.userRole === roleVisitor && !this.props.browser.isSharing
        ? "hidden"
        : "";

    return (
      <div className="theme-white desktop">
        <div id="bg" className="bg bg-img font-m" style={bgStyle}>
          <Panes
            panes={this.props.panes}
            login={this.props.login}
            admin={this.props.admin}
            ui={this.props.ui}
            msg={this.props.msg}
            update={this.props.update}
          />

          <TopBar
            login={this.props.login}
            panes={this.props.panes}
            msg={this.props.msg}
            update={this.props.update}
          />

          <div className={`container-center ${showBrowser}`}>
            <Browser
              browser={this.props.browser}
              msg={this.props.msg}
              login={this.props.login}
              ui={this.props.ui}
              update={this.props.update}
            />
          </div>

          <div id="tail" className="container-center black0-font">
            <a href="https://github.com/ihexxa/quickshare">Quickshare</a> -
            quick and simple file sharing.
          </div>
        </div>
      </div>
    );
  }
}
