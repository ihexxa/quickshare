import * as React from "react";
import { List } from "immutable";

export interface Props {
  children: List<JSX.Element>;
  childrenStyles?: List<React.CSSProperties>;
  style?: React.CSSProperties;
  className?: string;
}

const containerStyle = {
  display: "flex",
  "flex-direction": "row",
  "flex-wrap": "nowrap",
  "align-items": "center",
  "justify-content": "flex-start",
};

const childrenStyle = {
  flex: "50%",
  display: "flex",
  "align-items": "flex-start",
  "justify-content": "flex-start",
};

export const Flexbox = (props: Props) => {
  const childrenCount = props.children.size;
  const children = props.children.map(
    (child: JSX.Element, i: number): JSX.Element => {
      return (
        <div
          key={`fb-${i}`}
          style={{
            ...childrenStyle,
            flex: `${Math.floor(100 / childrenCount)}%`,
            ...(props.childrenStyles != null
              ? props.childrenStyles.get(i)
              : {}),
          }}
        >
          {child}
        </div>
      );
    }
  );

  return <div style={{ ...containerStyle, ...props.style }}>{children}</div>;
};
