import * as React from "react";
import { List } from "immutable";

import { ICoreState } from "./core_state";
import { IUsersClient } from "../client";
import { UsersClient } from "../client/users";
import { Updater as PanesUpdater } from "./panes";
import { updater as BrowserUpdater } from "./browser.updater";
import { Layouter } from "./layouter";

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
    Updater.login(this.state.user, this.state.pwd)
      .then((ok: boolean) => {
        if (ok) {
          this.update(Updater.setAuthPane);
          this.setState({ user: "", pwd: "" });
          // close all the panes
          PanesUpdater.displayPane("");
          this.update(PanesUpdater.updateState);

          // refresh
          return BrowserUpdater().setHomeItems();
        } else {
          this.setState({ user: "", pwd: "" });
          alert("Failed to login.");
        }
      })
      .then(() => {
        return BrowserUpdater().refreshUploadings();
      })
      .then((_: boolean) => {
        this.update(BrowserUpdater().setBrowser);
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
    const elements: Array<JSX.Element> = [
      <input
        name="user"
        type="text"
        onChange={this.changeUser}
        value={this.state.user}
        className="black0-font margin-t-m margin-b-m"
        // style={{ width: "80%" }}
        placeholder="user name"
      />,
      <input
        name="pwd"
        type="password"
        onChange={this.changePwd}
        value={this.state.pwd}
        className="black0-font margin-t-m margin-b-m"
        // style={{ width: "80%" }}
        placeholder="password"
      />,
      <button
        onClick={this.login}
        className="green0-bg white-font margin-t-m margin-b-m"
      >
        Log in
      </button>,
    ];

    return (
      <span>
        <div
          className="margin-l-l"
          style={{ display: this.props.authed ? "none" : "block" }}
        >
          {/* <h5 className="black-font">Login</h5> */}
          <Layouter isHorizontal={false} elements={elements} />
        </div>

        <span style={{ display: this.props.authed ? "inherit" : "none" }}>
          <button onClick={this.logout} className="grey1-bg white-font">
            Log out
          </button>
        </span>
      </span>
    );
  }
}
