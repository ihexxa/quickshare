import * as React from "react";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { alertMsg, confirmMsg } from "../common/env";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { User, Quota } from "../client";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";

export interface AdminProps {
  users: Map<string, User>;
  roles: Set<string>;
}

export interface Props {
  admin: AdminProps;
  ui: UIProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface UserFormProps {
  key: string;
  id: string;
  name: string;
  role: string;
  quota: Quota;
  roles: Set<string>;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface UserFormState {
  id: string;
  name: string;
  newPwd1: string;
  newPwd2: string;
  role: string;
  quota: Quota;
}

export class UserForm extends React.Component<
  UserFormProps,
  UserFormState,
  {}
> {
  constructor(p: UserFormProps) {
    super(p);
    this.state = {
      id: p.id,
      name: p.name,
      newPwd1: "",
      newPwd2: "",
      role: p.role,
      quota: {
        spaceLimit: p.quota.spaceLimit,
        uploadSpeedLimit: p.quota.uploadSpeedLimit,
        downloadSpeedLimit: p.quota.downloadSpeedLimit,
      },
    };
  }

  changePwd1 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd1: ev.target.value });
  };
  changePwd2 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd2: ev.target.value });
  };
  changeRole = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ role: ev.target.value });
  };
  changeSpaceLimit = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({
      quota: {
        spaceLimit: ev.target.value,
        uploadSpeedLimit: this.state.quota.uploadSpeedLimit,
        downloadSpeedLimit: this.state.quota.downloadSpeedLimit,
      },
    });
  };
  changeUploadSpeedLimit = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({
      quota: {
        spaceLimit: this.state.quota.spaceLimit,
        uploadSpeedLimit: parseInt(ev.target.value, 10),
        downloadSpeedLimit: this.state.quota.downloadSpeedLimit,
      },
    });
  };
  changeDownloadSpeedLimit = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({
      quota: {
        spaceLimit: this.state.quota.spaceLimit,
        uploadSpeedLimit: this.state.quota.uploadSpeedLimit,
        downloadSpeedLimit: parseInt(ev.target.value, 10),
      },
    });
  };

  setPwd = async () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
      return;
    }

    return updater()
      .forceSetPwd(this.state.id, this.state.newPwd1)
      .then((ok: boolean) => {
        if (ok) {
          alertMsg(this.props.msg.pkg.get("update.ok"));
        } else {
          alertMsg(this.props.msg.pkg.get("update.fail"));
        }
        this.setState({
          newPwd1: "",
          newPwd2: "",
        });
      });
  };

  setUser = async () => {
    return updater()
      .setUser(this.props.id, this.state.role, this.state.quota)
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("update.fail"));
        } else {
          alertMsg(this.props.msg.pkg.get("update.ok"));
        }
        return updater().listUsers();
      })
      .then(() => {
        this.props.update(updater().updateAdmin);
      });
  };

  delUser = async () => {
    return updater()
      .delUser(this.state.id)
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("delete.fail"));
        }
        return updater().listUsers();
      })
      .then((_: boolean) => {
        this.props.update(updater().updateAdmin);
      });
  };

  render() {
    return (
      <div className="user-form">
        <Flexbox
          children={List([
            <div>
              <div className="key-value">
                {`${this.props.msg.pkg.get("user.name")}: `}
                <span className="value">{this.props.name}</span>
              </div>
              <div className="key-value">
                {`${this.props.msg.pkg.get("user.id")}: `}
                <span className="value">{this.props.id}</span>
              </div>
            </div>,

            <button onClick={this.delUser}>
              {this.props.msg.pkg.get("delete")}
            </button>,
          ])}
          childrenStyles={List([
            { alignItems: "flex-start", flexBasis: "70%" },
            {
              justifyContent: "flex-end",
            },
          ])}
        />

        <div className="hr"></div>

        <Flexbox
          children={List([
            <div>
              <span className="float-input">
                <div className="label">
                  {this.props.msg.pkg.get("user.role")}
                </div>
                <input
                  name={`${this.props.id}-role`}
                  type="text"
                  onChange={this.changeRole}
                  value={this.state.role}
                  placeholder={this.state.role}
                />
              </span>

              <span className="float-input">
                <div className="label">
                  {`${this.props.msg.pkg.get("spaceLimit")} (${FileSize(
                    parseInt(this.state.quota.spaceLimit, 10),
                    { round: 0 }
                  )})`}
                </div>
                <input
                  name={`${this.props.id}-spaceLimit`}
                  type="number"
                  onChange={this.changeSpaceLimit}
                  value={this.state.quota.spaceLimit}
                  placeholder={`${this.state.quota.spaceLimit}`}
                />
              </span>

              <span className="float-input">
                <div className="label">
                  {`${this.props.msg.pkg.get("uploadLimit")} (${FileSize(
                    this.state.quota.uploadSpeedLimit,
                    { round: 0 }
                  )})`}
                </div>
                <input
                  name={`${this.props.id}-uploadSpeedLimit`}
                  type="number"
                  onChange={this.changeUploadSpeedLimit}
                  value={this.state.quota.uploadSpeedLimit}
                  placeholder={`${this.state.quota.uploadSpeedLimit}`}
                />
              </span>

              <span className="float-input">
                <div className="label">
                  {`${this.props.msg.pkg.get("downloadLimit")} (${FileSize(
                    this.state.quota.downloadSpeedLimit,
                    { round: 0 }
                  )})`}
                </div>
                <input
                  name={`${this.props.id}-downloadSpeedLimit`}
                  type="number"
                  onChange={this.changeDownloadSpeedLimit}
                  value={this.state.quota.downloadSpeedLimit}
                  placeholder={`${this.state.quota.downloadSpeedLimit}`}
                />
              </span>
            </div>,

            <button onClick={this.setUser}>
              {this.props.msg.pkg.get("update")}
            </button>,
          ])}
          childrenStyles={List([
            { alignItems: "flex-start", flexBasis: "70%" },
            {
              justifyContent: "flex-end",
              flexBasis: "30%",
            },
          ])}
        />

        <div className="hr"></div>

        <Flexbox
          children={List([
            <div>
              <div className="float-input">
                <div className="label">
                  {this.props.msg.pkg.get("settings.pwd.new1")}
                </div>

                <input
                  name={`${this.props.id}-pwd1`}
                  type="password"
                  onChange={this.changePwd1}
                  value={this.state.newPwd1}
                  placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
                />
              </div>

              <div className="float-input">
                <div className="label">
                  {this.props.msg.pkg.get("settings.pwd.new2")}
                </div>
                <input
                  name={`${this.props.id}-pwd2`}
                  type="password"
                  onChange={this.changePwd2}
                  value={this.state.newPwd2}
                  placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
                />
              </div>
            </div>,

            <button onClick={this.setPwd}>
              {this.props.msg.pkg.get("update")}
            </button>,
          ])}
          childrenStyles={List([
            { alignItems: "flex-start", flexBasis: "70%" },
            {
              justifyContent: "flex-end",
            },
          ])}
        />
      </div>
    );
  }
}

