import * as React from "react";
import { List } from "immutable";

export interface Props {
  grids: List<JSX.Element>;
  style?: React.CSSProperties;
  className?: string;
}

export const Flowgrid = (props: Props) => {
  const children = props.grids.map(
    (child: JSX.Element, i: number): JSX.Element => {
      return (
        <span key={`flowgrid-${i}`} className="inline-block">
          {child}
        </span>
      );
    }
  );

  return (
    <div style={props.style} className={props.className}>
      {children}
    </div>
  );
};
