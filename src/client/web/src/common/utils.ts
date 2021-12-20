import { Map } from "immutable";

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
