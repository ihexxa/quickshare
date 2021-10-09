import * as React from "react";
import FileSize from "filesize";
import { List } from "immutable";

import { ICoreState, MsgProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { Flexbox } from "./layout/flexbox";
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
    this.state = {
      oldPwd: "",
      newPwd1: "",
      newPwd2: "",
    };
  }

  setPwd = async (): Promise<any> => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
    } else if (
      this.state.oldPwd == "" ||
      this.state.newPwd1 == "" ||
      this.state.newPwd2 == ""
    ) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.empty"));
    } else if (this.state.oldPwd == this.state.newPwd1) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notChanged"));
    } else {
      return updater()
        .setPwd(this.state.oldPwd, this.state.newPwd1)
        .then((ok: boolean) => {
          if (ok) {
            alertMsg(this.props.msg.pkg.get("update.ok"));
          } else {
            alertMsg(this.props.msg.pkg.get("update.fail"));
          }
          this.setState({
            oldPwd: "",
            newPwd1: "",
            newPwd2: "",
          });
        });
    }
  };

  setLan = (lan: string) => {
    updater().setLan(lan);
    this.props.update(updater().updateMsg);
  };

  render() {
    return (
      <div className="container">
        <div className="padding-l">
          <div className="grey3-font">
            <h5 className="black-font">
              {this.props.msg.pkg.get("user.profile")}
            </h5>

            <div className="font-size-s margin-t-s">
              <span className="grey0-font">
                {`${this.props.msg.pkg.get("user.name")}:`}{" "}
              </span>
              <span>{`${this.props.login.userName}`}</span>
            </div>
            <div className="font-size-s margin-t-s">
              <span className="grey0-font">
                {`${this.props.msg.pkg.get("user.role")}:`}{" "}
              </span>
              <span>{`${this.props.login.userRole}`}</span>
            </div>
            <div className="font-size-s margin-t-s">
              <span className="grey0-font">
                {`${this.props.msg.pkg.get("user.spaceLimit")}:`}{" "}
              </span>
              <span>
                {" "}
                {`${FileSize(parseInt(this.props.login.quota.spaceLimit, 10), {
                  round: 0,
                })}`}
              </span>
            </div>
            <div className="font-size-s margin-t-s">
              <span className="grey0-font">
                {`${this.props.msg.pkg.get("user.upLimit")}:`}{" "}
              </span>
              <span>
                {" "}
                {`${FileSize(this.props.login.quota.uploadSpeedLimit, {
                  round: 0,
                })}`}
              </span>
            </div>
            <div className="font-size-s margin-t-s">
              <span className="grey0-font">
                {`${this.props.msg.pkg.get("user.downLimit")}:`}{" "}
              </span>
              <span>
                {" "}
                {`${FileSize(this.props.login.quota.downloadSpeedLimit, {
                  round: 0,
                })}`}
              </span>
            </div>
          </div>
          <div className="hr white0-bg margin-t-m margin-b-m"></div>
          <div>
            <Flexbox
              children={List([
                <h5 className="black-font">
                  {this.props.msg.pkg.get("settings.pwd.update")}
                </h5>,
                <button onClick={this.setPwd}>
                  {this.props.msg.pkg.get("update")}
                </button>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />

            <span className="inline-block margin-r-m margin-b-m">
              <div className="font-size-s grey1-font">
                {this.props.msg.pkg.get("settings.pwd.old")}
              </div>
              <input
                name="old_pwd"
                type="password"
                onChange={this.changeOldPwd}
                value={this.state.oldPwd}
                className="black0-font"
                placeholder={this.props.msg.pkg.get("settings.pwd.old")}
              />
            </span>

            <span className="inline-block margin-r-m  margin-b-m">
              <div className="font-size-s grey1-font">
                {this.props.msg.pkg.get("settings.pwd.new1")}
              </div>
              <input
                name="new_pwd1"
                type="password"
                onChange={this.changeNewPwd1}
                value={this.state.newPwd1}
                className="black0-font"
                placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
              />
            </span>

            <span className="inline-block margin-r-m margin-b-m">
              <div className="font-size-s grey1-font">
                {this.props.msg.pkg.get("settings.pwd.new2")}
              </div>
              <input
                name="new_pwd2"
                type="password"
                onChange={this.changeNewPwd2}
                value={this.state.newPwd2}
                className="black0-font"
                placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
              />
            </span>
          </div>
          <div className="hr white0-bg margin-t-m margin-b-m"></div>
          <div className="margin-b-m">
            <Flexbox
              children={List([
                <h5 className="black-font">
                  {this.props.msg.pkg.get("settings.chooseLan")}
                </h5>,
                <span>
                  <button
                    onClick={() => {
                      this.setLan("en_US");
                    }}
                    className="margin-r-m"
                  >
                    {this.props.msg.pkg.get("enUS")}
                  </button>
                  <button
                    onClick={() => {
                      this.setLan("zh_CN");
                    }}
                  >
                    {this.props.msg.pkg.get("zhCN")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
          </div>
        </div>
      </div>
    );
  }
}
