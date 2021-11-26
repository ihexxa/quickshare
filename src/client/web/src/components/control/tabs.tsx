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
  color: `${colorClass("cyan0")}-font`,
};

export interface Props {
  targetSwitch: string;
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

  setTab = (targetSwitch: string, targetOption: string) => {
    if (!updater().setControlOption(targetSwitch, targetOption)) {
      alertMsg(this.props.msg.pkg.get("op.fail"));
    }
  };

  render() {
    const displaying = this.props.ui.control.controls.get(
      this.props.targetSwitch
    );
    const options = this.props.ui.control.options.get(this.props.targetSwitch);
    const tabs = options.map((option: string, targetSwitch: string) => {
      const iconProps = this.props.tabIcons.has(option)
        ? this.props.tabIcons.get(option)
        : defaultIconProps;

      // <RiFolder2Fill size="1.6rem" className="margin-r-s cyan0-font" />,
      const icon = getIcon(iconProps.name, iconProps.size, iconProps.color);

      return (
        <button
          key={`${targetSwitch}-${option}`}
          onClick={() => {
            this.setTab(targetSwitch, option);
          }}
          className="float"
        >
          <Flexbox
            children={List([
              <span className="margin-r-s">{icon}</span>,
              <span>
                {this.props.msg.pkg.get(`switch.${targetSwitch}.${option}`)}
              </span>,
            ])}
            childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
          />
        </button>
      );
    });

    return (
      <div className={`tabs control-${this.props.targetSwitch}`}>{tabs}</div>
    );
  }
}
