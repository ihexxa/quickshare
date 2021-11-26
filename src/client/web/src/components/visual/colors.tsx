import { Set, Map } from "immutable";

export const colors = Set<string>([
  "blue0",
  "blue1",
  "cyan0",
  "cyan1",
  "purple0",
  "purple1",
  "red0",
  "red1",
  "yellow0",
  "yellow1",
  "yellow2",
  "yellow3",
  "green0",
  "green1",
  "green2",
  "white",
  "white0",
  "white1",
  "grey0",
  "grey1",
  "grey2",
  "grey3",
  "black",
  "black0",
  "black1",
]);

export function colorClass(name: string): string {
  if (!colors.has(name)) {
    console.error(`color ${name} not found`);
    return colors.get("black");
  }
  return colors.get(name);
}
