export function alertMsg(msg: string) {
  if (alert != null) {
    alert(msg);
  } else {
    console.log(msg);
  }
}

export function confirmMsg(msg: string): boolean {
  try {
    return confirm(msg);
  } catch (e) {
    console.log(`${msg}: yes (confirm is not implemented)`);
    return true;
  }
}
