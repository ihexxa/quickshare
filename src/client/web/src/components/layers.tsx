import * as React from "react";
import { List } from "immutable";

import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { AdminProps } from "./pane_admin";
import { SettingsDialog } from "./dialog_settings";

import { AuthPane, LoginProps } from "./pane_login";
import { FilesProps } from "./panel_files";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { sharingCtrl, loadingCtrl, ctrlOn } from "../common/controls";
import { LoadingIcon } from "./visual/loading";

export interface Props {
  filesInfo: FilesProps;
  login: LoginProps;
  admin: AdminProps;
  ui: UIProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}
export class Layers extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  setControlOption = (targetControl: string, option: string) => {
    updater().setControlOption(targetControl, option);
    this.props.update(updater().updateUI);
  };

  render() {
    const showLogin =
      this.props.login.authed ||
      (this.props.ui.control.controls.get(sharingCtrl) === ctrlOn &&
        this.props.filesInfo.isSharing)
        ? "hidden"
        : "";
    const showSettings =
      this.props.ui.control.controls.get("settingsDialog") === ctrlOn
        ? ""
        : "hidden";
    const showLoading =
      this.props.ui.control.controls.get(loadingCtrl) == ctrlOn ? "" : "hidden";

    return (
      <div id="layers">
        <div id="loading-layer" className={showLoading}>
          <LoadingIcon />
        </div>

        <div id="login-layer" className={`layer ${showLogin}`}>
          <div id="root-container">
            <AuthPane
              login={this.props.login}
              update={this.props.update}
              msg={this.props.msg}
            />
          </div>
        </div>

        <div id="settings-layer" className={`layer ${showSettings}`}>
          <div id="root-container">
            <Container>
              <Flexbox
                children={List([
                  <h4 id="title">{this.props.msg.pkg.get("pane.settings")}</h4>,
                  <button
                    onClick={() => {
                      this.setControlOption("settingsDialog", "off");
                    }}
                  >
                    {this.props.msg.pkg.get("panes.close")}
                  </button>,
                ])}
                childrenStyles={List([{}, { justifyContent: "flex-end" }])}
              />
            </Container>

            <SettingsDialog
              admin={this.props.admin}
              login={this.props.login}
              msg={this.props.msg}
              ui={this.props.ui}
              update={this.props.update}
            />
          </div>
        </div>
      </div>
    );
  }
}
