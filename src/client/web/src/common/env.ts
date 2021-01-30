export function alertMsg(msg: string) {
  if (alert != null) {
    alert(msg);
  } else {
    console.log(msg);
  }
}
