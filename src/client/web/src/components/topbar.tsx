import * as React from "react";
import { List } from "immutable";
import { RiGithubFill } from "@react-icons/all-files/ri/RiGithubFill";

import { ICoreState, MsgProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { updater } from "./state_updater";
import { Flexbox } from "./layout/flexbox";

export interface State {}
export interface Props {
  login: LoginProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class TopBar extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  showSettings = () => {
    updater().displayPane("settings");
    this.props.update(updater().updatePanes);
  };

  showAdmin = async () => {
    return updater()
      .self()
      .then(() => {
        // TODO: remove hardcode role
        if (this.props.login.authed && this.props.login.userRole === "admin") {
          return Promise.all([updater().listRoles(), updater().listUsers()]);
        }
      })
      .then(() => {
        updater().displayPane("admin");
        this.props.update(updater().updateAdmin);
        this.props.update(updater().updatePanes);
      });
  };

  render() {
    const adminBtn =
      this.props.login.userRole === "admin" ? (
        <button
          onClick={this.showAdmin}
          className="grey3-bg grey4-font margin-r-m"
          style={{ minWidth: "7rem" }}
        >
          {this.props.msg.pkg.get("admin")}
        </button>
      ) : null;

    return (
      <div
        id="top-bar"
        className="top-bar cyan1-font padding-t-m padding-b-m padding-l-l padding-r-l"
      >
        <Flexbox
          children={List([
            <a
              href="https://github.com/ihexxa/quickshare"
              target="_blank"
              className="h5"
            >
              Quickshare
              {/* <RiGithubFill size="2rem" className="grey4-font margin-r-m" /> */}
            </a>,

            <Flexbox
              children={List([
                <span>
                  <span className="grey3-font font-s">
                    {this.props.login.userName}
                  </span>
                  &nbsp;-&nbsp;
                  <span className="grey0-font font-s margin-r-m">
                    {this.props.login.userRole}
                  </span>
                </span>,

                <button
                  onClick={this.showSettings}
                  className="grey3-bg grey4-font margin-r-m"
                  style={{ minWidth: "7rem" }}
                >
                  {this.props.msg.pkg.get("settings")}
                </button>,

                adminBtn,
              ])}
              childrenStyles={List([{}, {}, {}, {}])}
            />,
          ])}
          childrenStyles={List([
            {},
            { justifyContent: "flex-end", alignItems: "center" },
          ])}
        />
        {/* <div className="flex-2col-parent">
          <a
            href="https://github.com/ihexxa/quickshare"
            className="flex-13col h5"
          >
            Quickshare
          </a>
          <span className="flex-23col text-right">
            <span className="grey3-font font-s">
              {this.props.login.userName}
            </span>
            &nbsp;-&nbsp;
            <span className="grey0-font font-s margin-r-m">
              {this.props.login.userRole}
            </span>
            <button
              onClick={this.showSettings}
              className="grey3-bg grey4-font margin-r-m"
            >
              {this.props.msg.pkg.get("settings")}
            </button>
            {adminBtn}

            <RiGithubFill
              size="2rem"
              className="grey4-font margin-r-m"
            />
          </span>
        </div> */}
      </div>
    );
  }
}
