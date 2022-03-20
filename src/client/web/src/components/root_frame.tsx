import * as React from "react";
import { Map } from "immutable";

import { ICoreState, MsgProps, UIProps } from "./core_state";
import { FilesPanel, FilesProps } from "./panel_files";
import { UploadingsPanel, UploadingsProps } from "./panel_uploadings";
import { SharingsPanel, SharingsProps } from "./panel_sharings";
import { IconProps, iconSize } from "./visual/icons";
import { Tabs } from "./control/tabs";
import { LoginProps } from "./pane_login";
import { Layers } from "./layers";
import { AdminProps } from "./pane_admin";
import { TopBar } from "./topbar";

export const controlName = "panelTabs";
export interface Props {
  filesInfo: FilesProps;
  uploadingsInfo: UploadingsProps;
  sharingsInfo: SharingsProps;
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

  makeBgStyle = (): Object => {
    if (
      this.props.login.preferences != null &&
      this.props.login.preferences.bg.url !== ""
    ) {
      const bgConfig = this.props.login.preferences.bg;
      return {
        background: `url("${bgConfig.url}") ${bgConfig.repeat} ${bgConfig.position} ${bgConfig.align}`,
      };
    }

    if (this.props.login.preferences.bg.bgColor !== "") {
      return {
        backgroundColor: this.props.login.preferences.bg.bgColor,
      };
    }

    if (this.props.ui.bg.url !== "") {
      return {
        background: `url("${this.props.ui.bg.url}") ${this.props.ui.bg.repeat} ${this.props.ui.bg.position} ${this.props.ui.bg.align}`,
      };
    }

    if (this.props.ui.bg.bgColor !== "") {
      return {
        backgroundColor: this.props.ui.bg.bgColor,
      };
    }

    return {};
  };

  render() {
    const bgStyle = this.makeBgStyle();
    const theme =
      this.props.login.preferences.theme === "light"
        ? "theme-default"
        : "theme-dark";
    const fontSizeClass = "font-m";

    const displaying = this.props.ui.control.controls.get(controlName);
    const filesPanelClass = displaying === "filesPanel" ? "" : "hidden";
    const uploadingsPanelClass =
      displaying === "uploadingsPanel" ? "" : "hidden";
    const sharingsPanelClass = displaying === "sharingsPanel" ? "" : "hidden";

    return (
      <div id="root-frame" className={`${theme} ${fontSizeClass}`}>
        <div id="bg" style={bgStyle}>
          <Layers
            login={this.props.login}
            admin={this.props.admin}
            ui={this.props.ui}
            msg={this.props.msg}
            filesInfo={this.props.filesInfo}
            update={this.props.update}
          />

          <TopBar
            login={this.props.login}
            msg={this.props.msg}
            ui={this.props.ui}
            update={this.props.update}
          />

          <div id="top-menu">
            <Tabs
              targetControl={controlName}
              tabIcons={Map<string, IconProps>({
                filesPanel: {
                  name: "RiFolder2Fill",
                  size: iconSize("s"),
                  color: "focus",
                },
                uploadingsPanel: {
                  name: "RiUploadCloudFill",
                  size: iconSize("s"),
                  color: "focus",
                },
                sharingsPanel: {
                  name: "RiShareBoxLine",
                  size: iconSize("s"),
                  color: "focus",
                },
              })}
              ui={this.props.ui}
              msg={this.props.msg}
              update={this.props.update}
            />
          </div>

          <div className="container-center">
            <span className={filesPanelClass}>
              <FilesPanel
                filesInfo={this.props.filesInfo}
                msg={this.props.msg}
                login={this.props.login}
                ui={this.props.ui}
                enabled={displaying === "filesPanel"}
                update={this.props.update}
              />
            </span>

            <span className={uploadingsPanelClass}>
              <UploadingsPanel
                uploadingsInfo={this.props.uploadingsInfo}
                msg={this.props.msg}
                login={this.props.login}
                ui={this.props.ui}
                update={this.props.update}
              />
            </span>

            <span className={sharingsPanelClass}>
              <SharingsPanel
                sharingsInfo={this.props.sharingsInfo}
                msg={this.props.msg}
                login={this.props.login}
                ui={this.props.ui}
                update={this.props.update}
              />
            </span>
          </div>

          <div id="tail" className="container-center">
            <a href="https://github.com/ihexxa/quickshare">Quickshare</a> -
            quick and simple file sharing.
          </div>
        </div>
      </div>
    );
  }
}
