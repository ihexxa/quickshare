import * as React from "react";

import { ICoreState, MsgProps, UIProps } from "./core_state";
import { Browser, BrowserProps } from "./browser";
import { LoginProps } from "./pane_login";
import { Panes, PanesProps } from "./panes";
import { AdminProps } from "./pane_admin";
import { TopBar } from "./topbar";

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
    const wallpaperStyle =
      this.props.ui.wallpaper !== ""
        ? {
            background: `url("${this.props.ui.wallpaper}") ${this.props.ui.repeat} ${this.props.ui.position} ${this.props.ui.align}`,
          }
        : {};

    return (
      <div className="theme-white desktop">
        <div id="bg" className="bg bg-img font-m" style={wallpaperStyle}>
          <Panes
            panes={this.props.panes}
            login={this.props.login}
            admin={this.props.admin}
            msg={this.props.msg}
            update={this.props.update}
          />

          <TopBar
            login={this.props.login}
            msg={this.props.msg}
            update={this.props.update}
          />

          <div className="container-center">
            <Browser
              browser={this.props.browser}
              msg={this.props.msg}
              update={this.props.update}
            />
          </div>

          <div id="tail" className="container-center black0-font">
            <a href="https://github.com/ihexxa/quickshare">Quickshare</a> -
            sharing in simple way.
          </div>
        </div>
      </div>
    );
  }
}
