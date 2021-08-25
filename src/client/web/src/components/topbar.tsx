import * as React from "react";

import { ICoreState } from "./core_state";
import { updater } from "./state_updater";

export interface State {}
export interface Props {
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class TopBar extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  showSettings = () => {
    updater().displayPane("settings");
    this.props.update(updater().updatePanes);
  };

  showAdmin = async () => {
    return updater()
      .self()
      .then(() => {
        // TODO: use props instead
        // TODO: remove hardcode role
        if (
          updater().props.login.authed &&
          updater().props.panes.userRole === "admin"
        ) {
          return Promise.all([updater().listRoles(), updater().listUsers()]);
        }
      })
      .then(() => {
        updater().displayPane("admin");
        this.props.update(updater().updateAdmin);
        this.props.update(updater().updatePanes);
      });
  };

  render() {
    return (
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
    );
  }
}
