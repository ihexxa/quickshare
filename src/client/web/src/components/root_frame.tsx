import * as React from "react";

import { ICoreState } from "./core_state";
import { Browser, Props as BrowserProps } from "./browser";
import { Props as PaneLoginProps } from "./pane_login";
import { Panes, Props as PanesProps } from "./panes";
import { TopBar } from "./topbar";

export interface Props {
  displaying: string;
  browser: BrowserProps;
  authPane: PaneLoginProps;
  panes: PanesProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}
export class RootFrame extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  render() {
    return (
      <div className="theme-white desktop">
        <div id="bg" className="bg bg-img font-m">
          <Panes
            userRole={this.props.panes.userRole}
            displaying={this.props.panes.displaying}
            paneNames={this.props.panes.paneNames}
            login={this.props.authPane}
            admin={this.props.panes.admin}
            update={this.props.update}
          />

          <TopBar update={this.props.update}></TopBar>

          <div className="container-center">
            <Browser
              dirPath={this.props.browser.dirPath}
              items={this.props.browser.items}
              uploadings={this.props.browser.uploadings}
              sharings={this.props.browser.sharings}
              isSharing={this.props.browser.isSharing}
              update={this.props.update}
              uploadFiles={this.props.browser.uploadFiles}
              uploadValue={this.props.browser.uploadValue}
              isVertical={this.props.browser.isVertical}
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
