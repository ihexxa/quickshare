import * as React from "react";
import { Map } from "immutable";

import { ICoreState, MsgProps, UIProps } from "./core_state";
import { FilesPanel, FilesProps } from "./panel_files";
import { UploadingsPanel, UploadingsProps } from "./panel_uploadings";
import { SharingsPanel, SharingsProps } from "./panel_sharings";
import { IconProps } from "./visual/icons";

import { PaneSettings } from "./pane_settings";
import { AdminPane, AdminProps } from "./pane_admin";

import { Tabs } from "./control/tabs";
import { LoginProps } from "./pane_login";
import { RiShareBoxLine } from "@react-icons/all-files/ri/RiShareBoxLine";
import { roleAdmin } from "../client";

export const settingsTabsCtrl = "settingsTabs";

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
    const showSettings = displaying === "settingsPane" ? "" : "hidden";
    const showManagement =
      this.props.login.userRole === roleAdmin && displaying === "managementPane"
        ? ""
        : "hidden";

    return (
      <div id="settings-dialog">
        <div className="container">
          <Tabs
            targetControl={settingsTabsCtrl}
            tabIcons={Map<string, IconProps>({
              settingsPane: {
                name: "RiSettings3Fill",
                size: "1.6rem",
                color: "cyan0",
              },
              managementPane: {
                name: "RiWindowFill",
                size: "1.6rem",
                color: "cyan0",
              },
            })}
            login={this.props.login}
            admin={this.props.admin}
            ui={this.props.ui}
            msg={this.props.msg}
            update={this.props.update}
          />
        </div>

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
