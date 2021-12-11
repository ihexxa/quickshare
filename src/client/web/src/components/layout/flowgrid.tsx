import * as React from "react";
import { List } from "immutable";

export interface Props {
  grids: List<React.ReactNode>;
  style?: React.CSSProperties;
  className?: string;
}

export const Flowgrid = (props: Props) => {
  const children = props.grids.map(
    (child: React.ReactNode, i: number): React.ReactNode => {
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
