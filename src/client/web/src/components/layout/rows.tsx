import * as React from "react";
import { List } from "immutable";

export interface Row {
  elem: React.ReactNode; // element to display
  val: Object; // original object value
  sortVals: List<string>; // sortable values in order
}

export interface Props {
  rows: List<React.ReactNode>;
  id?: string;
  style?: React.CSSProperties;
  className?: string;
}

export interface State {}

export class Rows extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  render() {
    const bodyRows = this.props.rows.map(
      (row: React.ReactNode, i: number): React.ReactNode => {
        return <div key={`rows-r-${i}`}>{row}</div>;
      }
    );

    return (
      <div
        id={this.props.id}
        style={this.props.style}
        className={this.props.className}
      >
        {bodyRows}
      </div>
    );
  }
}
