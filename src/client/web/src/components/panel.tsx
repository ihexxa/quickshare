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
      <div className="theme-white desktop">
        <div id="bg" className="bg bg-img font-m">
          <div
            id="top-bar"
            className="top-bar cyan1-font padding-t-m padding-b-m padding-l-l padding-r-l"
          >
            <div className="flex-2col-parent">
              <a href="https://github.com/ihexxa/quickshare" className="flex-13col h5">Quickshare</a>
              <span className="flex-23col text-right">
                <AuthPane
                  authed={this.props.authPane.authed}
                  update={this.update}                  
                />
              </span>
            </div>
          </div>

          <div id="container-center">
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
    );
  }
}
