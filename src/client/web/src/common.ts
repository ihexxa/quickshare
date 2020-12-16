export const range = (start: number, end: number): Array<number> => {
  let array = new Array(0);
  for (let i = start; i <= end; i++) {
    array.push(i);
  }
  return array;
};
