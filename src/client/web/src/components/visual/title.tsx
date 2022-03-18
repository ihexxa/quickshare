import * as React from "react";
import { List } from "immutable";

import { Flexbox } from "../layout/flexbox";
import { getIconWithProps } from "./icons";
import { colorClass } from "./colors";

export interface Props {
  iconName: string;
  iconColor: string;
  title: string;
}

export const Title = (props: Props) => {
  return (
    <Flexbox
      children={List([
        <Flexbox
          children={List([
            getIconWithProps(props.iconName, {
              size: "3.2rem",
              className: `margin-r-m ${colorClass(props.iconColor)}-font`,
            }),
            <span className="title-m bold">{props.title}</span>,
          ])}
        />,

        <span></span>,
      ])}
    />
  );
};
