import * as React from "react";
import * as ReactDOM from "react-dom";

import { StateMgr } from "./components/state_mgr";

import "./theme/reset.css";
import "./theme/white.css";
// TODO: it fails in jest preprocessor now
import "./theme/style.css";
import "./theme/desktop.css";
import "./theme/color.css";


ReactDOM.render(<StateMgr />, document.getElementById("mount"));
