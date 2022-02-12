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
import {
  settingsDialogCtrl,
  ctrlOff,
  sharingCtrl,
  loadingCtrl,
  ctrlOn,
  ctrlHidden,
} from "../common/controls";
import { LoadingIcon } from "./visual/loading";
import { HotkeyHandler } from "../common/hotkeys";

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
  private hotkeyHandler: HotkeyHandler;
  constructor(p: Props) {
    super(p);
  }

  componentDidMount(): void {
    this.hotkeyHandler = new HotkeyHandler();

    const closeHandler = () => {
      this.setControlOption(settingsDialogCtrl, ctrlOff);
    };
    this.hotkeyHandler.add({ key: "Escape" }, closeHandler);

    document.addEventListener("keyup", this.hotkeyHandler.handle);
  }

  componentWillUnmount() {
    document.removeEventListener("keyup", this.hotkeyHandler.handle);
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
      this.props.ui.control.controls.get(settingsDialogCtrl) === ctrlOn
        ? ""
        : "hidden";
    const showLoading =
      this.props.ui.control.controls.get(loadingCtrl) == ctrlOn
        ? ""
        : ctrlHidden;

    return (
      <div id="layers">
        <div id="loading-layer" className={showLoading}>
          <LoadingIcon />
        </div>

        <div id="login-layer" className={`layer ${showLogin}`}>
          <div id="root-container">
            <AuthPane
              login={this.props.login}
              ui={this.props.ui}
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
                      this.setControlOption(settingsDialogCtrl, ctrlOff);
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
