import * as React from "react";
import { Set, Map } from "immutable";

import { ICoreState } from "./core_state";
import { PaneSettings } from "./pane_settings";
import { AuthPane, Props as AuthPaneProps } from "./pane_login";

export interface Props {
  displaying: string;
  paneNames: Set<string>;
  login: AuthPaneProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  static props: Props;

  static init = (props: Props) => (Updater.props = { ...props });

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
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  constructor(p: Props) {
    super(p);
    Updater.init(p);
    this.update = p.update;
  }

  closePane = () => {
    if (this.props.displaying !== "login") {
      Updater.displayPane("");
      this.update(Updater.updateState);
    }
  };

  render() {
    let displaying = this.props.displaying;
    if (!this.props.login.authed) {
      // TODO: use constant instead
      displaying = "login";
    }

    const panesMap: Map<string, JSX.Element> = Map({
      settings: <PaneSettings login={this.props.login} update={this.update} />,
      login: <AuthPane authed={this.props.login.authed} update={this.update} />,
    });

    const panes = panesMap.keySeq().map(
      (paneName: string): JSX.Element => {
        const isDisplay = displaying === paneName ? "" : "hidden";
        return (
          <div key={paneName} className={`${isDisplay}`}>
            {panesMap.get(paneName)}
          </div>
        );
      }
    );

    const btnClass = displaying === "login" ? "hidden" : "";
    return (
      <div id="panes" className={displaying === "" ? "hidden" : ""}>
        <div className="container">
          <div className="padding-l">
            <div className={btnClass}>
              <button onClick={this.closePane} className="black0-bg white-font">
                Return
              </button>
              <div className="hr"></div>
            </div>
            {panes}
          </div>
        </div>
      </div>
    );
  }
}
