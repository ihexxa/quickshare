export function alertMsg(msg: string) {
  if (alert != null) {
    alert(msg);
  } else {
    console.log(msg);
  }
}

export function confirmMsg(msg: string): boolean {
  if (confirm != null) {
    return confirm(msg);
  } else {
    console.warn(`${msg}: return yes (confirm is not implemented)`);
    return true;
  }
}
