import * as React from "react";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { RiMenuUnfoldFill } from "@react-icons/all-files/ri/RiMenuUnfoldFill";

import { Env } from "../common/env";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { User, Quota } from "../client";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { loadingCtrl, ctrlOn, ctrlOff } from "../common/controls";
import { getIconWithProps, iconSize } from "./visual/icons";
import { Columns } from "./layout/columns";

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
  usedSpace: string;
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
  folded: boolean;
  usedSpace: string;
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
      folded: true,
      usedSpace: p.usedSpace,
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

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  resetUsedSpace = async (userID: string) => {
    if (!Env().confirmMsg(this.props.msg.pkg.get("confirm.resetUsedSpace"))) {
      return;
    }

    const status = await updater().resetUsedSpace(userID);
    if (status !== "") {
      Env().alertMsg(this.props.msg.pkg.get("resetUsedSpace"));
    }
  };

  setPwd = async () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      Env().alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
      return;
    }

    this.setLoading(true);
    try {
      const status = await updater().forceSetPwd(
        this.state.id,
        this.state.newPwd1
      );
      if (status !== "") {
        Env().alertMsg(this.props.msg.pkg.get("update.fail"));
        return;
      }
      Env().alertMsg(this.props.msg.pkg.get("update.ok"));
    } finally {
      this.setLoading(false);
    }
  };

  setUser = async () => {
    this.setLoading(true);
    try {
      const status = await updater().setUser(
        this.props.id,
        this.state.role,
        this.state.quota
      );
      if (status !== "") {
        Env().alertMsg(this.props.msg.pkg.get("update.fail"));
        return;
      }

      const listStatus = await updater().listUsers();
      if (listStatus !== "") {
        Env().alertMsg(this.props.msg.pkg.get("update.fail"));
        return;
      }

      Env().alertMsg(this.props.msg.pkg.get("update.ok"));
    } finally {
      this.props.update(updater().updateAdmin);
      this.setLoading(false);
    }
  };

  delUser = async () => {
    if (!Env().confirmMsg(this.props.msg.pkg.get("op.confirm"))) {
      return;
    }

    this.setLoading(true);
    try {
      const status = await updater().delUser(this.state.id);
      if (status !== "") {
        Env().alertMsg(this.props.msg.pkg.get("delete.fail"));
        return;
      }

      const listStatus = await updater().listUsers();
      if (listStatus !== "") {
        Env().alertMsg(this.props.msg.pkg.get("op.fail"));
        return;
      }

      Env().alertMsg(this.props.msg.pkg.get("delete.ok"));
    } finally {
      this.props.update(updater().updateAdmin);
      this.setLoading(false);
    }
  };

  toggle = () => {
    this.setState({ folded: !this.state.folded });
  };

  render() {
    const foldedClass = this.state.folded ? "hidden" : "";
    const foldIconColor = this.state.folded ? "major-font" : "focus-font";
    const resetUsedSpace = () => {
      this.resetUsedSpace(this.props.id);
    };

    return (
      <div className="padding-t-m padding-b-m">
        <Columns
          colKey={"paneAdmin"}
          rows={List([
            List([
              <div className="title-m-wrap">
                <span className="">{`${this.props.msg.pkg.get(
                  "user.name"
                )}: `}</span>
                <span className="margin-r-m">{this.props.name}</span>
                <span className="">{`${this.props.msg.pkg.get(
                  "user.id"
                )}: `}</span>
                <span>{this.props.id}</span>
              </div>,

              <div className="txt-align-r">
                <div className="icon-s inline-block">
                  {getIconWithProps("GoUnfold", {
                    size: iconSize("s"),
                    className: `mr-8 ${foldIconColor}`,
                    onClick: this.toggle,
                  })}
                  {/* <RiMenuUnfoldFill
                    size={iconSize("s")}
                   
                  /> */}
                </div>
              </div>,

              <button className="" onClick={this.delUser}>
                {this.props.msg.pkg.get("delete")}
              </button>,
            ]),
          ])}
          widths={List(["calc(100% - 10rem)", "3rem", "7rem"])}
          childrenClassNames={List(["", "txt-align-r", "txt-align-r"])}
        />

        <div></div>

        <div className={`info major-bg ${foldedClass}`}>
          <div>
            <Flexbox
              children={List([
                <span>
                  {`${this.props.msg.pkg.get("usedSpace")}: ${FileSize(
                    parseInt(this.state.usedSpace, 10),
                    { round: 0 }
                  )}`}
                </span>,
                <button className="" onClick={resetUsedSpace}>
                  {this.props.msg.pkg.get("resetUsedSpace")}
                </button>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
          </div>

          <div className="hr"></div>

          <Flexbox
            className="margin-t-m"
            children={List([
              <div>
                <span className="inline-block margin-r-m">
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

                <span className="inline-block margin-r-m">
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

                <span className="inline-block margin-r-m">
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

                <span className="inline-block margin-r-m">
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

              <div>
                <button className="" onClick={this.setUser}>
                  {this.props.msg.pkg.get("update")}
                </button>
              </div>,
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
            className="margin-t-m"
            children={List([
              <div>
                <div className="inline-block margin-r-m">
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

                <div className="inline-block">
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

              <button className="" onClick={this.setPwd}>
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

        <div className="hr"></div>
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

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  addRole = async () => {
    this.setLoading(true);
    try {
      const status = await updater().addRole(this.state.newRole);
      if (status !== "") {
        Env().alertMsg(this.props.msg.pkg.get("add.fail"));
        return;
      }

      const listStatus = await updater().listRoles();
      if (listStatus !== "") {
        Env().alertMsg(this.props.msg.pkg.get("add.fail"));
        return;
      }

      Env().alertMsg(this.props.msg.pkg.get("add.ok"));
    } finally {
      this.props.update(updater().updateAdmin);
      this.setLoading(false);
    }
  };

  delRole = async (role: string) => {
    if (!Env().confirmMsg(this.props.msg.pkg.get("role.delete.warning"))) {
      return;
    }

    this.setLoading(true);

    try {
      const status = await updater().delRole(role);
      if (status !== "") {
        this.props.msg.pkg.get("delete.fail");
        return;
      }

      const listStatus = await updater().listRoles();
      if (listStatus !== "") {
        Env().alertMsg(this.props.msg.pkg.get("add.fail"));
        return;
      }

      this.props.msg.pkg.get("delete.ok");
    } finally {
      this.setLoading(false);
      this.props.update(updater().updateAdmin);
    }
  };

  addUser = async () => {
    if (this.state.newUserPwd1 !== this.state.newUserPwd2) {
      Env().alertMsg(this.props.msg.pkg.get("settings.pwd.notSame"));
      return;
    }

    this.setLoading(true);

    try {
      const status = await updater().addUser({
        id: "", // backend will fill it
        name: this.state.newUserName,
        pwd: this.state.newUserPwd1,
        role: this.state.newUserRole,
        quota: undefined,
        usedSpace: "0",
        preferences: undefined,
      });
      if (status !== "") {
        Env().alertMsg(this.props.msg.pkg.get("add.fail"));
        return;
      }

      const listStatus = await updater().listUsers();
      if (listStatus !== "") {
        Env().alertMsg(this.props.msg.pkg.get("op.fail"));
        return;
      }

      Env().alertMsg(this.props.msg.pkg.get("add.ok"));
    } finally {
      this.setState({
        newUserName: "",
        newUserPwd1: "",
        newUserPwd2: "",
        newUserRole: "",
      });
      this.setLoading(false);
      this.props.update(updater().updateAdmin);
    }
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
            usedSpace={user.usedSpace}
            roles={this.props.admin.roles}
            msg={this.props.msg}
            update={this.props.update}
          />
        </div>
      );
    });

    const roleList = this.props.admin.roles.valueSeq().map((role: string) => {
      return (
        <div key={role} className="margin-b-m">
          <Flexbox
            children={List([
              <span>{role}</span>,
              <button
                onClick={() => {
                  this.delRole(role);
                }}
                className=""
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
      <div>
        <Container>
          <SiteCfg
            ui={this.props.ui}
            msg={this.props.msg}
            update={this.props.update}
          />
        </Container>

        <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">{this.props.msg.pkg.get("user.add")}</h5>,
              <button onClick={this.addUser} className="">
                {this.props.msg.pkg.get("add")}
              </button>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="hr"></div>

          <span className="inline-block margin-r-m">
            <div className="label">{this.props.msg.pkg.get("user.name")}</div>
            <input
              type="text"
              onChange={this.onChangeUserName}
              value={this.state.newUserName}
              placeholder={this.props.msg.pkg.get("user.name")}
            />
          </span>

          <span className="inline-block margin-r-m">
            <div className="label">{this.props.msg.pkg.get("user.role")}</div>
            <input
              type="text"
              onChange={this.onChangeUserRole}
              value={this.state.newUserRole}
              placeholder={this.props.msg.pkg.get("user.role")}
            />
          </span>

          <span className="inline-block margin-r-m">
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

          <span className="inline-block margin-r-m">
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
              <h5 className="title-m">
                {this.props.msg.pkg.get("admin.users")}
              </h5>,
              <span></span>,
            ])}
          />

          <div className="hr"></div>

          {userList}
        </Container>

        {/* <Container>
          <div>
            <Flexbox
              children={List([
                <h5 className="title-m">
                  {this.props.msg.pkg.get("role.add")}
                </h5>,
                <span></span>,
              ])}
            />

            <div className="hr"></div>

            <Flexbox
              children={List([
                <input
                  type="text"
                  onChange={this.onChangeRole}
                  value={this.state.newRole}
                  placeholder={this.props.msg.pkg.get("role.name")}
                />,
                <button className="" onClick={this.addRole}>
                  {this.props.msg.pkg.get("add")}
                </button>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
          </div>
        </Container> */}

        {/* <Container>
          <Flexbox
            children={List([
              <h5 className="title-m">
                {this.props.msg.pkg.get("admin.roles")}
              </h5>,
              <span></span>,
            ])}
          />

          <div className="hr"></div>

          {roleList}
        </Container> */}
      </div>
    );
  }
}

interface SiteCfgProps {
  msg: MsgProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

interface SiteCfgState {}
export class SiteCfg extends React.Component<SiteCfgProps, SiteCfgState, {}> {
  onChangeSiteName = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      siteName: ev.target.value,
    });
    this.props.update(updater().updateUI);
  };
  onChangeSiteDesc = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      siteDesc: ev.target.value,
    });
    this.props.update(updater().updateUI);
  };
  onChangeAllowSetBg = (enabled: boolean) => {
    updater().setClientCfg({ ...this.props.ui.clientCfg, allowSetBg: enabled });
    this.props.update(updater().updateUI);
  };
  onChangeAutoTheme = (enabled: boolean) => {
    updater().setClientCfg({ ...this.props.ui.clientCfg, autoTheme: enabled });
    this.props.update(updater().updateUI);
  };

  changeBgUrl = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      bg: { ...this.props.ui.clientCfg.bg, url: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  changeBgRepeat = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      bg: { ...this.props.ui.clientCfg.bg, repeat: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  changeBgPos = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      bg: { ...this.props.ui.clientCfg.bg, position: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  changeBgAlign = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      bg: { ...this.props.ui.clientCfg.bg, align: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };
  changeBgBgColor = (ev: React.ChangeEvent<HTMLInputElement>) => {
    updater().setClientCfg({
      ...this.props.ui.clientCfg,
      bg: { ...this.props.ui.clientCfg.bg, bgColor: ev.target.value },
    });
    this.props.update(updater().updateUI);
  };

  constructor(p: SiteCfgProps) {
    super(p);
  }

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  setClientCfg = async () => {
    const bgURL = this.props.ui.clientCfg.bg.url;
    if (bgURL.length >= 4096) {
      Env().alertMsg(this.props.msg.pkg.get("bg.url.alert"));
      return;
    }

    const bgRepeat = this.props.ui.clientCfg.bg.repeat;
    if (
      bgRepeat !== "repeat-x" &&
      bgRepeat !== "repeat-y" &&
      bgRepeat !== "repeat" &&
      bgRepeat !== "space" &&
      bgRepeat !== "round" &&
      bgRepeat !== "no-repeat"
    ) {
      Env().alertMsg(this.props.msg.pkg.get("bg.repeat.alert"));
      return;
    }

    const bgPos = this.props.ui.clientCfg.bg.position;
    if (
      bgPos !== "top" &&
      bgPos !== "bottom" &&
      bgPos !== "left" &&
      bgPos !== "right" &&
      bgPos !== "center"
    ) {
      Env().alertMsg(this.props.msg.pkg.get("bg.pos.alert"));
      return;
    }

    const bgAlign = this.props.ui.clientCfg.bg.align;
    if (bgAlign !== "scroll" && bgAlign !== "fixed" && bgAlign !== "local") {
      Env().alertMsg(this.props.msg.pkg.get("bg.align.alert"));
      return;
    }

    this.setLoading(true);

    try {
      const status = await updater().setClientCfgRemote({
        clientCfg: this.props.ui.clientCfg,
        // captchaEnabled is omitted
      });
      if (status !== "") {
        Env().alertMsg(this.props.msg.pkg.get("update.fail"));
        return;
      }

      Env().alertMsg(this.props.msg.pkg.get("update.ok"));
    } finally {
      this.setLoading(false);
    }
  };

  resetClientCfg = () => {
    // TODO: move this to backend
    updater().setClientCfg({
      siteName: "Quickshare",
      siteDesc: "Quickshare",
      bg: {
        url: "",
        repeat: "repeat",
        position: "center",
        align: "fixed",
        bgColor: "",
      },
      allowSetBg: false,
      autoTheme: true,
    });
    this.props.update(updater().updateUI);
  };

  reindex = async () => {
    if (!Env().confirmMsg(this.props.msg.pkg.get("action.reindex.confirm"))) {
      return;
    }
    return updater().reindex();
  };

  render() {
    return (
      <div>
        <Flexbox
          children={List([
            <h5 className="title-m">
              {this.props.msg.pkg.get("siteSettings")}
            </h5>,

            <span>
              <button
                onClick={this.resetClientCfg}
                className="inline-block margin-r-m "
              >
                {this.props.msg.pkg.get("reset")}
              </button>
              <button className="inline-block " onClick={this.setClientCfg}>
                {this.props.msg.pkg.get("update")}
              </button>
            </span>,
          ])}
          childrenStyles={List([{}, { justifyContent: "flex-end" }])}
        />

        <div className="hr"></div>

        <span className="inline-block margin-r-m">
          <div className="label">{this.props.msg.pkg.get("siteName")}</div>
          <input
            type="text"
            onChange={this.onChangeSiteName}
            value={this.props.ui.clientCfg.siteName}
            placeholder={this.props.msg.pkg.get("siteName")}
          />
        </span>

        <span className="inline-block margin-r-m">
          <div className="label">{this.props.msg.pkg.get("siteDesc")}</div>
          <input
            type="text"
            onChange={this.onChangeSiteDesc}
            value={this.props.ui.clientCfg.siteDesc}
            placeholder={this.props.msg.pkg.get("siteDesc")}
          />
        </span>

        <div className="hr"></div>

        <div>
          <div className="label">{this.props.msg.pkg.get("allowSetBg")}</div>
          <button
            onClick={() => {
              this.onChangeAllowSetBg(true);
            }}
            className={`${
              this.props.ui.clientCfg.allowSetBg ? "focus-font" : ""
            } inline-block margin-r-m`}
          >
            {this.props.msg.pkg.get("term.enabled")}
          </button>
          <button
            onClick={() => {
              this.onChangeAllowSetBg(false);
            }}
            className={`${
              this.props.ui.clientCfg.allowSetBg ? "" : "focus-font"
            } inline-block margin-r-m`}
          >
            {this.props.msg.pkg.get("term.disabled")}
          </button>
        </div>

        <div className="hr"></div>

        <div>
          <div className="label">{this.props.msg.pkg.get("autoTheme")}</div>
          <button
            onClick={() => {
              this.onChangeAutoTheme(true);
            }}
            className={`${
              this.props.ui.clientCfg.autoTheme ? "focus-font" : ""
            } inline-block margin-r-m`}
          >
            {this.props.msg.pkg.get("term.enabled")}
          </button>
          <button
            onClick={() => {
              this.onChangeAutoTheme(false);
            }}
            className={`${
              this.props.ui.clientCfg.autoTheme ? "" : "focus-font"
            } inline-block margin-r-m`}
          >
            {this.props.msg.pkg.get("term.disabled")}
          </button>
        </div>

        <div className="hr"></div>

        <div>
          <div className="inline-block margin-r-m">
            <div className="label">{this.props.msg.pkg.get("cfg.bg.url")}</div>
            <input
              name="bg_url"
              type="text"
              onChange={this.changeBgUrl}
              value={this.props.ui.clientCfg.bg.url}
              style={{ width: "20rem" }}
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
              value={this.props.ui.clientCfg.bg.repeat}
              placeholder={this.props.msg.pkg.get("cfg.bg.repeat")}
            />
          </div>

          <div className="inline-block margin-r-m">
            <div className="label">{this.props.msg.pkg.get("cfg.bg.pos")}</div>
            <input
              name="bg_pos"
              type="text"
              onChange={this.changeBgPos}
              value={this.props.ui.clientCfg.bg.position}
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
              value={this.props.ui.clientCfg.bg.align}
              placeholder={this.props.msg.pkg.get("cfg.bg.align")}
            />
          </div>

          <div className="inline-block">
            <div className="label">
              {this.props.msg.pkg.get("cfg.bg.bgColor")}
            </div>
            <input
              name="bg_bgColor"
              type="text"
              onChange={this.changeBgBgColor}
              value={this.props.ui.clientCfg.bg.bgColor}
              placeholder={this.props.msg.pkg.get("cfg.bg.bgColor")}
            />
          </div>
        </div>

        <div className="hr"></div>

        <div>
          <div className="label">
            {this.props.msg.pkg.get("action.reindex.desc")}
          </div>
          <button className="" onClick={this.reindex}>
            {this.props.msg.pkg.get("action.reindex")}
          </button>
        </div>
      </div>
    );
  }
}
