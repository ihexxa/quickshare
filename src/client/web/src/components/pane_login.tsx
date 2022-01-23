import * as React from "react";
import { List } from "immutable";

import { ICoreState, MsgProps } from "./core_state";
import { Flexbox } from "./layout/flexbox";
import { updater } from "./state_updater";
import { alertMsg } from "../common/env";
import { Quota, Preferences } from "../client";
import { getErrMsg } from "../common/utils";
import { ctrlOn, ctrlOff, loadingCtrl } from "../common/controls";

export interface ExtInfo {
  usedSpace: string;
}
export interface LoginProps {
  authed: boolean;
  captchaID: string;
  userID: string;
  userName: string;
  userRole: string;
  extInfo: ExtInfo;
  quota: Quota;
  preferences: Preferences;
}

export interface Props {
  login: LoginProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {
  user: string;
  pwd: string;
  captchaInput: string;
}

export class AuthPane extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  constructor(p: Props) {
    super(p);
    this.update = p.update;
    this.state = {
      user: "",
      pwd: "",
      captchaInput: "",
    };
  }

  changeUser = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: ev.target.value });
  };

  changePwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ pwd: ev.target.value });
  };

  changeCaptcha = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ captchaInput: ev.target.value });
  };

  login = async () => {
    updater().setControlOption(loadingCtrl, ctrlOn);
    this.props.update(updater().updateUI);

    try {
      const loginStatus = await updater().login(
        this.state.user,
        this.state.pwd,
        this.props.login.captchaID,
        this.state.captchaInput
      );
      if (loginStatus !== "") {
        alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", loginStatus.toString())
        );
        return;
      }

      const params = new URLSearchParams(document.location.search.substring(1));
      const initStatus = await updater().initAll(params);
      if (initStatus !== "") {
        alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", initStatus.toString())
        );
      }

      this.setState({ user: "", pwd: "" });
    } finally {
      this.setState({ pwd: "", captchaInput: "" });
      updater().setControlOption(loadingCtrl, ctrlOff);
      await this.refreshCaptcha();
      this.props.update(updater().updateAll);
    }
  };

  refreshCaptcha = async () => {
    const status = await updater().getCaptchaID();
    if (status !== "") {
      alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
    } else {
      this.props.update(updater().updateLogin);
    }
  };

  render() {
    return (
      <div
        id="pane-login"
        className="container"
        style={{ display: this.props.login.authed ? "none" : "block" }}
      >
        <div className="login-container">
          <a
            href="https://github.com/ihexxa/quickshare"
            target="_blank"
            className="h5"
            id="title"
          >
            Quickshare
          </a>

          <div className="hr"></div>

          <div className="input-wrap">
            <input
              name="user"
              type="text"
              onChange={this.changeUser}
              value={this.state.user}
              placeholder={this.props.msg.pkg.get("login.username")}
            />
          </div>

          <div className="input-wrap">
            <input
              name="pwd"
              type="password"
              onChange={this.changePwd}
              value={this.state.pwd}
              placeholder={this.props.msg.pkg.get("login.pwd")}
            />
          </div>

          <Flexbox
            children={List([
              <div className="input-wrap">
                <input
                  id="captcha-input"
                  type="text"
                  onChange={this.changeCaptcha}
                  value={this.state.captchaInput}
                  placeholder={this.props.msg.pkg.get("login.captcha")}
                />
              </div>,
              <img
                id="captcha"
                src={`/v1/captchas/imgs?capid=${this.props.login.captchaID}`}
                className="captcha"
                onClick={this.refreshCaptcha}
              />,
            ])}
            childrenStyles={List([
              { justifyContent: "flex-start" },
              { justifyContent: "flex-end" },
            ])}
          />

          <button id="btn-login" onClick={this.login}>
            {this.props.msg.pkg.get("login.login")}
          </button>
        </div>
      </div>
    );
  }
}
