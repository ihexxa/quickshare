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

  static isAuthed = async (): Promise<boolean> => {
    const status = await usersClient.isAuthed();
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
  user: string;
  pwd: string;
}

export class AuthPane extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  constructor(p: Props) {
    super(p);
    Updater.init(p);
    this.update = p.update;
    this.state = {
      user: "visitor",
      pwd: "",
    };

    this.checkAuthed();
  }

  checkAuthed = () => {
    Updater.isAuthed().then((isAuthed) => {
      Updater.setAuthed(isAuthed);
      this.update(Updater.setAuthPane);
    })
  }

  changeUser = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: ev.target.value });
  };

  changePwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ pwd: ev.target.value });
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

  render() {
    return (
      <div>
        <div style={{ display: this.props.authed ? "none" : "inherit" }}>
          <input
            name="user"
            type="text"
            onChange={this.changeUser}
            value={this.state.user}
            className="margin-r-m black0-font"
          />
          <input
            name="pwd"
            type="password"
            onChange={this.changePwd}
            value={this.state.pwd}
            className="margin-r-m black0-font"
          />
          <button onClick={this.login} className="green0-bg white-font">
            Log in
          </button>
        </div>
        <div style={{ display: this.props.authed ? "inherit" : "none" }}>
          <button
            onClick={this.logout}
            className="grey1-bg white-font margin-r-m"
          >
            Log out
          </button>
        </div>
      </div>
    );
  }
}
