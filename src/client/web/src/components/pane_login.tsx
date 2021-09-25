import * as React from "react";
import { Set } from "immutable";

import { ICoreState, MsgProps } from "./core_state";
import { updater } from "./state_updater";
import { alertMsg } from "../common/env";
import { Quota } from "../client";

export interface LoginProps {
  userID: string;
  userName: string;
  userRole: string;
  usedSpace: string;
  quota: Quota;
  authed: boolean;
  captchaID: string;
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
      .then((ok: boolean): Promise<any> => {
        if (ok) {
          const params = new URLSearchParams(document.location.search.substring(1));
          return updater().initAll(params);
        } else {
          this.setState({ user: "", pwd: "", captchaInput: "" });
          alertMsg("Failed to login.");
          return updater().getCaptchaID();
        }
      })
      .then(() => {
        this.update(updater().updateBrowser);
        this.update(updater().updateLogin);
        this.update(updater().updatePanes);
        this.update(updater().updateAdmin);

        updater().initLan();
        this.update(updater().updateMsg);
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
    return (
      <span>
        <div
          className="container"
          style={{ display: this.props.login.authed ? "none" : "block" }}
        >
          <div className="padding-l">
            <div className="flex-list-container">
              <div className="flex-list-item-l">
                <input
                  name="user"
                  type="text"
                  onChange={this.changeUser}
                  value={this.state.user}
                  className="black0-font margin-t-m margin-b-m margin-r-m"
                  placeholder={this.props.msg.pkg.get("login.username")}
                />
                <input
                  name="pwd"
                  type="password"
                  onChange={this.changePwd}
                  value={this.state.pwd}
                  className="black0-font margin-t-m margin-b-m"
                  placeholder={this.props.msg.pkg.get("login.pwd")}
                />
              </div>
              <div className="flex-list-item-r">
                <button
                  onClick={this.login}
                  className="green0-bg white-font margin-t-m margin-b-m"
                >
                  {this.props.msg.pkg.get("login.login")}
                </button>
              </div>
            </div>

            <div className="flex-list-container">
              <div className="flex-list-item-l">
                <input
                  name="captcha"
                  type="text"
                  onChange={this.changeCaptcha}
                  value={this.state.captchaInput}
                  className="black0-font margin-t-m margin-b-m margin-r-m"
                  placeholder={this.props.msg.pkg.get("login.captcha")}
                />
                <img
                  src={`/v1/captchas/imgs?capid=${this.props.login.captchaID}`}
                  className="captcha"
                  onClick={this.refreshCaptcha}
                />
              </div>
              <div className="flex-list-item-l"></div>
            </div>
          </div>
        </div>
      </span>
    );
  }
}
