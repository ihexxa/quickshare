import * as React from "react";

import { ICoreState, init } from "./core_state";
import { Panel } from "./panel";

export interface Props {}
export interface State extends ICoreState {}

export class UpdaterBase {
  private static props: any;
  static init = (props: any) => (UpdaterBase.props = {...props});
}

export class StateMgr extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = init();
  }

  // TODO: any can be eliminated by adding union type of children states
  update = (updater: (prevState:ICoreState) => ICoreState): void => {
    console.log("before", this.state)
    this.setState(updater(this.state));
    console.log("after", this.state)
  };

  render() {
    return (
      <div>
        <Panel
          authPane = {this.state.panel.authPane}
          displaying={this.state.panel.displaying}
          update={this.update}
          browser={this.state.panel.browser}
          panes={this.state.panel.panes}
        />
      </div>
    );
  }
}
