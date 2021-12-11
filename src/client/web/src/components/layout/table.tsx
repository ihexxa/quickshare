import * as React from "react";
import { List } from "immutable";

export interface Props {
  head: List<React.ReactNode>;
  rows: List<List<React.ReactNode>>;
  foot: List<React.ReactNode>;
  colStyles?: List<React.CSSProperties>;
  id?: string;
  style?: React.CSSProperties;
  className?: string;
}

export const Table = (props: Props) => {
  const headCols = props.head.map(
    (elem: React.ReactNode, i: number): React.ReactNode => {
      const style = props.colStyles != null ? props.colStyles.get(i) : {};
      return (
        <th key={`h-${i}`} style={style}>
          {elem}
        </th>
      );
    }
  );
  const bodyRows = props.rows.map(
    (row: List<React.ReactNode>, i: number): React.ReactNode => {
      const tds = row.map((elem: React.ReactNode, j: number) => {
        const style = props.colStyles != null ? props.colStyles.get(j) : {};
        return (
          <td key={`rc-${i}-${j}`} style={style}>
            {elem}
          </td>
        );
      });
      return <tr key={`r-${i}`}>{tds}</tr>;
    }
  );
  const footCols = props.foot.map(
    (elem: React.ReactNode, i: number): React.ReactNode => {
      const style = props.colStyles != null ? props.colStyles.get(i) : {};
      return (
        <th key={`f-${i}`} style={style}>
          {elem}
        </th>
      );
    }
  );

  return (
    <table id={props.id} style={props.style} className={props.className}>
      <thead>
        <tr>{headCols}</tr>
      </thead>
      <tbody>{bodyRows}</tbody>
      <tfoot>
        <tr>{footCols}</tr>
      </tfoot>
    </table>
  );
};
