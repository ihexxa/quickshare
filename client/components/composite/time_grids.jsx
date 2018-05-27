import React from "react";
import { config } from "../../config";
import { Grids } from "../layout/grids";

const styleTitle = {
  color: "#fff",
  backgroundColor: "rgba(0, 0, 0, 0.4)",
  display: "inline-block",
  padding: "0.5rem 1rem",
  fontSize: "1rem",
  margin: "2rem 0 0.5rem 0",
  lineHeight: "1rem",
  height: "1rem"
};

export class TimeGrids extends React.PureComponent {
  render() {
    const groups = new Map();

    this.props.items.forEach(item => {
      const date = new Date(item.timestamp);
      const key = `${date.getFullYear()}-${date.getMonth() +
        1}-${date.getDate()}`;

      if (groups.has(key)) {
        groups.set(key, [...groups.get(key), item]);
      } else {
        groups.set(key, [item]);
      }
    });

    var timeGrids = [];
    groups.forEach((gridGroup, groupKey) => {
      const year = parseInt(groupKey.split("-")[0]);
      const month = parseInt(groupKey.split("-")[1]);
      const date = parseInt(groupKey.split("-")[2]);

      const sortedGroup = gridGroup.sort((item1, item2) => {
        return item2.timestamp - item1.timestamp;
      });

      timeGrids = [
        ...timeGrids,
        <div key={year * 365 + month * 30 + date}>
          <div style={styleTitle}>
            <span>{groupKey}</span>
          </div>
          <Grids nodes={sortedGroup} />
        </div>
      ];
    });

    const sortedGroups = timeGrids.sort((group1, group2) => {
      return group2.key - group1.key;
    });
    return <div style={this.props.styleContainer}>{sortedGroups}</div>;
  }
}

TimeGrids.defaultProps = {
  items: [
    {
      key: "",
      timestamp: -1,
      component: <span>no grid found</span>
    }
  ],
  styleContainer: {}
};
