import * as React from "react";
import { Map, Set } from "immutable";

import { ICoreState } from "./core_state";
import { IUsersClient, User, ListUsersResp, ListRolesResp } from "../client";
import { UsersClient } from "../client/users";
import { Updater as PanesUpdater } from "./panes";
import { updater as BrowserUpdater } from "./browser.updater";
import { Layouter } from "./layouter";

export interface Props {
  users: Map<string, User>;
  roles: Set<string>;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;
  private static client: IUsersClient;

  static init = (props: Props) => (Updater.props = { ...props });

  static setClient = (client: IUsersClient): void => {
    Updater.client = client;
  };

  static addUser = async (user: User): Promise<boolean> => {
    const resp = await Updater.client.addUser(user.name, user.pwd, user.role);
    // TODO: should return uid instead
    return resp.status === 200;
  };

  static delUser = async (userID: string): Promise<boolean> => {
    const resp = await Updater.client.delUser(userID);
    return resp.status === 200;
  };

  static listUsers = async (): Promise<boolean> => {
    const resp = await Updater.client.listUsers();
    if (resp.status !== 200) {
      return false;
    }

    const lsRes = resp.data as ListUsersResp;
    let users = Map<User>({});
    lsRes.users.forEach((user: User) => {
      users = users.set(user.name, user);
    });
    Updater.props.users = users;

    return true;
  };

  static addRole = async (role: string): Promise<boolean> => {
    const resp = await Updater.client.addRole(role);
    // TODO: should return uid instead
    return resp.status === 200;
  };

  static delRole = async (userID: string): Promise<boolean> => {
    const resp = await Updater.client.delUser(userID);
    return resp.status === 200;
  };

  static listRoles = async (): Promise<boolean> => {
    const resp = await Updater.client.listRoles();
    if (resp.status !== 200) {
      return false;
    }

    const lsRes = resp.data as ListRolesResp;
    let roles = Set<string>();
    lsRes.roles.forEach((role: string) => {
      roles = roles.add(role);
    });
    Updater.props.roles = roles;

    return true;
  };

  static setState = (preState: ICoreState): ICoreState => {
    preState.panel.panes.admin = {
      ...preState.panel.panes.admin,
      ...Updater.props,
    };
    return preState;
  };
}

export interface State {}

export class AdminPane extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  constructor(p: Props) {
    super(p);
    Updater.init(p);
    Updater.setClient(new UsersClient(""));
    this.update = p.update;
    this.state = {};
  }

  // changeUser = (ev: React.ChangeEvent<HTMLInputElement>) => {
  //   this.setState({ user: ev.target.value });
  // };

  // changePwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
  //   this.setState({ pwd: ev.target.value });
  // };

  // initIsAuthed = () => {
  //   Updater.initIsAuthed().then(() => {
  //     this.update(Updater.setAuthPane);
  //   });
  // };

  // login = () => {
  //   Updater.login(this.state.user, this.state.pwd)
  //     .then((ok: boolean) => {
  //       if (ok) {
  //         this.update(Updater.setAuthPane);
  //         this.setState({ user: "", pwd: "" });
  //         // close all the panes
  //         PanesUpdater.displayPane("");
  //         this.update(PanesUpdater.updateState);

  //         // refresh
  //         return BrowserUpdater().setHomeItems();
  //       } else {
  //         this.setState({ user: "", pwd: "" });
  //         alert("Failed to login.");
  //       }
  //     })
  //     .then(() => {
  //       return BrowserUpdater().refreshUploadings();
  //     })
  //     .then((_: boolean) => {
  //       this.update(BrowserUpdater().setBrowser);
  //     });
  // };

  // logout = () => {
  //   Updater.logout().then((ok: boolean) => {
  //     if (ok) {
  //       this.update(Updater.setAuthPane);
  //     } else {
  //       alert("Failed to logout.");
  //     }
  //   });
  // };

  render() {
    const users = this.props.users.valueSeq().map((user: User) => {
      return (
        <div key={user.id} className="flex-list-container">
          <div className="flex-list-item-l">
            <span className="vbar blue2-bg"></span>
            <span className="bold">{`${user.id} - ${user.name}`}</span>
            {/* <input
              name={`${user.id}-role`}
              type="text"
              onChange={(ev: React.ChangeEvent<HTMLInputElement>) => {this.setRole(e, user.id )}}
              value={this.state.pwd}
              className="black0-font margin-t-m margin-b-m"
              // style={{ width: "80%" }}
              placeholder="password"
            /> */}
          </div>
          <div className="flex-list-item-r">
            <button
              onClick={() => {}}
              className="grey1-bg white-font margin-r-m"
            >
              Update
            </button>
            <button
              onClick={() => {}}
              className="grey1-bg white-font margin-r-m"
            >
              Select
            </button>
          </div>
        </div>
      );
    });

    const roles = this.props.roles.valueSeq().map((role: string) => {
      return (
        <div key={role} className="flex-list-container">
          <div className="flex-list-item-l">
            <span className="dot blue2-bg"></span>
            <span className="bold">{role}</span>
          </div>
          <div className="flex-list-item-r">
            <button
              onClick={() => {}}
              className="grey1-bg white-font margin-r-m"
            >
              Update
            </button>
            <button
              onClick={() => {}}
              className="grey1-bg white-font margin-r-m"
            >
              Select
            </button>
          </div>
        </div>
      );
    });

    return (
      <div>
        <div className="container">
          <div className="padding-l">
            <div className="flex-list-container bold">
              <span className="flex-list-item-l">
                <span className="dot black-bg"></span>
                <span>Users</span>
              </span>
              <span className="flex-list-item-r padding-r-m"></span>
            </div>
            {users}
          </div>
        </div>

        <div className="container">
          <div className="padding-l">
            <div className="flex-list-container bold">
              <span className="flex-list-item-l">
                <span className="dot black-bg"></span>
                <span>Roles</span>
              </span>
              <span className="flex-list-item-r padding-r-m"></span>
            </div>
            {roles}
          </div>
        </div>
      </div>
    );
  }
}
