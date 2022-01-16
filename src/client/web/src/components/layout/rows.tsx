import * as React from "react";
import { List } from "immutable";

import { BiSortUp } from "@react-icons/all-files/bi/BiSortUp";


import { Flexbox } from "./flexbox";

export interface Row {
  elem: React.ReactNode; // element to display
  val: Object; // original object value
  sortVals: List<string>; // sortable values in order
}

export interface Props {
  sortKeys: List<string>; // display names in order for sorting
  rows: List<Row>;
  id?: string;
  style?: React.CSSProperties;
  className?: string;
  updateRows?: (rows: Object) => void; // this is a callback which update state with re-sorted rows
}

export interface State {
  orders: List<boolean>; // asc = true, desc = false
}

export class Rows extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = {
      orders: p.sortKeys.map((_: string, i: number) => {
        return false;
      }),
    };
  }

  sortRows = (key: number) => {
    if (this.props.updateRows == null) {
      return;
    }
    const sortOption = this.props.sortKeys.get(key);
    if (sortOption == null) {
      return;
    }
    const currentOrder = this.state.orders.get(key);
    if (currentOrder == null) {
      return;
    }
    const expectedOrder = !currentOrder;

    const sortedRows = this.props.rows.sort((row1: Row, row2: Row) => {
      const val1 = row1.sortVals.get(key);
      const val2 = row2.sortVals.get(key);
      if (val1 == null || val2 == null) {
        // elements without the sort key will be moved to the last
        if (val1 == null && val2 != null) {
          return 1;
        } else if (val1 != null && val2 == null) {
          return -1;
        }
        return 0;
      } else if (val1 < val2) {
        return expectedOrder ? -1 : 1;
      } else if (val1 === val2) {
        return 0;
      }
      return expectedOrder ? 1 : -1;
    });

    const sortedItems = sortedRows.map((row: Row): Object => {
      return row.val;
    });
    const newOrders = this.state.orders.set(key, !currentOrder);
    this.setState({ orders: newOrders });
    this.props.updateRows(sortedItems);
  };

  render() {
    const sortBtns = this.props.sortKeys.map(
      (displayName: string, i: number): React.ReactNode => {
        return (
          <button
            key={`rows-${i}`}
            className="float"
            onClick={() => {
              this.sortRows(i);
            }}
          >
            {displayName}
          </button>
        );
      }
    );

    const bodyRows = this.props.rows.map(
      (row: Row, i: number): React.ReactNode => {
        return <div key={`rows-r-${i}`}>{row.elem}</div>;
      }
    );

    const orderByList =
      sortBtns.size > 0 ? (
        <div className="margin-b-l">
          <Flexbox
            children={List([
              <BiSortUp
                size="3rem"
                className="black-font margin-r-m"
              />,
              <span>{sortBtns}</span>,
            ])}
            childrenStyles={List([{ flex: "0 0 auto" }, { flex: "0 0 auto" }])}
          />
        </div>
      ) : null;

    return (
      <div
        id={this.props.id}
        style={this.props.style}
        className={this.props.className}
      >
        {orderByList}
        {bodyRows}
      </div>
    );
  }
}
