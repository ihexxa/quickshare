import * as React from "react";
import * as ReactDOM from "react-dom";

import { StateMgr } from "./components/state_mgr";
import { ErrorLogger } from "./common/log_error";
import { errCorsScript } from "./common/errors";

window.onerror = (
  msg: string,
  source: string,
  lineno: number,
  colno: number,
  error: Error
) => {
  const lowerMsg = msg.toLowerCase();
  if (lowerMsg.indexOf("script error") > -1) {
    ErrorLogger().error(errCorsScript, "Check Browser Console for Detail");
  }
  ErrorLogger().error(`${source}:${lineno}:${colno}: ${error.toString()}`, "");
};

ReactDOM.render(<StateMgr />, document.getElementById("mount"));
