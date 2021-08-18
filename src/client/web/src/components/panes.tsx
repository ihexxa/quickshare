import * as React from "react";
import { Set, Map } from "immutable";

import { IUsersClient, User, ListUsersResp, ListRolesResp } from "../client";
import { UsersClient } from "../client/users";
import { ICoreState } from "./core_state";
import { PaneSettings } from "./pane_settings";
import { AdminPane, Props as AdminPaneProps } from "./pane_admin";
import { AuthPane, Props as AuthPaneProps } from "./pane_login";

export interface Props {
  userRole: string;
  displaying: string;
  paneNames: Set<string>;
  login: AuthPaneProps;
  admin: AdminPaneProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  static props: Props;
  private static client: IUsersClient;

  static init = (props: Props) => (Updater.props = { ...props });
  static setClient = (client: IUsersClient): void => {
    Updater.client = client;
  };

  static displayPane = (paneName: string) => {
    if (paneName === "") {
      // hide all panes
      Updater.props.displaying = "";
    } else {
      const pane = Updater.props.paneNames.get(paneName);
      if (pane != null) {
        Updater.props.displaying = paneName;
      } else {
        alert(`dialgos: pane (${paneName}) not found`);
      }
    }
  };

  static self = async (): Promise<boolean> => {
    const resp = await Updater.client.self();
    if (resp.status === 200) {
      Updater.props.userRole = resp.data.role;
      return true;
    }
    return false;
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

  static setRole = async (userID: string, role: string): Promise<boolean> => {
    const resp = await Updater.client.delUser(userID);
    return resp.status === 200;
  };

  static forceSetPwd = async (
    userID: string,
    pwd: string
  ): Promise<boolean> => {
    const resp = await Updater.client.forceSetPwd(userID, pwd);
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
    Updater.props.admin.users = users;

    return true;
  };

  static addRole = async (role: string): Promise<boolean> => {
    const resp = await Updater.client.addRole(role);
    // TODO: should return uid instead
    return resp.status === 200;
  };

  static delRole = async (role: string): Promise<boolean> => {
    const resp = await Updater.client.delRole(role);
    return resp.status === 200;
  };

  static listRoles = async (): Promise<boolean> => {
    const resp = await Updater.client.listRoles();
    if (resp.status !== 200) {
      return false;
    }

    const lsRes = resp.data as ListRolesResp;
    let roles = Set<string>();
    Object.keys(lsRes.roles).forEach((role: string) => {
      roles = roles.add(role);
    });
    Updater.props.admin.roles = roles;

    return true;
  };

  static updateState = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panel: {
        ...prevState.panel,
        panes: { ...prevState.panel.panes, ...Updater.props },
      },
    };
  };
}

export interface State {}
export class Panes extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    Updater.init(p);
    Updater.setClient(new UsersClient(""));
  }

  closePane = () => {
    if (this.props.displaying !== "login") {
      Updater.displayPane("");
      this.props.update(Updater.updateState);
    }
  };

  render() {
    let displaying = this.props.displaying;
    if (!this.props.login.authed) {
      // TODO: use constant instead
      displaying = "login";
    }

    let panesMap: Map<string, JSX.Element> = Map({
      settings: (
        <PaneSettings login={this.props.login} update={this.props.update} />
      ),
      login: (
        <AuthPane
          authed={this.props.login.authed}
          captchaID={this.props.login.captchaID}
          update={this.props.update}
        />
      ),
    });

    if (this.props.userRole === "admin") {
      panesMap = panesMap.set(
        "admin",
        <AdminPane
          users={this.props.admin.users}
          roles={this.props.admin.roles}
          update={this.props.update}
        />
      );
    }

    const panes = panesMap.keySeq().map((paneName: string): JSX.Element => {
      const isDisplay = displaying === paneName ? "" : "hidden";
      return (
        <div key={paneName} className={`${isDisplay}`}>
          {panesMap.get(paneName)}
        </div>
      );
    });

    const btnClass = displaying === "login" ? "hidden" : "";
    return (
      <div id="panes" className={displaying === "" ? "hidden" : ""}>
        <div className="root-container">
          <div className="container">
            <div className="flex-list-container padding-l">
              <h3 className="flex-list-item-l txt-cap">{displaying}</h3>
              <div className="flex-list-item-r">
                <button
                  onClick={this.closePane}
                  className={`red0-bg white-font ${btnClass}`}
                >
                  Close
                </button>
              </div>
            </div>
          </div>

          {panes}
        </div>
        {/* <div className="hr white0-bg margin-b-m margin-l-m margin-r-m"></div> */}
        {/* <div className="padding-l"></div> */}
      </div>
    );
  }
}
