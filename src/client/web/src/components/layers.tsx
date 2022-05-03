import * as React from "react";
import { throttle } from "throttle-debounce";

import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { AdminProps } from "./pane_admin";
import { SettingsDialog } from "./dialog_settings";

import { AuthPane, LoginProps } from "./pane_login";
import { FilesProps } from "./panel_files";
import { Container } from "./layout/container";
import {
  settingsDialogCtrl,
  ctrlOff,
  sharingCtrl,
  loadingCtrl,
  ctrlOn,
  ctrlHidden,
  dropAreaCtrl,
} from "../common/controls";
import { LoadingIcon } from "./visual/loading";
import { Title } from "./visual/title";
import { HotkeyHandler } from "../common/hotkeys";
import { getIcon } from "./visual/icons";

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
    this.hotkeyHandler.add({ key: "Escape" }, this.closeHandler);

    document.addEventListener("keyup", this.hotkeyHandler.handle);
  }

  componentWillUnmount() {
    document.removeEventListener("keyup", this.hotkeyHandler.handle);
  }

  closeHandler = () => {
    if (this.props.ui.control.controls.get(settingsDialogCtrl) === ctrlOn) {
      this.setControlOption(settingsDialogCtrl, ctrlOff);
    }
  };

  setControlOption = (targetControl: string, option: string) => {
    updater().setControlOption(targetControl, option);
    this.props.update(updater().updateUI);
  };

  render() {
    const hideLogin =
      this.props.login.authed ||
      (this.props.ui.control.controls.get(sharingCtrl) === ctrlOn &&
        this.props.filesInfo.isSharing);
    const loginPaneClass = hideLogin ? "hidden" : "";
    const dropAreaClass =
      this.props.ui.control.controls.get(dropAreaCtrl) === ctrlOn
        ? ""
        : "hidden";

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

        <div id="login-layer" className={`layer ${loginPaneClass}`}>
          <AuthPane
            login={this.props.login}
            ui={this.props.ui}
            update={this.props.update}
            msg={this.props.msg}
            enabled={!hideLogin}
          />
        </div>

        <div id="drop-area-layer" className={`${dropAreaClass}`}>
          <div className="drop-area-container">
            <div className="drop-area major-bg focus-font">
              <div>{getIcon("RiFolderUploadFill", "4rem", "focus")}</div>
              <span>{this.props.msg.pkg.get("term.dropAnywhere")}</span>
            </div>
          </div>
        </div>

        <div id="settings-layer" className={`layer ${showSettings}`}>
          <div id="root-container">
            <Container>
              <div className="col-l">
                <Title
                  title={this.props.msg.pkg.get("pane.settings")}
                  iconColor="major"
                  iconName="RiListSettingsFill"
                />
              </div>
              <div className="col-r">
                <button
                  className="button-default"
                  onClick={() => {
                    this.setControlOption(settingsDialogCtrl, ctrlOff);
                  }}
                >
                  {this.props.msg.pkg.get("panes.close")}
                </button>
              </div>
              <div className="fix"></div>
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
