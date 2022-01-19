import * as React from "react";
import { List } from "immutable";

import { ICoreState, MsgProps } from "./core_state";
import { Flexbox } from "./layout/flexbox";
import { updater } from "./state_updater";
import { alertMsg } from "../common/env";
import { Quota, Preferences } from "../client";
import { getErrMsg } from "../common/utils";

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
    return updater()
      .login(
        this.state.user,
        this.state.pwd,
        this.props.login.captchaID,
        this.state.captchaInput
      )
      .then((status: string): Promise<any> => {
        this.setState({ captchaInput: "" });
        if (status === "") {
          const params = new URLSearchParams(
            document.location.search.substring(1)
          );
          return updater().initAll(params);
        } else {
          throw status;
        }
      })
      .then((status: string) => {
        if (status !== "") {
          throw status;
        }
        this.update(updater().updateAll);
      })
      .catch((status: Error) => {
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status.toString()));
        return updater().getCaptchaID();
      });
  };

  refreshCaptcha = async () => {
    return updater()
      .getCaptchaID()
      .then((status: string) => {
        if (status !== "") {
          alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        } else {
          this.props.update(updater().updateLogin);
        }
      });
  };

  render() {
    return (
      <div
        id="pane-login"
        className="container"
        style={{ display: this.props.login.authed ? "none" : "block" }}
      >
        <span className="float-input">
          <div className="label">
            {this.props.msg.pkg.get("login.username")}
          </div>
          <input
            name="user"
            type="text"
            onChange={this.changeUser}
            value={this.state.user}
            placeholder={this.props.msg.pkg.get("login.username")}
          />
        </span>

        <span className="float-input">
          <div className="label">{this.props.msg.pkg.get("login.pwd")}</div>
          <input
            name="pwd"
            type="password"
            onChange={this.changePwd}
            value={this.state.pwd}
            placeholder={this.props.msg.pkg.get("login.pwd")}
          />
        </span>

        <span className="float-input">
          <div className="label">{this.props.msg.pkg.get("login.captcha")}</div>
          <Flexbox
            children={List([
              <input
                name="captcha"
                type="text"
                onChange={this.changeCaptcha}
                value={this.state.captchaInput}
                placeholder={this.props.msg.pkg.get("login.captcha")}
              />,
              <img
                src={`/v1/captchas/imgs?capid=${this.props.login.captchaID}`}
                className="captcha"
                onClick={this.refreshCaptcha}
              />,
            ])}
          />
        </span>

        <span className="float-input">
          <button id="btn-login" onClick={this.login}>
            {this.props.msg.pkg.get("login.login")}
          </button>
        </span>
      </div>
    );
  }
}
