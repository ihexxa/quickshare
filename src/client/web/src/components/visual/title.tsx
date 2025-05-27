import * as React from "react";
import { List } from "immutable";

import { Flexbox } from "../layout/flexbox";

export interface Props {
  iconName: string;
  iconColor: string;
  title: string;
}

export const Title = (props: Props) => {
  return (
    <Flexbox
      children={List([
        <div>
          {/* <div className="float-l icon-l">
            {getIconWithProps(props.iconName, {
              size: iconSize("l"),
              className: `margin-r-m ${colorClass(props.iconColor)}-font`,
            })}
          </div> */}
          <div className="text-left leading-default text-2xl">{props.title}</div>
        </div>,
        <span></span>,
      ])}
    />
  );
};
