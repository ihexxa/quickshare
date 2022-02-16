import { List, Map } from "immutable";

import { Row } from "../components/layout/rows";

export function getItemPath(dirPath: string, itemName: string): string {
  return dirPath.endsWith("/")
    ? `${dirPath}${itemName}`
    : `${dirPath}/${itemName}`;
}

export function getErrMsg(
  msgPkg: Map<string, string>,
  msg: string,
  status: string
): string {
  return `${msgPkg.get(msg)}: ${msgPkg.get(status)}`;
}

export function sortRows(
  rows: List<Row>,
  key: number,
  order: boolean
): List<Row> {
  return rows.sort((row1: Row, row2: Row) => {
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
      return order ? -1 : 1;
    } else if (val1 === val2) {
      return 0;
    }
    return order ? 1 : -1;
  });
}
