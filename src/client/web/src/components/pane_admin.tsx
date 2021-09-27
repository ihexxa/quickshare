import * as React from "react";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { alertMsg, confirmMsg } from "../common/env";
import { ICoreState, MsgProps } from "./core_state";
import { User, Quota } from "../client";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";

export interface AdminProps {
  users: Map<string, User>;
  roles: Set<string>;
}

export interface Props {
  admin: AdminProps;
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
      <div
        style={{
          border: "dashed 2px #ccc",
          padding: "1rem",
        }}
      >
        <Flexbox
          children={List([
            <div>
              <span className="bold">
                {this.props.msg.pkg.get("user.name")} {this.props.name}
              </span>
              &nbsp;/&nbsp;
              <span className="grey0-font">
                {this.props.msg.pkg.get("user.id")} {this.props.id}
              </span>
            </div>,

            <button onClick={this.delUser} className="margin-r-m">
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

        <div className="hr white0-bg margin-t-m margin-b-m"></div>

        <Flexbox
          children={List([
            <div>
              <span className="inline-block">
                <div className="margin-r-m font-size-s grey1-font">
                  {this.props.msg.pkg.get("user.role")}
                </div>
                <input
                  name={`${this.props.id}-role`}
                  type="text"
                  onChange={this.changeRole}
                  value={this.state.role}
                  className="black0-font margin-r-m"
                  placeholder={this.state.role}
                />
              </span>

              <span className="margin-t-m inline-block">
                <div className="margin-r-m font-size-s grey1-font">
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
                  className="black0-font margin-r-m"
                  placeholder={`${this.state.quota.spaceLimit}`}
                />
              </span>

              <span className="margin-t-m inline-block">
                <div className="margin-r-m font-size-s grey1-font">
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
                  className="black0-font margin-r-m"
                  placeholder={`${this.state.quota.uploadSpeedLimit}`}
                />
              </span>

              <span className="margin-t-m inline-block">
                <div className="margin-r-m font-size-s grey1-font">
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
                  className="black0-font margin-r-m"
                  placeholder={`${this.state.quota.downloadSpeedLimit}`}
                />
              </span>
            </div>,

            <button onClick={this.setUser} className="margin-r-m">
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

        <div className="hr white0-bg margin-t-m margin-b-m"></div>

        <Flexbox
          children={List([
            <div>
              <div className="font-size-s grey1-font">
                {this.props.msg.pkg.get("user.password")}
              </div>

              <input
                name={`${this.props.id}-pwd1`}
                type="password"
                onChange={this.changePwd1}
                value={this.state.newPwd1}
                className="black0-font margin-b-m margin-r-m"
                placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
              />
              <input
                name={`${this.props.id}-pwd2`}
                type="password"
                onChange={this.changePwd2}
                value={this.state.newPwd2}
                className="black0-font margin-b-m"
                placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
              />
            </div>,

            <button onClick={this.setPwd} className="margin-r-m">
              {this.props.msg.pkg.get("update")}
            </button>,
          ])}
          childrenStyles={List([
            { alignItems: "flex-start", flexBasis: "70%" },
            {
              justifyContent: "flex-end",
            },
          ])}
          className="margin-t-m"
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
    if (
      !confirmMsg(
        // "After deleting this role, some of users may not be able to login."
        this.props.msg.pkg.get("role.delete.warning")
      )
    ) {
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
        <div key={user.id} className="margin-t-m">
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
        <div key={role} className="flex-list-container margin-b-m">
          <div className="flex-list-item-l">
            <span>{role}</span>
          </div>
          <div className="flex-list-item-r">
            <button
              onClick={() => {
                this.delRole(role);
              }}
              className="margin-r-m"
            >
              {this.props.msg.pkg.get("delete")}
            </button>
          </div>
        </div>
      );
    });

    return (
      <div className="font-size-m">
        <div className="container padding-l">
          <Flexbox
            children={List([
              <span>{this.props.msg.pkg.get("user.add")}</span>,
              <span></span>,
            ])}
            className="title-m bold margin-b-m"
          />

          <span className="inline-block margin-r-m margin-b-m">
            <div className="font-size-s grey1-font">
              {this.props.msg.pkg.get("user.name")}
            </div>
            <input
              type="text"
              onChange={this.onChangeUserName}
              value={this.state.newUserName}
              className="black0-font margin-r-m margin-b-m"
              placeholder={this.props.msg.pkg.get("user.name")}
            />
          </span>

          <span className="inline-block margin-r-m margin-b-m">
            <div className="font-size-s grey1-font">
              {this.props.msg.pkg.get("user.role")}
            </div>
            <input
              type="text"
              onChange={this.onChangeUserRole}
              value={this.state.newUserRole}
              className="black0-font margin-r-m margin-b-m"
              placeholder={this.props.msg.pkg.get("user.role")}
            />
          </span>

          <span className="inline-block margin-r-m margin-b-m">
            <div className="font-size-s grey1-font">
              {this.props.msg.pkg.get("settings.pwd.new1")}
            </div>
            <input
              type="password"
              onChange={this.onChangeUserPwd1}
              value={this.state.newUserPwd1}
              className="black0-font margin-r-m margin-b-m"
              placeholder={this.props.msg.pkg.get("settings.pwd.new1")}
            />
          </span>

          <span className="inline-block margin-r-m margin-b-m">
            <div className="font-size-s grey1-font">
              {this.props.msg.pkg.get("settings.pwd.new2")}
            </div>
            <input
              type="password"
              onChange={this.onChangeUserPwd2}
              value={this.state.newUserPwd2}
              className="black0-font margin-r-m margin-b-m"
              placeholder={this.props.msg.pkg.get("settings.pwd.new2")}
            />
          </span>

          <span className="inline-block margin-r-m margin-b-m">
            <button onClick={this.addUser} className="margin-r-m">
              {this.props.msg.pkg.get("add")}
            </button>
          </span>
        </div>

        <div className="container padding-l">
          <Flexbox
            children={List([
              <span>{this.props.msg.pkg.get("admin.users")}</span>,
              <span></span>,
            ])}
            className="title-m bold margin-b-m"
          />
          {userList}
        </div>

        <div className="container padding-l">
          <div>
            <Flexbox
              children={List([
                <span>{this.props.msg.pkg.get("role.add")}</span>,
                <span></span>,
              ])}
              className="title-m bold margin-b-m"
            />

            <Flexbox
              children={List([
                <span className="inline-block margin-t-m margin-b-m">
                  <input
                    type="text"
                    onChange={this.onChangeRole}
                    value={this.state.newRole}
                    className="black0-font margin-r-m"
                    placeholder={this.props.msg.pkg.get("role.name")}
                  />
                </span>,
                <button onClick={this.addRole} className="margin-r-m">
                  {this.props.msg.pkg.get("add")}
                </button>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
          </div>

          <div className="hr white0-bg margin-t-m margin-b-m"></div>

          <Flexbox
            children={List([
              <span>{this.props.msg.pkg.get("admin.roles")}</span>,
              <span></span>,
            ])}
            className="title-m bold margin-b-m"
          />
          {roleList}
        </div>
      </div>
    );
  }
}
