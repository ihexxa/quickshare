import axios from "axios";
import React from "react";
import ReactDOM from "react-dom";

import { config } from "../config";
import { addLocalFiles, list } from "../libs/api_share";
import { login, logout } from "../libs/api_auth";
import { FilePane } from "../components/composite/file_pane";
import { InfoBar } from "../components/composite/info_bar";
import { Log } from "../components/composite/log";

function getWidth() {
  if (window.innerWidth >= window.innerHeight) {
    return `${Math.floor(
      (window.innerWidth * 0.95) / config.rootSize / config.colWidth
    ) * config.colWidth}rem`;
  }
  return "auto";
}

const styleLogContainer = {
  paddingTop: "1rem",
  textAlign: "center",
  height: "2rem",
  overflowX: "hidden" // TODO: should no hidden
};

const styleLogContent = {
  color: "#333",
  fontSize: "0.875rem",
  opacity: 0.6,
  backgroundColor: "#fff",
  borderRadius: "1rem",
  whiteSpace: "nowrap"
};

class AdminPanel extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      isLogin: false,
      filterName: "",
      serverAddr: `${window.location.protocol}//${window.location.hostname}:${
        window.location.port
      }`,
      width: getWidth()
    };
    this.log = {
      ok: msg => console.log(msg),
      warning: msg => console.log(msg),
      info: msg => console.log(msg),
      error: msg => console.log(msg),
      start: msg => console.log(msg),
      end: msg => console.log(msg)
    };
    this.logComponent = <Log ref={this.assignLog} styleLog={styleLogContent} />;
  }

  componentWillMount() {
    list().then(infos => {
      if (infos != null) {
        this.setState({ isLogin: true });
      }
    });
  }

  setWidth = () => {
    this.setState({ width: getWidth() });
  };

  // componentDidMount() {
  //   window.addEventListener("resize", this.setWidth);
  // }

  // componentWillUnmount() {
  //   window.removeEventListener("resize", this.setWidth);
  // }

  onLogin = (serverAddr, adminId, adminPwd) => {
    login(serverAddr, adminId, adminPwd).then(ok => {
      if (ok === true) {
        this.setState({ isLogin: true });
      } else {
        this.log.error("Fail to login");
        this.setState({ isLogin: false });
      }
    });
  };

  onLogout = serverAddr => {
    logout(serverAddr).then(ok => {
      if (ok === false) {
        this.log.error("Fail to log out");
      } else {
        this.log.ok("You are logged out");
      }
      this.setState({ isLogin: false });
    });
  };

  onSearch = fileName => {
    this.setState({ filterName: fileName });
  };

  assignLog = logRef => {
    this.log = logRef;
    this.log.info(
      <span>
        Know more about <a href="https://github.com/ihexxa">Quickshare</a>
      </span>
    );
  };

  render() {
    const width = this.state.width;

    return (
      <div>
        <InfoBar
          compact={width === "auto"}
          width={width}
          serverAddr={this.state.serverAddr}
          isLogin={this.state.isLogin}
          onLogin={this.onLogin}
          onLogout={this.onLogout}
          onAddLocalFiles={addLocalFiles}
          onSearch={this.onSearch}
          onOk={this.log.ok}
          onError={this.log.error}
        >
          <div style={{ ...styleLogContainer, width }}>{this.logComponent}</div>
        </InfoBar>
        {this.state.isLogin ? (
          <FilePane
            width={width}
            colWidth={config.colWidth}
            onList={list}
            onOk={this.log.ok}
            onError={this.log.error}
            filterName={this.state.filterName}
          />
        ) : (
          <div />
        )}
      </div>
    );
  }
}

ReactDOM.render(<AdminPanel />, document.getElementById("app"));
