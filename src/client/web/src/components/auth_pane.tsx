import * as React from "react";

import { ICoreState } from "./core_state";
import { usersClient } from "../client";

export interface Props {
  authed: boolean;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;

  static init = (props: Props) => (Updater.props = { ...props });

  static login = async (user: string, pwd: string): Promise<boolean> => {
    const status = await usersClient.login(user, pwd);
    return status == 200;
  };

  static logout = async (): Promise<boolean> => {
    const status = await usersClient.logout();
    return status == 200;
  };

  static setPwd = async (oldPwd: string, newPwd: string): Promise<boolean> => {
    const status = await usersClient.setPwd(oldPwd, newPwd);
    return status == 200;
  };

  static setAuthed = (isAuthed: boolean) => {
    Updater.props.authed = isAuthed;
  };

  static setAuthPane = (preState: ICoreState): ICoreState => {
    preState.panel.authPane = {
      ...preState.panel.authPane,
      ...Updater.props,
    };
    return preState;
  };
}

export interface State {
  show: boolean;
  user: string;
  pwd: string;
  oldPwd: string;
  newPwd1: string;
  newPwd2: string;
}

export class AuthPane extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  constructor(p: Props) {
    super(p);
    Updater.init(p);
    this.update = p.update;
    this.state = {
      show: false,
      user: "visitor",
      pwd: "",
      oldPwd: "",
      newPwd1: "",
      newPwd2: "",
    }
  }

  changeUser = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: ev.target.value });
  };

  changePwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ pwd: ev.target.value });
  };
  changeOldPwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ oldPwd: ev.target.value });
  };
  changeNewPwd1 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd1: ev.target.value });
  };
  changeNewPwd2 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd2: ev.target.value });
  };

  showPane = () => {
    this.setState({ show: !this.state.show });
  };

  login = () => {
    Updater.login(this.state.user, this.state.pwd).then((ok: boolean) => {
      if (ok) {
        Updater.setAuthed(true);
        this.update(Updater.setAuthPane);
      } else {
        // alert
      }
    });
  };

  logout = () => {
    Updater.logout().then((ok: boolean) => {
      if (ok) {
        Updater.setAuthed(false);
        this.update(Updater.setAuthPane);
      } else {
        alert("fail");
      }
    });
  };

  setPwd = () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      // alert
      alert("new pwds not same");
      return;
    }
    if (this.state.oldPwd == this.state.newPwd1) {
      // alert
      alert("old and new pwds are same");
      return;
    }
    Updater.setPwd(this.state.oldPwd, this.state.newPwd1).then(
      (ok: boolean) => {
        if (ok) {
          // hint
          alert("ok");
        } else {
          // alert
        }
      }
    );
  };

  render() {
    return (
      <div>
        <div style={{ display: this.props.authed ? "none" : "inherit" }}>
          <input
            name="user"
            type="text"
            onChange={this.changeUser}
            value={this.state.user}
          />
          <input
            name="pwd"
            type="text"
            onChange={this.changePwd}
            value={this.state.pwd}
          />
          <button onClick={this.login}>Log in</button>
        </div>
        <div style={{ display: this.props.authed ? "none" : "inherit" }}>
          <button onClick={this.logout}>Log out</button>

          <button onClick={this.showPane}>show</button>
          <div style={{ display: this.state.show ? "inherit" : "none" }}>
            <input
              name="old_pwd"
              type="text"
              onChange={this.changeOldPwd}
              value={this.state.oldPwd}
            />
            <input
              name="new_pwd1"
              type="text"
              onChange={this.changeNewPwd1}
              value={this.state.newPwd1}
            />
            <input
              name="new_pwd2"
              type="text"
              onChange={this.changeNewPwd2}
              value={this.state.newPwd2}
            />
            <button onClick={this.setPwd}>Update</button>
          </div>
        </div>
      </div>
    );
  }
}
