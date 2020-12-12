import * as React from "react";

import { ICoreState } from "./core_state";
import { Browser, Props as BrowserProps } from "./browser";
import { AuthPane, Props as AuthPaneProps } from "./auth_pane";

export interface Props {
  displaying: string;
  browser: BrowserProps;
  authPane: AuthPaneProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;

  static init = (props: Props) => (Updater.props = { ...props });

  static setPanel = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panel: { ...prevState.panel, ...Updater.props },
    };
  };
}

export interface State {}
export class Panel extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  constructor(p: Props) {
    super(p);
    Updater.init(p);
    this.update = p.update;
  }

  render() {
    return (
      <div className="theme-white">
        <div id="bg" className="bg">
          <div id="win" className="win">
            <div id="win-head" className="win-head" >
              Quickshare
            </div>
            <div id="win-head-menu"className="win-head-menu">
              <AuthPane authed={this.props.authPane.authed} />
            </div>
            <div id="win-body"className="win-body">
              <Browser
                dirPath={this.props.browser.dirPath}
                items={this.props.browser.items}
                update={this.update}
                uploadFiles={this.props.browser.uploadFiles}
                uploadValue={this.props.browser.uploadValue}
              />
            </div>
          </div>
        </div>
      </div>
    );
  }
}
