import React from "react";
import { Button } from "../control/button";
import { Input } from "../control/input";

import { config } from "../../config";
import { getIcon } from "../display/icon";
import { makePostBody } from "../../libs/utils";
import { styleButtonLabel } from "./info_bar";

export const classLogin = "auth-pane-login";
export const classLogout = "auth-pane-logout";
const IconSignIn = getIcon("signIn");
const IconSignOut = getIcon("signOut");
const IconAngRight = getIcon("angRight");

export class AuthPane extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      adminId: "",
      adminPwd: ""
    };
  }

  onLogin = e => {
    e.preventDefault();
    this.props.onLogin(
      this.props.serverAddr,
      this.state.adminId,
      this.state.adminPwd
    );
  };

  onLogout = () => {
    this.props.onLogout(this.props.serverAddr);
  };

  onChangeAdminId = adminId => {
    this.setState({ adminId });
  };

  onChangeAdminPwd = adminPwd => {
    this.setState({ adminPwd });
  };

  render() {
    if (this.props.isLogin) {
      return (
        <span className={classLogout} style={this.props.styleContainer}>
          <Button
            onClick={this.onLogout}
            icon={<IconSignOut size={config.rootSize} />}
            label={"Logout"}
            styleLabel={styleButtonLabel}
            styleDefault={{ color: "#666" }}
            styleContainer={{ backgroundColor: "#ccc" }}
          />
        </span>
      );
    } else {
      if (this.props.compact) {
        return (
          <form
            onSubmit={this.onLogin}
            className={classLogin}
            style={this.props.styleContainer}
          >
            <Input
              placeholder="user name"
              type="text"
              onChange={this.onChangeAdminId}
              value={this.state.adminId}
              styleContainer={{ margin: "0.5rem" }}
              icon={<IconAngRight size={config.rootSize} />}
            />
            <Input
              placeholder="password"
              type="password"
              onChange={this.onChangeAdminPwd}
              value={this.state.adminPwd}
              styleContainer={{ margin: "0.5rem" }}
              icon={<IconAngRight size={config.rootSize} />}
            />
            <Button
              type="submit"
              icon={<IconSignIn size={config.rootSize} />}
              label={"login"}
              styleLabel={styleButtonLabel}
              styleDefault={{ color: "#fff" }}
              styleContainer={{
                backgroundColor: "#2c3e50",
                marginLeft: "0.5rem"
              }}
            />
          </form>
        );
      } else {
        return (
          <form
            onSubmit={this.onLogin}
            className={classLogin}
            style={this.props.styleContainer}
          >
            <Input
              placeholder="user name"
              type="text"
              onChange={this.onChangeAdminId}
              value={this.state.adminId}
              icon={<IconAngRight size={config.rootSize} />}
            />
            <Input
              placeholder="password"
              type="password"
              onChange={this.onChangeAdminPwd}
              value={this.state.adminPwd}
              styleContainer={{ marginLeft: "0.5rem" }}
              icon={<IconAngRight size={config.rootSize} />}
            />
            <Button
              type="submit"
              icon={<IconSignIn size={config.rootSize} />}
              label={"login"}
              styleLabel={styleButtonLabel}
              styleDefault={{ color: "#fff" }}
              styleContainer={{
                backgroundColor: "#2c3e50",
                marginLeft: "0.5rem"
              }}
            />
          </form>
        );
      }
    }
  }
}

AuthPane.defaultProps = {
  onLogin: () => console.error("undefined"),
  onLogout: () => console.error("undefined"),
  compact: false,
  isLogin: false,
  serverAddr: "",
  styleContainer: {},
  styleStr: ""
};
