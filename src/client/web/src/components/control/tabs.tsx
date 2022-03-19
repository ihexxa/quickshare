import * as React from "react";
import { List, Map } from "immutable";

import { updater } from "../state_updater";
import { Flexbox } from "../layout/flexbox";
import { ICoreState, MsgProps, UIProps } from "../core_state";
import { alertMsg } from "../../common/env";
import { IconProps, getIcon } from "../visual/icons";
import { colorClass } from "../visual/colors";

const defaultIconProps: IconProps = {
  name: "RiFolder2Fill",
  size: "1.6rem",
  color: `${colorClass("cyan1")}`,
};

export interface Props {
  targetControl: string;
  tabIcons: Map<string, IconProps>; // option name -> icon name
  titleIcon?: string;
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

    const titleIcon =
      this.props.titleIcon != null
        ? getIcon(this.props.titleIcon, "2rem", "normal")
        : null;
    const options = this.props.ui.control.options.get(this.props.targetControl);
    const tabs = options.map((option: string) => {
      const iconProps = this.props.tabIcons.has(option)
        ? this.props.tabIcons.get(option)
        : defaultIconProps;

      const iconColor = displaying === option ? iconProps.color : "normal";
      const icon = getIcon(iconProps.name, iconProps.size, iconColor);
      const fontColor =
        displaying === option ? `${colorClass(iconColor)}-font` : "normal-font";

      return (
        <button
          key={`${this.props.targetControl}-${option}`}
          onClick={() => {
            this.setTab(this.props.targetControl, option);
          }}
          className="float-l margin-r-m minor-bg"
        >
          <div className="float-l icon-s margin-r-s">{icon}</div>
          <div className={`float-l font-xs ${fontColor}`}>
            {this.props.msg.pkg.get(
              `control.${this.props.targetControl}.${option}`
            )}
          </div>
          <div className="fix"></div>
        </button>
      );
    });

    return (
      <div className={`tabs control-${this.props.targetControl}`}>
        {tabs}
        <div className="fix"></div>
      </div>
    );
  }
}
