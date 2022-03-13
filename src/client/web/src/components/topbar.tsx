import * as React from "react";
import { List } from "immutable";

import { alertMsg, confirmMsg } from "../common/env";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";
import { ctrlOn, ctrlHidden, settingsDialogCtrl } from "../common/controls";
import { QRCodeIcon } from "./visual/qrcode";

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

    const status = await updater().logout();
    if (status !== "") {
      alertMsg(this.props.msg.pkg.get("login.logout.fail"));
      return;
    }

    const params = new URLSearchParams(document.location.search.substring(1));
    const initStatus = await updater().initAll(params);
    if (initStatus !== "") {
      alertMsg(this.props.msg.pkg.get("op.fail"));
      return;
    }
    this.props.update(updater().updateAll);
  };

  render() {
    const loginPanelClass = this.props.login.authed ? "" : "hidden";
    const settingsPanelClass =
      this.props.ui.control.controls.get(settingsDialogCtrl) === ctrlHidden
        ? "hidden"
        : "";

    return (
      <div id="top-bar" className="highlight-font light-bg">
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
            <QRCodeIcon
              value={document.URL}
              size={128}
              pos={true}
              className="margin-l-m"
            />,

            <Flexbox
              children={List([
                <button
                  onClick={this.openSettings}
                  className={`margin-r-m ${settingsPanelClass}`}
                >
                  {this.props.msg.pkg.get("settings")}
                </button>,

                <button onClick={this.logout} className={`${loginPanelClass}`}>
                  {this.props.msg.pkg.get("login.logout")}
                </button>,
              ])}
              childrenStyles={List([
                { flex: "0 0 auto" },
                { flex: "0 0 auto" },
              ])}
            />,
          ])}
          childrenStyles={List([
            { flex: "0 0 auto" },
            { flex: "0 0 auto" },
            { justifyContent: "flex-end", alignItems: "center" },
          ])}
        />
      </div>
    );
  }
}
