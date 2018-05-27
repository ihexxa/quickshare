import React from "react";

import { Button } from "../control/button";
import { Input } from "../control/input";
import { getIcon, getIconColor } from "../display/icon";
import { AuthPane } from "./auth_pane";
import { rootSize } from "../../config";

let styleInfoBar = {
  textAlign: "left",
  color: "#999",
  marginBottom: "1rem",
  margin: "auto"
};

const styleContainer = {
  padding: "0.5rem",
  backgroundColor: "rgba(255, 255, 255, 0.5)"
};

const styleLeft = {
  float: "left",
  width: "50%",
  heigth: "2rem"
};

const styleRight = {
  float: "right",
  width: "50%",
  textAlign: "right",
  heigth: "2rem"
};

const styleButtonLabel = {
  verticalAlign: "middle"
};

const IconPlusCir = getIcon("pluscir");
const IconSearch = getIcon("search");
const clear = <div style={{ clear: "both" }} />;

export class InfoBar extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      filterFileName: "",
      fold: this.props.compact
    };
  }

  onLogin = (serverAddr, adminId, adminPwd) => {
    this.props.onLogin(serverAddr, adminId, adminPwd);
  };

  onLogout = serverAddr => {
    this.props.onLogout(serverAddr);
  };

  onSearch = value => {
    // TODO: need debounce
    this.props.onSearch(value);
    this.setState({ filterFileName: value });
  };

  onAddLocalFiles = () => {
    return this.props.onAddLocalFiles().then(ok => {
      if (ok) {
        // TODO: need to add refresh
        this.props.onOk("Local files are added, please refresh.");
      } else {
        this.props.onError("Fail to add local files");
      }
    });
  };

  onToggle = () => {
    this.setState({ fold: !this.state.fold });
  };

  render() {
    styleInfoBar = { ...styleInfoBar, width: this.props.width };

    if (this.props.compact) {
      const IconMore = getIcon("bars");

      const menuIcon = (
        <div style={{ backgroundColor: "rgba(255, 255, 255, 0.5)" }}>
          <div>
            <div style={{ float: "right" }}>
              <Button
                onClick={this.onToggle}
                label={""}
                // styleLabel={styleButtonLabel}
                styleContainer={{
                  height: "2.5rem",
                  width: "2.5rem",
                  backgroundColor: "rgba(255, 255, 255, 0.2)",
                  margin: "0.5rem"
                }}
                styleDefault={{ height: "auto" }}
                styleIcon={{
                  lineHeight: "1rem",
                  height: "1rem",
                  margin: "0.75rem",
                  display: "inline-block"
                }}
                icon={<IconMore size={rootSize} style={{ color: "#fff" }} />}
              />
            </div>
            <div style={{ float: "right" }}>
              <Input
                onChange={this.onSearch}
                placeholder="Search..."
                type="text"
                value={this.state.filterFileName}
                styleContainer={{
                  backgroundColor: "rgba(255, 255, 255, 0.5)",
                  margin: "0.5rem",
                  textAlign: "left"
                }}
                icon={<IconSearch size={16} />}
              />
            </div>
            <div style={{ clear: "both" }} />
          </div>
        </div>
      );
      const menuList = !this.state.fold ? (
        <div style={styleContainer}>
          <div>
            <AuthPane
              onLogin={this.onLogin}
              onLogout={this.onLogout}
              isLogin={this.props.isLogin}
              serverAddr={this.props.serverAddr}
              compact={this.props.compact}
            />
            <Button
              onClick={this.onAddLocalFiles}
              label={"Scan Files"}
              styleLabel={styleButtonLabel}
              styleContainer={{
                backgroundColor: "#2ecc71",
                marginLeft: "0.5rem"
              }}
              styleDefault={{ color: "#fff" }}
              icon={<IconPlusCir size={16} style={{ color: "#fff" }} />}
            />
          </div>
        </div>
      ) : (
        <span />
      );

      const menu = (
        <div>
          {menuIcon}
          {menuList}
        </div>
      );
      return (
        <div
          className="info-bar"
          style={{ ...styleInfoBar, textAlign: "right" }}
        >
          {this.props.isLogin ? (
            menu
          ) : (
            <AuthPane
              onLogin={this.onLogin}
              onLogout={this.onLogout}
              isLogin={this.props.isLogin}
              serverAddr={this.props.serverAddr}
              styleContainer={{ textAlign: "left" }}
              compact={this.props.compact}
            />
          )}
          <div>{this.props.children}</div>
        </div>
      );
    }

    const visitorPane = (
      <AuthPane
        onLogin={this.onLogin}
        onLogout={this.onLogout}
        isLogin={this.props.isLogin}
        serverAddr={this.props.serverAddr}
        compact={this.props.compact}
      />
    );

    const memberPane = (
      <div>
        <div style={styleLeft}>
          <AuthPane
            onLogin={this.onLogin}
            onLogout={this.onLogout}
            isLogin={this.props.isLogin}
            serverAddr={this.props.serverAddr}
          />
          <Button
            onClick={this.onAddLocalFiles}
            label={"Scan Files"}
            styleLabel={styleButtonLabel}
            styleContainer={{
              backgroundColor: "#2ecc71",
              marginLeft: "0.5rem"
            }}
            styleDefault={{ color: "#fff" }}
            icon={<IconPlusCir size={16} style={{ color: "#fff" }} />}
          />
        </div>
        <div style={styleRight}>
          <Input
            onChange={this.onSearch}
            placeholder="Search..."
            type="text"
            value={this.state.filterFileName}
            styleContainer={{ backgroundColor: "rgba(255, 255, 255, 0.5)" }}
            icon={<IconSearch size={16} />}
          />
        </div>
        {clear}
      </div>
    );

    return (
      <div className="info-bar" style={styleInfoBar}>
        <div style={styleContainer}>
          {this.props.isLogin ? memberPane : visitorPane}
        </div>
        <div>{this.props.children}</div>
      </div>
    );
  }
}

InfoBar.defaultProps = {
  compact: false,
  width: "-1",
  isLogin: false,
  serverAddr: "",
  onLogin: () => console.error("undefined"),
  onLogout: () => console.error("undefined"),
  onAddLocalFiles: () => console.error("undefined"),
  onSearch: () => console.error("undefined"),
  onOk: () => console.error("undefined"),
  onError: () => console.error("undefined")
};
