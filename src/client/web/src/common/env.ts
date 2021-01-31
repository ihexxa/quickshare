export function alertMsg(msg: string) {
  if (alert != null) {
    alert(msg);
  } else {
    console.log(msg);
  }
}

export function comfirmMsg(msg: string): boolean {
  return confirm(msg);
}
