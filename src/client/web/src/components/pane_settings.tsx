import * as React from "react";
import FileSize from "filesize";
import { List } from "immutable";

import { ICoreState, UIProps, MsgProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { Flexbox } from "./layout/flexbox";
import { Flowgrid } from "./layout/flowgrid";
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
  changeBgUrl = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setPreferences({
      ...this.props.login.preferences,
      bg: { ...this.props.login.preferences.bg, url: ev.target.value },
    });
    this.props.update(updater().updateLogin);
  };
  changeBgRepeat = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setPreferences({
      ...this.props.login.preferences,
      bg: { ...this.props.login.preferences.bg, repeat: ev.target.value },
    });
    this.props.update(updater().updateLogin);
  };
  changeBgPos = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setPreferences({
      ...this.props.login.preferences,
      bg: { ...this.props.login.preferences.bg, position: ev.target.value },
    });
    this.props.update(updater().updateLogin);
  };
  changeBgAlign = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setPreferences({
      ...this.props.login.preferences,
      bg: { ...this.props.login.preferences.bg, align: ev.target.value },
    });
    this.props.update(updater().updateLogin);
  };
  changeCSSURL = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setPreferences({
      ...this.props.login.preferences,
      cssURL: ev.target.value,
    });
    this.props.update(updater().updateLogin);
  };
  changeLanPackURL = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setPreferences({
      ...this.props.login.preferences,
      lanPackURL: ev.target.value,
    });
    this.props.update(updater().updateLogin);
  };

  constructor(p: Props) {
    super(p);
    this.state = {
      oldPwd: "",
      newPwd1: "",
      newPwd2: "",
    };
  }

  syncPreferences = async () => {
    updater()
      .syncPreferences()
      .then((status: number) => {
        if (status === 200) {
          alertMsg(this.props.msg.pkg.get("update.ok"));
        } else {
          alertMsg(this.props.msg.pkg.get("update.fail"));
        }
      });
  };

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

          <div className="hr white0-bg margin-t-m margin-b-m"></div>
          <div>
            <Flexbox
              children={List([
                <span className="title-m bold">
                  {this.props.msg.pkg.get("settings.customLan")}
                </span>,

                <span>
                  <button onClick={this.syncPreferences}>
                    {this.props.msg.pkg.get("update")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            <Flowgrid
              grids={List([
                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-size-s grey1-font">
                    {this.props.msg.pkg.get("settings.lanPackURL")}
                  </div>
                  <input
                    type="text"
                    onChange={this.changeLanPackURL}
                    value={this.props.login.preferences.lanPackURL}
                    className="black0-font"
                    style={{ width: "20rem" }}
                    placeholder={this.props.msg.pkg.get("settings.lanPackURL")}
                  />
                </div>,
              ])}
            />
          </div>

          <div className="hr white0-bg margin-t-m margin-b-m"></div>
          <div>
            <Flexbox
              children={List([
                <span className="title-m bold">
                  {this.props.msg.pkg.get("cfg.bg")}
                </span>,

                <span>
                  <button onClick={this.syncPreferences}>
                    {this.props.msg.pkg.get("update")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            <Flowgrid
              grids={List([
                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-size-s grey1-font">
                    {this.props.msg.pkg.get("cfg.bg.url")}
                  </div>
                  <input
                    name="bg_url"
                    type="text"
                    onChange={this.changeBgUrl}
                    value={this.props.login.preferences.bg.url}
                    className="black0-font"
                    style={{ width: "20rem" }}
                    placeholder={this.props.msg.pkg.get("cfg.bg.url")}
                  />
                </div>,

                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-size-s grey1-font">
                    {this.props.msg.pkg.get("cfg.bg.repeat")}
                  </div>
                  <input
                    name="bg_repeat"
                    type="text"
                    onChange={this.changeBgRepeat}
                    value={this.props.login.preferences.bg.repeat}
                    className="black0-font"
                    placeholder={this.props.msg.pkg.get("cfg.bg.repeat")}
                  />
                </div>,

                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-size-s grey1-font">
                    {this.props.msg.pkg.get("cfg.bg.pos")}
                  </div>
                  <input
                    name="bg_pos"
                    type="text"
                    onChange={this.changeBgPos}
                    value={this.props.login.preferences.bg.position}
                    className="black0-font"
                    placeholder={this.props.msg.pkg.get("cfg.bg.pos")}
                  />
                </div>,

                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-size-s grey1-font">
                    {this.props.msg.pkg.get("cfg.bg.align")}
                  </div>
                  <input
                    name="bg_align"
                    type="text"
                    onChange={this.changeBgAlign}
                    value={this.props.login.preferences.bg.align}
                    className="black0-font"
                    placeholder={this.props.msg.pkg.get("cfg.bg.align")}
                  />
                </div>,
              ])}
            />
          </div>

          <div className="hr white0-bg margin-t-m margin-b-m"></div>
          <div>
            <Flexbox
              children={List([
                <span className="title-m bold">
                  {this.props.msg.pkg.get("prefer.theme")}
                </span>,

                <span>
                  <button onClick={this.syncPreferences}>
                    {this.props.msg.pkg.get("update")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            <Flowgrid
              grids={List([
                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-size-s grey1-font">
                    {this.props.msg.pkg.get("prefer.theme.url")}
                  </div>
                  <input
                    type="text"
                    onChange={this.changeCSSURL}
                    value={this.props.login.preferences.cssURL}
                    className="black0-font"
                    style={{ width: "20rem" }}
                    placeholder={this.props.msg.pkg.get("prefer.theme.url")}
                  />
                </div>,
              ])}
            />
          </div>

        </div>
      </div>
    );
  }
}
