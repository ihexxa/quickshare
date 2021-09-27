export function alertMsg(msg: string) {
  if (alert != null) {
    alert(msg);
  } else {
    console.log(msg);
  }
}

export function confirmMsg(msg: string): boolean {
  return confirm(msg);
}
