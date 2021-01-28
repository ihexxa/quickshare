import * as React from "react";

import { ICoreState } from "./core_state";
import { IUsersClient } from "../client";
import { UsersClient } from "../client/users";
import { Updater as PanesUpdater } from "./panes";

export interface Props {
  authed: boolean;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;
  private static client: IUsersClient;

  static init = (props: Props) => (Updater.props = { ...props });

  static setClient = (client: IUsersClient): void => {
    Updater.client = client;
  };

  static login = async (user: string, pwd: string): Promise<boolean> => {
    const resp = await Updater.client.login(user, pwd);
    Updater.setAuthed(resp.status === 200);
    return resp.status === 200;
  };

  static logout = async (): Promise<boolean> => {
    const resp = await Updater.client.logout();
    Updater.setAuthed(false);
    return resp.status === 200;
  };

  static isAuthed = async (): Promise<boolean> => {
    const resp = await Updater.client.isAuthed();
    return resp.status === 200;
  };

  static initIsAuthed = async (): Promise<void> => {
    return Updater.isAuthed().then((isAuthed) => {
      Updater.setAuthed(isAuthed);
    });
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
    Updater.setClient(new UsersClient(""));
    this.update = p.update;
    this.state = {
      user: "",
      pwd: "",
    };

    this.initIsAuthed();
  }

  changeUser = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: ev.target.value });
  };

  changePwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ pwd: ev.target.value });
  };

  initIsAuthed = () => {
    Updater.initIsAuthed().then(() => {
      this.update(Updater.setAuthPane);
    });
  };

  login = () => {
    Updater.login(this.state.user, this.state.pwd).then((ok: boolean) => {
      if (ok) {
        this.update(Updater.setAuthPane);
        this.setState({ user: "", pwd: "" });
        // close all the panes
        PanesUpdater.displayPane("");
        this.update(PanesUpdater.updateState);
      } else {
        this.setState({ user: "", pwd: "" });
        alert("Failed to login.");
      }
    });
  };

  logout = () => {
    Updater.logout().then((ok: boolean) => {
      if (ok) {
        this.update(Updater.setAuthPane);
      } else {
        alert("Failed to logout.");
      }
    });
  };

  render() {
    return (
      <span>
        <h5 className="grey0-font">Login</h5>
        <span style={{ display: this.props.authed ? "none" : "inherit" }}>
          <input
            name="user"
            type="text"
            onChange={this.changeUser}
            value={this.state.user}
            className="margin-r-m black0-font"
            style={{ width: "12rem" }}
            placeholder="user name"
          />
          <input
            name="pwd"
            type="password"
            onChange={this.changePwd}
            value={this.state.pwd}
            className="margin-r-m black0-font"
            style={{ width: "12rem" }}
            placeholder="password"
          />
          <button onClick={this.login} className="green0-bg white-font">
            Log in
          </button>
        </span>
        <span style={{ display: this.props.authed ? "inherit" : "none" }}>
          <button
            onClick={this.logout}
            className="grey1-bg white-font margin-r-m"
          >
            Log out
          </button>
        </span>
      </span>
    );
  }
}
