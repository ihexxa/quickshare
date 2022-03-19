import * as React from "react";
import { Map } from "immutable";

import { ICoreState, MsgProps, UIProps } from "./core_state";
import { IconProps } from "./visual/icons";

import { PaneSettings } from "./pane_settings";
import { AdminPane, AdminProps } from "./pane_admin";

import { Tabs } from "./control/tabs";
import { Container } from "./layout/container";
import { LoginProps } from "./pane_login";
import { roleAdmin } from "../client";
import { settingsTabsCtrl } from "../common/controls";
import { Title } from "./visual/title";

export interface Props {
  admin: AdminProps;
  login: LoginProps;
  msg: MsgProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}
export class SettingsDialog extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  render() {
    const displaying = this.props.ui.control.controls.get(settingsTabsCtrl);
    const showSettings = displaying === "preferencePane" ? "" : "hidden";
    const showManagement =
      this.props.login.userRole === roleAdmin && displaying === "managementPane"
        ? ""
        : "hidden";

    return (
      <div id="settings-dialog">
        <Container>
          <Tabs
            targetControl={settingsTabsCtrl}
            tabIcons={Map<string, IconProps>({
              preferencePane: {
                name: "RiSettings3Fill",
                size: "1.6rem",
                color: "cyan1",
              },
              managementPane: {
                name: "RiWindowFill",
                size: "1.6rem",
                color: "cyan1",
              },
            })}
            ui={this.props.ui}
            msg={this.props.msg}
            update={this.props.update}
          />
        </Container>

        <div className={`${showSettings}`}>
          <PaneSettings
            login={this.props.login}
            msg={this.props.msg}
            update={this.props.update}
          />
        </div>

        <div className={`${showManagement}`}>
          <AdminPane
            admin={this.props.admin}
            ui={this.props.ui}
            msg={this.props.msg}
            update={this.props.update}
          />
        </div>
      </div>
    );
  }
}
