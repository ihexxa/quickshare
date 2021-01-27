import * as React from "react";

import { ICoreState } from "./core_state";
import { IUsersClient } from "../client";
import { AuthPane, Props as LoginProps } from "./pane_login";
import { Layouter } from "./layouter";

export interface Props {
  login: LoginProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;
  private static usersClient: IUsersClient;

  static init = (props: Props) => (Updater.props = { ...props });
  static setClient(usersClient: IUsersClient) {
    Updater.usersClient = usersClient;
  }

  static setPwd = async (oldPwd: string, newPwd: string): Promise<boolean> => {
    const resp = await Updater.usersClient.setPwd(oldPwd, newPwd);
    return resp.status === 200;
  };

  static updateState = (prevState: ICoreState): ICoreState => {
    return {
      ...prevState,
      panel: { ...prevState.panel, ...Updater.props },
    };
  };
}

export interface State {
  oldPwd: string;
  newPwd1: string;
  newPwd2: string;
}

export class PaneSettings extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  changeOldPwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ oldPwd: ev.target.value });
  };
  changeNewPwd1 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd1: ev.target.value });
  };
  changeNewPwd2 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd2: ev.target.value });
  };

  constructor(p: Props) {
    super(p);
    // Updater.init(p);
    this.update = p.update;
    this.state = {
      oldPwd: "",
      newPwd1: "",
      newPwd2: "",
    };
  }

  setPwd = () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      alert("new passwords are not same");
    } else if (this.state.newPwd1 == "") {
      alert("new passwords can not be empty");
    } else if (this.state.oldPwd == this.state.newPwd1) {
      alert("old and new passwords are same");
    } else {
      Updater.setPwd(this.state.oldPwd, this.state.newPwd1).then(
        (ok: boolean) => {
          if (ok) {
            alert("Password is updated");
          } else {
            alert("fail to update password");
          }
          this.setState({
            oldPwd: "",
            newPwd1: "",
            newPwd2: "",
          });
        }
      );
    }
  };

  render() {
    const inputs: Array<JSX.Element> = [
      <input
        name="old_pwd"
        type="password"
        onChange={this.changeOldPwd}
        value={this.state.oldPwd}
        className="black0-font"
        placeholder="old password"
      />,
      <input
        name="new_pwd1"
        type="password"
        onChange={this.changeNewPwd1}
        value={this.state.newPwd1}
        className="black0-font"
        placeholder="new password"
      />,
      <input
        name="new_pwd2"
        type="password"
        onChange={this.changeNewPwd2}
        value={this.state.newPwd2}
        className="black0-font"
        placeholder="new password again"
      />,
      <button onClick={this.setPwd} className="grey1-bg white-font">
        Update
      </button>,
    ];

    return (
      <div>
        <h4 className="grey0-font">Update Password</h4>
        <Layouter isHorizontal={false} elements={inputs}></Layouter>
        <div className="hr"></div>

        <AuthPane authed={this.props.login.authed} update={this.update} />
        {/* <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br />
        <br /> */}
      </div>
    );
  }
}
