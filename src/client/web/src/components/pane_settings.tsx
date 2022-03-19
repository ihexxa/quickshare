import * as React from "react";
import FileSize from "filesize";
import { List } from "immutable";

import { ICoreState, UIProps, MsgProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { Flexbox } from "./layout/flexbox";
import { updater } from "./state_updater";
import { alertMsg, confirmMsg } from "../common/env";
import { Container } from "./layout/container";
import { Card } from "./layout/card";
import { Rows } from "./layout/rows";
import { ClientErrorV001, ErrorLogger } from "../common/log_error";
import { loadingCtrl, ctrlOn, ctrlOff } from "../common/controls";
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

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  syncPreferences = async () => {
    this.setLoading(true);
    try {
      const status = await updater().syncPreferences();
      if (status !== "") {
        alertMsg(this.props.msg.pkg.get("update.fail"));
        return;
      }
      alertMsg(this.props.msg.pkg.get("update.ok"));
    } finally {
      this.setLoading(false);
    }
  };

  setPwd = async (): Promise<any> => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
      return;
    } else if (
      this.state.oldPwd == "" ||
      this.state.newPwd1 == "" ||
      this.state.newPwd2 == ""
    ) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.empty"));
      return;
    } else if (this.state.oldPwd == this.state.newPwd1) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notChanged"));
      return;
    }

    this.setLoading(true);
    try {
      const status = await updater().setPwd(
        this.state.oldPwd,
        this.state.newPwd1
      );
      if (status !== "") {
        alertMsg(this.props.msg.pkg.get("update.fail"));
        return;
      }

      alertMsg(this.props.msg.pkg.get("update.ok"));
    } finally {
      this.setState({
        oldPwd: "",
        newPwd1: "",
        newPwd2: "",
      });
      this.setLoading(false);
    }
  };

  setLan = async (lan: string) => {
    updater().setLan(lan);
    try {
      const status = await updater().syncPreferences();
      if (status !== "") {
        alertMsg(this.props.msg.pkg.get("update.fail"));
      }
    } finally {
      alertMsg(this.props.msg.pkg.get("update.ok"));
      this.props.update(updater().updateMsg);
    }
  };

  truncateErrors = () => {
    if (confirmMsg(this.props.msg.pkg.get("op.confirm"))) {
      ErrorLogger().truncate();
    }
  };

  reportErrors = () => {
    if (confirmMsg(this.props.msg.pkg.get("op.confirm"))) {
      ErrorLogger().report();
    }
  };

  prepareErrorRows = (): List<React.ReactNode> => {
    let errRows = List<React.ReactNode>();

    ErrorLogger()
      .readErrs()
      .forEach((clientErr: ClientErrorV001, sign: string) => {
        const elem = (
          <div>
            <div className="error-inline">{JSON.stringify(clientErr)}</div>
            <div className="hr"></div>
          </div>
        );

        errRows = errRows.push(elem);
      });

    return errRows;
  };

  render() {
    const errRows = this.prepareErrorRows();
    const errorReportPane =
      errRows.size > 0 ? (
        <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">
                {this.props.msg.pkg.get("error.report.title")}
              </h5>,

              <span>
                <button
                  className="button-default margin-r-m"
                  onClick={this.reportErrors}
                >
                  {this.props.msg.pkg.get("op.submit")}
                </button>
                <button
                  className="button-default"
                  onClick={this.truncateErrors}
                >
                  {this.props.msg.pkg.get("op.truncate")}
                </button>
              </span>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="hr"></div>

          <Rows rows={errRows} />
        </Container>
      ) : null;

    return (
      <div id="pane-settings">
        <Container>
          <div id="profile">
            <h5 className="title-m">
              {this.props.msg.pkg.get("user.profile")}
            </h5>

            <div className="hr"></div>

            <div>
              <Card
                name={`${this.props.msg.pkg.get("user.name")}`}
                value={`${this.props.login.userName}`}
              />
              <Card
                name={`${this.props.msg.pkg.get("user.role")}`}
                value={`${this.props.login.userRole}`}
              />
              <Card
                name={`${this.props.msg.pkg.get("user.spaceLimit")}`}
                value={`${FileSize(
                  parseInt(this.props.login.quota.spaceLimit, 10),
                  {
                    round: 0,
                  }
                )}`}
              />
              <Card
                name={`${this.props.msg.pkg.get("user.upLimit")}`}
                value={`${FileSize(this.props.login.quota.uploadSpeedLimit, {
                  round: 0,
                })}`}
              />
              <Card
                name={`${this.props.msg.pkg.get("user.downLimit")}`}
                value={`${FileSize(this.props.login.quota.downloadSpeedLimit, {
                  round: 0,
                })}`}
              />
              <div className="fix"></div>
            </div>
          </div>
        </Container>

        <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">
                {this.props.msg.pkg.get("settings.pwd.update")}
              </h5>,
              <button className="button-default" onClick={this.setPwd}>
                {this.props.msg.pkg.get("update")}
              </button>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="hr"></div>

          <span className="inline-block margin-r-m">
            <div className="label">
              {this.props.msg.pkg.get("settings.pwd.old")}
            </div>
            <input
              name="old_pwd"
              type="password"
              onChange={this.changeOldPwd}
              value={this.state.oldPwd}
              placeholder={this.props.msg.pkg.get("settings.pwd.old")}
            />
          </span>

          <span className="inline-block margin-r-m">
            <div className="label">
              {this.props.msg.pkg.get("settings.pwd.new1")}
            </div>
            <input
              name="new_pwd1"
              type="password"
              onChange={this.changeNewPwd1}
              value={this.state.newPwd1}
              placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
            />
          </span>

          <span className="inline-block margin-r-m">
            <div className="label">
              {this.props.msg.pkg.get("settings.pwd.new2")}
            </div>
            <input
              name="new_pwd2"
              type="password"
              onChange={this.changeNewPwd2}
              value={this.state.newPwd2}
              placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
            />
          </span>
        </Container>

        <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">
                {this.props.msg.pkg.get("settings.chooseLan")}
              </h5>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="hr"></div>

          <div>
            <button
              onClick={() => {
                this.setLan("en_US");
              }}
              className="button-default inline-block margin-r-m"
            >
              {this.props.msg.pkg.get("enUS")}
            </button>
            <button
              onClick={() => {
                this.setLan("zh_CN");
              }}
              className="button-default inline-block margin-r-m"
            >
              {this.props.msg.pkg.get("zhCN")}
            </button>
          </div>
        </Container>

        {/* disabled */}
        {/* <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">
                {this.props.msg.pkg.get("settings.customLan")}
              </h5>,

              <span>
                <button
                  className="button-default"
                  onClick={this.syncPreferences}
                >
                  {this.props.msg.pkg.get("update")}
                </button>
              </span>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="hr"></div>

          <div className="inline-block margin-r-m">
            <div className="label">
              {this.props.msg.pkg.get("settings.lanPackURL")}
            </div>
            <input
              type="text"
              onChange={this.changeLanPackURL}
              value={this.props.login.preferences.lanPackURL}
              className="minor-font"
              style={{ width: "20rem" }}
              placeholder={this.props.msg.pkg.get("settings.lanPackURL")}
            />
          </div>
        </Container> */}

        <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">{this.props.msg.pkg.get("cfg.bg")}</h5>,

              <button className="button-default" onClick={this.syncPreferences}>
                {this.props.msg.pkg.get("update")}
              </button>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="hr"></div>

          <div>
            <div className="inline-block margin-r-m">
              <div className="label">
                {this.props.msg.pkg.get("cfg.bg.url")}
              </div>
              <input
                name="bg_url"
                type="text"
                onChange={this.changeBgUrl}
                value={this.props.login.preferences.bg.url}
                placeholder={this.props.msg.pkg.get("cfg.bg.url")}
              />
            </div>

            <div className="inline-block margin-r-m">
              <div className="label">
                {this.props.msg.pkg.get("cfg.bg.repeat")}
              </div>
              <input
                name="bg_repeat"
                type="text"
                onChange={this.changeBgRepeat}
                value={this.props.login.preferences.bg.repeat}
                placeholder={this.props.msg.pkg.get("cfg.bg.repeat")}
              />
            </div>

            <div className="inline-block margin-r-m">
              <div className="label">
                {this.props.msg.pkg.get("cfg.bg.pos")}
              </div>
              <input
                name="bg_pos"
                type="text"
                onChange={this.changeBgPos}
                value={this.props.login.preferences.bg.position}
                placeholder={this.props.msg.pkg.get("cfg.bg.pos")}
              />
            </div>

            <div className="inline-block margin-r-m">
              <div className="label">
                {this.props.msg.pkg.get("cfg.bg.align")}
              </div>
              <input
                name="bg_align"
                type="text"
                onChange={this.changeBgAlign}
                value={this.props.login.preferences.bg.align}
                placeholder={this.props.msg.pkg.get("cfg.bg.align")}
              />
            </div>
          </div>
        </Container>

        {errorReportPane}

        {/* <div className="hr"></div>
          <div>
            <Flexbox
              children={List([
                <span className="title-m bold">
                  {this.props.msg.pkg.get("prefer.theme")}
                </span>,

                <span>
                  <button className="button-default" onClick={this.syncPreferences}>
                    {this.props.msg.pkg.get("update")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            <Flowgrid
              grids={List([
                <div className="padding-t-m padding-b-m padding-r-m">
                  <div className="font-s grey1-font">
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
          </div> */}
      </div>
    );
  }
}
