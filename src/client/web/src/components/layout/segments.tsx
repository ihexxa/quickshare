import * as React from "react";
import { List } from "immutable";

export interface Props {
  id?: string;
  children: List<React.ReactNode>;
  ratios: List<number>;
  dir: boolean; // true=left, false=right
  className?: string;
}

export const Segments = (props: Props) => {
  let sum = 0;
  props.ratios.forEach((ratio) => {
    sum += Math.trunc(ratio);
  });
  if (sum > 100) {
    throw `segments: ratio sum(${sum}) > 100`;
  } else if (props.children.size !== props.ratios.size) {
    throw `segments: children size(${props.children.size}) != ratio size(${props.ratios.size})`;
  }

  const children = props.children.map(
    (child: React.ReactNode, i: number): React.ReactNode => {
      const width = `${props.ratios.get(i, 0)}%`;
      return (
        <div
          key={`seg-${i}`}
          style={{ float: props.dir ? "left" : "right", width: width }}
        >
          {child}
        </div>
      );
    }
  );

  return (
    <div id={props.id}>
      <div className={props.className}>{children}</div>
      <div style={{ clear: "both" }}></div>
    </div>
  );
};
