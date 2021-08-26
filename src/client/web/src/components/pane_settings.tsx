import * as React from "react";

import { ICoreState, MsgProps } from "./core_state";
import { AuthPane, LoginProps } from "./pane_login";
import { updater } from "./state_updater";
import { alertMsg } from "../common/env";
export interface Props {
  login: LoginProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {
  oldPwd: string;
  newPwd1: string;
  newPwd2: string;
}

export class PaneSettings extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  changeOldPwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ oldPwd: ev.target.value });
  };
  changeNewPwd1 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd1: ev.target.value });
  };
  changeNewPwd2 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd2: ev.target.value });
  };

  constructor(p: Props) {
    super(p);
    this.update = p.update;
    this.state = {
      oldPwd: "",
      newPwd1: "",
      newPwd2: "",
    };
  }

  setPwd = () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
    } else if (this.state.newPwd1 == "") {
      alertMsg(this.props.msg.pkg.get("settings.pwd.empty"));
    } else if (this.state.oldPwd == this.state.newPwd1) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notChanged"));
    } else {
      updater()
        .setPwd(this.state.oldPwd, this.state.newPwd1)
        .then((ok: boolean) => {
          if (ok) {
            alertMsg(this.props.msg.pkg.get("settings.pwd.updated"));
          } else {
            alertMsg(this.props.msg.pkg.get("settings.pwd.fail"));
          }
          this.setState({
            oldPwd: "",
            newPwd1: "",
            newPwd2: "",
          });
        });
    }
  };

  render() {
    return (
      <div className="container">
        <div className="padding-l">
          <div>
            <div className="flex-list-container">
              <div className="flex-list-item-l">
                <h5 className="black-font">Update Password</h5>
              </div>
              <div className="flex-list-item-r">
                <button onClick={this.setPwd} className="grey1-bg white-font">
                  {this.props.msg.pkg.get("update")}
                </button>
              </div>
            </div>

            <div>
              <input
                name="old_pwd"
                type="password"
                onChange={this.changeOldPwd}
                value={this.state.oldPwd}
                className="black0-font margin-t-m margin-b-m"
                placeholder={this.props.msg.pkg.get("settings.pwd.old")}
              />
            </div>
            <div>
              <input
                name="new_pwd1"
                type="password"
                onChange={this.changeNewPwd1}
                value={this.state.newPwd1}
                className="black0-font margin-t-m margin-b-m margin-r-m"
                placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
              />
              <input
                name="new_pwd2"
                type="password"
                onChange={this.changeNewPwd2}
                value={this.state.newPwd2}
                className="black0-font margin-t-m margin-b-m"
                placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
              />
            </div>
          </div>

          <div className="hr white0-bg margin-t-m margin-b-m"></div>

          <div>
            <div className="flex-list-container">
              <div className="flex-list-item-l">
                <h5 className="black-font">Logout</h5>
              </div>
              <div className="flex-list-item-r">
                <AuthPane
                  login={this.props.login}
                  msg={this.props.msg}
                  update={this.update}
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }
}
