import * as React from "react";

import { ICoreState } from "./core_state";
import { Browser, Props as BrowserProps } from "./browser";
import { AuthPane, Props as AuthPaneProps } from "./pane_login";
import { Panes, Props as PanesProps, Updater as PanesUpdater } from "./panes";

export interface Props {
  displaying: string;
  browser: BrowserProps;
  authPane: AuthPaneProps;
  panes: PanesProps;
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

  showSettings = () => {
    PanesUpdater.displayPane("settings");
    this.update(PanesUpdater.updateState);
  };

  render() {
    return (
      <div className="theme-white desktop">
        <div id="bg" className="bg bg-img font-m">
          <Panes
            displaying={this.props.panes.displaying}
            paneNames={this.props.panes.paneNames}
            login={this.props.authPane}
            update={this.update}
          />

          <div
            id="top-bar"
            className="top-bar cyan1-font padding-t-m padding-b-m padding-l-l padding-r-l"
          >
            <div className="flex-2col-parent">
              <a
                href="https://github.com/ihexxa/quickshare"
                className="flex-13col h5"
              >
                Quickshare
              </a>
              <span className="flex-23col text-right">
                <button
                  onClick={this.showSettings}
                  className="grey1-bg white-font margin-r-m"
                >
                  Settings
                </button>
              </span>
            </div>
          </div>

          <div className="container-center">
            <Browser
              dirPath={this.props.browser.dirPath}
              items={this.props.browser.items}
              uploadings={this.props.browser.uploadings}
              update={this.update}
              uploadFiles={this.props.browser.uploadFiles}
              uploadValue={this.props.browser.uploadValue}
              isVertical={this.props.browser.isVertical}
            />
          </div>
          <div className="container-center black0-font tail margin-t-xl margin-b-xl">
            <a href="https://github.com/ihexxa/quickshare">Quickshare</a> -
            sharing in simple way.
          </div>
        </div>
      </div>
    );
  }
}
