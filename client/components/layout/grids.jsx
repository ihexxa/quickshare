import React from "react";

const styleGridBase = {
  float: "left",
  margin: 0
};

export const Grids = props => (
  <div style={props.containerStyle}>
    {props.nodes.map(node => (
      <div
        className="grid"
        key={node.key}
        style={{ ...props.gridStyle, ...node.style }}
      >
        {node.component}
      </div>
    ))}
    <div style={{ clear: "both" }} />
  </div>
);

Grids.defaultProps = {
  nodes: [{ key: "key", component: <span />, style: {} }],
  gridStyle: styleGridBase,
  containerStyle: {}
};
