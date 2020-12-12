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
        <div id="bg" className="bg font-m">
          <div id="panel" className="panel">
            <div id="panel-head" className="panel-head cyan1-font">
              <div className="flex-2col-parent">
                <div className="flex-2col">Quickshare</div>
                <div className="flex-2col text-right">
                  <AuthPane
                    authed={this.props.authPane.authed}
                    update={this.update}
                  />
                </div>
              </div>
            </div>
            {/* <div id="panel-head-menu" className="panel-head-menu"></div> */}
            <div id="panel-body" className="panel-body">
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
