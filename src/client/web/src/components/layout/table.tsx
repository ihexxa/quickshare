import * as React from "react";
import { List } from "immutable";

import { updater } from "../state_updater";

export interface Cell {
  elem: React.ReactNode;
  val: string; // original cell value
  // sortVal: string; // this value is for sorting
}

export interface Row {
  cells: List<Cell>;
  val: Object; // original object value
}

export interface Head {
  elem: React.ReactNode;
  sortable: boolean;
}

export interface Props {
  head: List<Head>;
  rows: List<Row>;
  foot: List<React.ReactNode>;
  colStyles?: List<React.CSSProperties>;
  id?: string;
  style?: React.CSSProperties;
  className?: string;
  originalVals?: List<Object>;
  updateRows?: (rows: Object) => void;
}

export interface State {
  orders: List<boolean>; // asc = true, desc = false
}

export class Table extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = {
      // TODO: if the size of column increases
      // state will be out of sync with props
      // so this will not work
      orders: p.head.map((_, i: number) => {
        return i === 0; // order by the first column
      }),
    };
  }

  sortRows = (colIndex: number) => {
    if (this.props.updateRows == null) {
      return;
    }
    const headCell = this.props.head.get(colIndex);
    if (headCell == null || !headCell.sortable) {
      return;
    }
    const currentOrder = this.state.orders.get(colIndex);
    if (currentOrder == null) {
      return;
    }

    const sortedRows = this.props.rows.sort((row1: Row, row2: Row) => {
      const cell1 = row1.cells.get(colIndex);
      const cell2 = row2.cells.get(colIndex);

      if (cell1 == null || cell2 == null) {
        // keep current order
        return currentOrder ? -1 : 1;
      } else if (cell1.val < cell2.val) {
        return -1;
      } else if (cell1.val == cell2.val) {
        return 0;
      } else {
        return currentOrder ? 1 : -1;
      }
    });

    const sortedItems = sortedRows.map((row: Row): Object => {
      return row.val;
    });

    const newOrders = this.state.orders.set(colIndex, !currentOrder);
    this.setState({ orders: newOrders });
    this.props.updateRows(sortedItems);
  };

  render() {
    const headCols = this.props.head.map(
      (head: Head, i: number): React.ReactNode => {
        const style =
          this.props.colStyles != null ? this.props.colStyles.get(i) : {};

        return (
          <th
            key={`h-${i}`}
            className="title-xs clickable"
            style={style}
            onClick={() => {
              this.sortRows(i);
            }}
          >
            {head.elem}
          </th>
        );
      }
    );

    const bodyRows = this.props.rows.map(
      (row: Row, i: number): React.ReactNode => {
        const tds = row.cells.map((cell: Cell, j: number) => {
          const style =
            this.props.colStyles != null ? this.props.colStyles.get(j) : {};
          return (
            <td key={`rc-${i}-${j}`} style={style}>
              {cell.elem}
            </td>
          );
        });
        return <tr key={`r-${i}`}>{tds}</tr>;
      }
    );

    const footCols = this.props.foot.map(
      (elem: React.ReactNode, i: number): React.ReactNode => {
        const style =
          this.props.colStyles != null ? this.props.colStyles.get(i) : {};
        return (
          <th key={`f-${i}`} style={style}>
            {elem}
          </th>
        );
      }
    );

    return (
      <table
        id={this.props.id}
        style={this.props.style}
        className={this.props.className}
      >
        <thead>
          <tr>{headCols}</tr>
        </thead>
        <tbody>{bodyRows}</tbody>
        <tfoot>
          <tr>{footCols}</tr>
        </tfoot>
      </table>
    );
  }
}

// export const Table = (props: Props) => {
//   const headCols = props.head.map((head: Head, i: number): React.ReactNode => {
//     const style = props.colStyles != null ? props.colStyles.get(i) : {};
//     return (
//       <th key={`h-${i}`} className="title-xs clickable" style={style}>
//         {head.elem}
//       </th>
//     );
//   });

//   const bodyRows = props.rows.map(
//     (row: List<Cell>, i: number): React.ReactNode => {
//       const tds = row.map((cell: Cell, j: number) => {
//         const style = props.colStyles != null ? props.colStyles.get(j) : {};
//         return (
//           <td key={`rc-${i}-${j}`} style={style}>
//             {cell.elem}
//           </td>
//         );
//       });
//       return <tr key={`r-${i}`}>{tds}</tr>;
//     }
//   );

//   const footCols = props.foot.map(
//     (elem: React.ReactNode, i: number): React.ReactNode => {
//       const style = props.colStyles != null ? props.colStyles.get(i) : {};
//       return (
//         <th key={`f-${i}`} style={style}>
//           {elem}
//         </th>
//       );
//     }
//   );

//   return (
//     <table id={props.id} style={props.style} className={props.className}>
//       <thead>
//         <tr>{headCols}</tr>
//       </thead>
//       <tbody>{bodyRows}</tbody>
//       <tfoot>
//         <tr>{footCols}</tr>
//       </tfoot>
//     </table>
//   );
// };
