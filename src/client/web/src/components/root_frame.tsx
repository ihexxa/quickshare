import * as React from "react";

import { ICoreState, BaseUpdater } from "./core_state";
import { Browser, Props as BrowserProps } from "./browser";
import { Props as PaneLoginProps } from "./pane_login";
import { Panes, Props as PanesProps, Updater as PanesUpdater } from "./panes";

export interface Props {
  displaying: string;
  browser: BrowserProps;
  authPane: PaneLoginProps;
  panes: PanesProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  public static props: Props;
  public static init = (props: Props) => (BaseUpdater.props = { ...props });
  public static apply = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panel: { ...prevState.panel, ...Updater.props },
    };
  };
}

export interface State {}
export class RootFrame extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    Updater.init(p);
  }

  showSettings = () => {
    PanesUpdater.displayPane("settings");
    this.props.update(PanesUpdater.updateState);
  };

  showAdmin = () => {
    PanesUpdater.displayPane("admin");
    this.props.update(PanesUpdater.updateState);
  };

  render() {
    const update = this.props.update;
    return (
      <div className="theme-white desktop">
        <div id="bg" className="bg bg-img font-m">
          <Panes
            userRole={this.props.panes.userRole}
            displaying={this.props.panes.displaying}
            paneNames={this.props.panes.paneNames}
            login={this.props.authPane}
            admin={this.props.panes.admin}
            update={update}
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
                <button
                  onClick={this.showAdmin}
                  className="grey1-bg white-font margin-r-m"
                >
                  Admin
                </button>
              </span>
            </div>
          </div>

          <div className="container-center">
            <Browser
              dirPath={this.props.browser.dirPath}
              items={this.props.browser.items}
              uploadings={this.props.browser.uploadings}
              update={update}
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
