import * as React from "react";
import { List, Map, update } from "immutable";

import { updater } from "../state_updater";
import { getIcon } from "../visual/icons";
import { Flexbox } from "../layout/flexbox";
import { ICoreState, MsgProps, UIProps } from "../core_state";
import { AdminProps } from "../pane_admin";
import { LoginProps } from "../pane_login";
import { alertMsg } from "../../common/env";
import { IconProps } from "../visual/icons";
import { colorClass } from "../visual/colors";

const defaultIconProps: IconProps = {
  name: "RiFolder2Fill",
  size: "1.6rem",
  color: `${colorClass("cyan1")}-font`,
};

export interface Props {
  targetControl: string;
  tabIcons: Map<string, IconProps>; // option name -> icon name
  login: LoginProps;
  admin: AdminProps;
  ui: UIProps;
  msg: MsgProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}
export class Tabs extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  setTab = (targetControl: string, targetOption: string) => {
    if (!updater().setControlOption(targetControl, targetOption)) {
      alertMsg(this.props.msg.pkg.get("op.fail"));
    }
    this.props.update(updater().updateUI);
  };

  render() {
    const displaying = this.props.ui.control.controls.get(
      this.props.targetControl
    );

    const options = this.props.ui.control.options.get(this.props.targetControl);
    const tabs = options.map((option: string) => {
      const iconProps = this.props.tabIcons.has(option)
        ? this.props.tabIcons.get(option)
        : defaultIconProps;

      const iconColor = displaying === option ? iconProps.color : "black0";
      const icon = getIcon(iconProps.name, iconProps.size, iconColor);
      const fontColor =
        displaying === option ? `${colorClass(iconColor)}-font` : "";

      return (
        <button
          key={`${this.props.targetControl}-${option}`}
          onClick={() => {
            this.setTab(this.props.targetControl, option);
          }}
          className="float-l"
        >
          <Flexbox
            children={List([
              <span className="margin-r-s">{icon}</span>,
              <span className={fontColor}>
                {this.props.msg.pkg.get(
                  `control.${this.props.targetControl}.${option}`
                )}
              </span>,
            ])}
            childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
          />
        </button>
      );
    });

    return (
      <div className={`tabs control-${this.props.targetControl}`}>{tabs}</div>
    );
  }
}
