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
  captchaID: string;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;
  private static client: IUsersClient;

  static init = (props: Props) => (Updater.props = { ...props });

  static setClient = (client: IUsersClient): void => {
    Updater.client = client;
  };

  static login = async (
    user: string,
    pwd: string,
    captchaID: string,
    captchaInput: string
  ): Promise<boolean> => {
    const resp = await Updater.client.login(user, pwd, captchaID, captchaInput);
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

  static getCaptchaID = async (): Promise<boolean> => {
    return Updater.client.getCaptchaID().then((resp) => {
      if (resp.status === 200) {
        Updater.props.captchaID = resp.data.id;
      }
      return resp.status === 200;
    });
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
  captchaInput: string;
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
      captchaInput: "",
    };

    this.initIsAuthed();
  }

  changeUser = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: ev.target.value });
  };

  changePwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ pwd: ev.target.value });
  };

  changeCaptcha = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ captchaInput: ev.target.value });
  };

  initIsAuthed = () => {
    Updater.initIsAuthed().then(() => {
      this.update(Updater.setAuthPane);
    });
  };

  login = () => {
    Updater.login(
      this.state.user,
      this.state.pwd,
      this.props.captchaID,
      this.state.captchaInput
    )
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
      .then(() => {
        return BrowserUpdater().isSharing(
          BrowserUpdater().props.dirPath.join("/")
        );
      })
      .then(() => {
        return BrowserUpdater().listSharings();
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
    return (
      <span>
        <div
          className="container"
          style={{ display: this.props.authed ? "none" : "block" }}
        >
          <div className="padding-l">
            <div className="flex-list-container">
              <div className="flex-list-item-l">
                <input
                  name="user"
                  type="text"
                  onChange={this.changeUser}
                  value={this.state.user}
                  className="black0-font margin-t-m margin-b-m margin-r-m"
                  placeholder="user name"
                />
                <input
                  name="pwd"
                  type="password"
                  onChange={this.changePwd}
                  value={this.state.pwd}
                  className="black0-font margin-t-m margin-b-m"
                  placeholder="password"
                />
              </div>
              <div className="flex-list-item-r">
                <button
                  onClick={this.login}
                  className="green0-bg white-font margin-t-m margin-b-m"
                >
                  Log in
                </button>
              </div>
            </div>

            <div className="flex-list-container">
              <div className="flex-list-item-l">
                <input
                  name="captcha"
                  type="text"
                  onChange={this.changeCaptcha}
                  value={this.state.captchaInput}
                  className="black0-font margin-t-m margin-b-m margin-r-m"
                  placeholder="captcha"
                />
                <img
                  src={`/v1/captchas/imgs?capid=${this.props.captchaID}`}
                  className="captcha"
                />
              </div>
              <div className="flex-list-item-l"></div>
            </div>
          </div>
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
