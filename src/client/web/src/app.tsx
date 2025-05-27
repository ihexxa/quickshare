import * as React from "react";
import * as ReactDOM from "react-dom/client";

import { StateMgr } from "./components/state_mgr";
import { ErrorLogger } from "./common/log_error";
import { errCorsScript } from "./common/errors";

import './style/tailwind.css';

window.onerror = (
  msg: string,
  source: string,
  lineno: number,
  colno: number,
  error: Error
) => {
  const lowerMsg = msg.toLowerCase();
  if (lowerMsg.indexOf("script error") > -1) {
    ErrorLogger().error("Check Browser Console for Detail");
  }
  ErrorLogger().error(`${source}:${lineno}:${colno}: ${error.toString()}`);
};

const root = ReactDOM.createRoot(document.getElementById("mount"));
root.render(<StateMgr />);
