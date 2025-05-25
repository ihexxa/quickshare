import * as React from "react";
import { List, Map } from "immutable";

import { updater } from "../state_updater";
import { ICoreState, MsgProps, UIProps } from "../core_state";
import { Env } from "../../common/env";
import { IconProps, getIcon } from "../visual/icons";
import { colorClass } from "../visual/colors";
import { Flexbox } from "../layout/flexbox";

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
      Env().alertMsg(this.props.msg.pkg.get("op.fail"));
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
      const fontWeight =
        displaying === option ? `font-bold` : "";

      return (
        <button
          key={`${this.props.targetControl}-${option}`}
          onClick={() => {
            this.setTab(this.props.targetControl, option);
          }}
          className="float-left mr-12"
        >
          <div className="float-left icon-s margin-r-s">{icon}</div>
          <div className={`float-left font-xs ${fontWeight}`}>
            {this.props.msg.pkg.get(
              `control.${this.props.targetControl}.${option}`
            )}
          </div>
          <div className="fix"></div>
        </button>
      );
    });

    return (
      <Flexbox
        children={List([
          <div className={`tabs control-${this.props.targetControl}`}>
            {tabs}
            <div className="fix"></div>
          </div>,
          <a href="//github.com/ihexxa/quickshare" className="major-font">
            {getIcon("FaGithub", "2rem", "normal")}
          </a>,
        ])}
        childrenStyles={List([
          { flex: "0 0 auto" },
          { justifyContent: "flex-end" },
        ])}
      />
    );
  }
}
