import * as React from "react";
import { List } from "immutable";
import { alertMsg, confirmMsg } from "../common/env";

import {
  ICoreState,
  MsgProps,
  UIProps,
  ctrlOn,
  ctrlHidden,
} from "./core_state";
import { LoginProps } from "./pane_login";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";
import { settingsDialogCtrl } from "./layers";

export interface State {}
export interface Props {
  login: LoginProps;
  msg: MsgProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class TopBar extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  openSettings = () => {
    updater().setControlOption(settingsDialogCtrl, ctrlOn);
    this.props.update(updater().updateUI);
  };

  logout = async (): Promise<void> => {
    if (!confirmMsg(this.props.msg.pkg.get("logout.confirm"))) {
      return;
    }

    return updater()
      .logout()
      .then((status: string) => {
        if (status === "") {
          const params = new URLSearchParams(
            document.location.search.substring(1)
          );
          return updater().initAll(params);
        } else {
          alertMsg(this.props.msg.pkg.get("login.logout.fail"));
        }
      })
      .then(() => {
        return this.refreshCaptcha();
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateUploadingsInfo);
        this.props.update(updater().updateSharingsInfo);
        this.props.update(updater().updateLogin);
        this.props.update(updater().updateAdmin);
        this.props.update(updater().updateUI);
        this.props.update(updater().updateMsg);
      });
  };

  refreshCaptcha = async () => {
    return updater()
      .getCaptchaID()
      .then(() => {
        this.props.update(updater().updateLogin);
      });
  };

  render() {
    const showLogin = this.props.login.authed ? "" : "hidden";
    const showSettings =
      this.props.ui.control.controls.get(settingsDialogCtrl) === ctrlHidden
        ? "hidden"
        : "";

    return (
      <div id="top-bar">
        <Flexbox
          children={List([
            <a
              id="topbar-title"
              href="https://github.com/ihexxa/quickshare"
              target="_blank"
              className="h5"
            >
              Quickshare
            </a>,

            <Flexbox
              children={List([
                <button
                  onClick={this.openSettings}
                  className={`margin-r-m ${showSettings}`}
                >
                  {this.props.msg.pkg.get("settings")}
                  {/* {getIcon("RiSettings4Line", "1.8rem", "cyan0")} */}
                </button>,

                <button onClick={this.logout} className={`${showLogin}`}>
                  {this.props.msg.pkg.get("login.logout")}
                </button>,
              ])}
              childrenStyles={List([{}, {}, {}, {}])}
            />,
          ])}
          childrenStyles={List([
            {},
            { justifyContent: "flex-end", alignItems: "center" },
          ])}
        />
      </div>
    );
  }
}
