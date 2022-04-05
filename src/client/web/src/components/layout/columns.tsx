import * as React from "react";
import { List } from "immutable";

export interface Props {
  rows: List<List<React.ReactNode>>;
  widths: List<string>;
  childrenClassNames?: List<string>;
  style?: React.CSSProperties;
  className?: string;
  colKey?: string;
}

export const Columns = (props: Props) => {
  const children = props.rows.map(
    (row: List<React.ReactNode>, i: number): React.ReactNode => {
      const cells = row.map((cell: React.ReactNode, j: number) => {
        const width = props.widths.get(j, Math.trunc(100 / row.size));
        const className = props.childrenClassNames.get(j, "");

        return (
          <div
            key={`${props.colKey}-${i}-${j}`}
            className={`float-l ${className}`}
            style={{ width }}
          >
            {cell}
          </div>
        );
      });

      return (
        <div key={`${props.colKey}-${i}`}>
          {cells}
          <div className="fix"></div>
        </div>
      );
    }
  );

  return (
    <div style={props.style} className={props.className}>
      {children}
    </div>
  );
};