export interface State {
  newUserName: string;
  newUserPwd1: string;
  newUserPwd2: string;
  newUserRole: string;
  newRole: string;
}
export class AdminPane extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = {
      newUserName: "",
      newUserPwd1: "",
      newUserPwd2: "",
      newUserRole: "",
      newRole: "",
    };
  }

  onChangeUserName = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newUserName: ev.target.value });
  };
  onChangeUserPwd1 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newUserPwd1: ev.target.value });
  };
  onChangeUserPwd2 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newUserPwd2: ev.target.value });
  };
  onChangeUserRole = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newUserRole: ev.target.value });
  };
  onChangeRole = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newRole: ev.target.value });
  };

  addRole = async () => {
    return updater()
      .addRole(this.state.newRole)
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("add.fail"));
        } else {
          alertMsg(this.props.msg.pkg.get("add.ok"));
        }
        return updater().listRoles();
      })
      .then(() => {
        this.props.update(updater().updateAdmin);
      });
  };

  delRole = async (role: string) => {
    if (!confirmMsg(this.props.msg.pkg.get("role.delete.warning"))) {
      return;
    }

    return updater()
      .delRole(role)
      .then((ok: boolean) => {
        if (!ok) {
          this.props.msg.pkg.get("delete.fail");
        } else {
          this.props.msg.pkg.get("delete.ok");
        }
        return updater().listRoles();
      })
      .then(() => {
        this.props.update(updater().updateAdmin);
      });
  };

  addUser = async () => {
    if (this.state.newUserPwd1 !== this.state.newUserPwd2) {
      alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
      return;
    }

    return updater()
      .addUser({
        id: "", // backend will fill it
        name: this.state.newUserName,
        pwd: this.state.newUserPwd1,
        role: this.state.newUserRole,
        quota: undefined,
        usedSpace: "0",
        preferences: undefined,
      })
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("add.fail"));
        } else {
          alertMsg(this.props.msg.pkg.get("add.ok"));
        }
        this.setState({
          newUserName: "",
          newUserPwd1: "",
          newUserPwd2: "",
          newUserRole: "",
        });
        return updater().listUsers();
      })
      .then(() => {
        this.props.update(updater().updateAdmin);
      });
  };

  render() {
    const userList = this.props.admin.users.valueSeq().map((user: User) => {
      return (
        <div key={user.id}>
          <UserForm
            key={user.id}
            id={user.id}
            name={user.name}
            role={user.role}
            quota={user.quota}
            roles={this.props.admin.roles}
            msg={this.props.msg}
            update={this.props.update}
          />
        </div>
      );
    });

    const roleList = this.props.admin.roles.valueSeq().map((role: string) => {
      return (
        <div key={role} className="role-list-item">
          <Flexbox
            children={List([
              <span>{role}</span>,
              <button
                onClick={() => {
                  this.delRole(role);
                }}
              >
                {this.props.msg.pkg.get("delete")}
              </button>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />
        </div>
      );
    });

    return (
      <div className="font-m">
        <Container>
          <BgCfg
            ui={this.props.ui}
            msg={this.props.msg}
            update={this.props.update}
          />
        </Container>

        <Container>
          <Flexbox
            children={List([
              <h5 className="title">{this.props.msg.pkg.get("user.add")}</h5>,
              <button onClick={this.addUser}>
                {this.props.msg.pkg.get("add")}
              </button>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <span className="float-input">
            <div className="label">{this.props.msg.pkg.get("user.name")}</div>
            <input
              type="text"
              onChange={this.onChangeUserName}
              value={this.state.newUserName}
              placeholder={this.props.msg.pkg.get("user.name")}
            />
          </span>

          <span className="float-input">
            <div className="label">{this.props.msg.pkg.get("user.role")}</div>
            <input
              type="text"
              onChange={this.onChangeUserRole}
              value={this.state.newUserRole}
              placeholder={this.props.msg.pkg.get("user.role")}
            />
          </span>

          <span className="float-input">
            <div className="label">
              {this.props.msg.pkg.get("settings.pwd.new1")}
            </div>
            <input
              type="password"
              onChange={this.onChangeUserPwd1}
              value={this.state.newUserPwd1}
              placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
            />
          </span>

          <span className="float-input">
            <div className="label">
              {this.props.msg.pkg.get("settings.pwd.new2")}
            </div>
            <input
              type="password"
              onChange={this.onChangeUserPwd2}
              value={this.state.newUserPwd2}
              placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
            />
          </span>
        </Container>

        <Container>
          <Flexbox
            children={List([
              <h5 className="title">
                {this.props.msg.pkg.get("admin.users")}
              </h5>,
              <span></span>,
            ])}
          />
          {userList}
        </Container>

        <Container>
          <div>
            <Flexbox
              children={List([
                <h5 className="title">{this.props.msg.pkg.get("role.add")}</h5>,
                <span></span>,
              ])}
            />

            <Flexbox
              children={List([
                <input
                  type="text"
                  onChange={this.onChangeRole}
                  value={this.state.newRole}
                  placeholder={this.props.msg.pkg.get("role.name")}
                />,
                <button onClick={this.addRole}>
                  {this.props.msg.pkg.get("add")}
                </button>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
          </div>

          <div className="hr"></div>

          <Flexbox
            children={List([
              <h5 className="title">
                {this.props.msg.pkg.get("admin.roles")}
              </h5>,
              <span></span>,
            ])}
          />
          {roleList}
        </Container>
      </div>
    );
  }
}

interface BgProps {
  msg: MsgProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

interface BgState {}
export class BgCfg extends React.Component<BgProps, BgState, {}> {
  changeSiteName = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({ ...this.props.ui, siteName: ev.target.value });
    this.props.update(updater().updateUI);
  };

  changeSiteDesc = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({ ...this.props.ui, siteDesc: ev.target.value });
    this.props.update(updater().updateUI);
  };

  changeBgUrl = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui,
      bg: { ...this.props.ui.bg, url: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  changeBgRepeat = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui,
      bg: { ...this.props.ui.bg, repeat: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  changeBgPos = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui,
      bg: { ...this.props.ui.bg, position: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  changeBgAlign = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui,
      bg: { ...this.props.ui.bg, align: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  constructor(p: BgProps) {
    super(p);
  }

  setClientCfg = async () => {
    const bgURL = this.props.ui.bg.url;
    if (bgURL.length === 0 || bgURL.length >= 4096) {
      alertMsg(this.props.msg.pkg.get("bg.url.alert"));
      return;
    }

    const bgRepeat = this.props.ui.bg.repeat;
    if (
      bgRepeat !== "repeat-x" &&
      bgRepeat !== "repeat-y" &&
      bgRepeat !== "repeat" &&
      bgRepeat !== "space" &&
      bgRepeat !== "round" &&
      bgRepeat !== "no-repeat"
    ) {
      alertMsg(this.props.msg.pkg.get("bg.repeat.alert"));
      return;
    }

    const bgPos = this.props.ui.bg.position;
    if (
      bgPos !== "top" &&
      bgPos !== "bottom" &&
      bgPos !== "left" &&
      bgPos !== "right" &&
      bgPos !== "center"
    ) {
      alertMsg(this.props.msg.pkg.get("bg.pos.alert"));
      return;
    }

    const bgAlign = this.props.ui.bg.align;
    if (bgAlign != "scroll" && bgAlign != "fixed" && bgAlign != "local") {
      alertMsg(this.props.msg.pkg.get("bg.align.alert"));
      return;
    }

    return updater()
      .setClientCfgRemote({
        siteName: this.props.ui.siteName,
        siteDesc: this.props.ui.siteDesc,
        bg: this.props.ui.bg,
      })
      .then((code: number) => {
        if (code === 200) {
          alertMsg(this.props.msg.pkg.get("update.ok"));
        } else {
          alertMsg(this.props.msg.pkg.get("update.fail"));
        }
      });
  };

  resetClientCfg = () => {
    // TODO: move this to backend
    updater().setClientCfg({
      siteName: "Quickshare",
      siteDesc: "Quickshare",
      bg: {
        url: "/static/img/textured_paper.png",
        repeat: "repeat",
        position: "center",
        align: "fixed",
      },
    });
    this.props.update(updater().updateUI);
  };

  render() {
    return (
      <div>
        <Flexbox
          children={List([
            <h5 className="title">
              {this.props.msg.pkg.get("cfg.bg")}
            </h5>,

            <span>
              <button onClick={this.resetClientCfg} className="margin-r-m">
                {this.props.msg.pkg.get("reset")}
              </button>
              <button onClick={this.setClientCfg}>
                {this.props.msg.pkg.get("update")}
              </button>
            </span>,
          ])}
          childrenStyles={List([{}, { justifyContent: "flex-end" }])}
        />

        <div>
          <div className="float-input">
            <div className="label">{this.props.msg.pkg.get("cfg.bg.url")}</div>
            <input
              name="bg_url"
              type="text"
              onChange={this.changeBgUrl}
              value={this.props.ui.bg.url}
              style={{ width: "20rem" }}
              placeholder={this.props.msg.pkg.get("cfg.bg.url")}
            />
          </div>

          <div className="float-input">
            <div className="label">
              {this.props.msg.pkg.get("cfg.bg.repeat")}
            </div>
            <input
              name="bg_repeat"
              type="text"
              onChange={this.changeBgRepeat}
              value={this.props.ui.bg.repeat}
              placeholder={this.props.msg.pkg.get("cfg.bg.repeat")}
            />
          </div>

          <div className="float-input">
            <div className="label">{this.props.msg.pkg.get("cfg.bg.pos")}</div>
            <input
              name="bg_pos"
              type="text"
              onChange={this.changeBgPos}
              value={this.props.ui.bg.position}
              placeholder={this.props.msg.pkg.get("cfg.bg.pos")}
            />
          </div>

          <div className="float-input">
            <div className="label">
              {this.props.msg.pkg.get("cfg.bg.align")}
            </div>
            <input
              name="bg_align"
              type="text"
              onChange={this.changeBgAlign}
              value={this.props.ui.bg.align}
              placeholder={this.props.msg.pkg.get("cfg.bg.align")}
            />
          </div>
        </div>
      </div>
    );
  }
}
