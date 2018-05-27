import React from "react";
import { getIcon, getIconColor } from "../display/icon";

const statusNull = "null";
const statusInfo = "info";
const statusWarn = "warn";
const statusError = "error";
const statusOk = "ok";
const statusStart = "start";
const statusEnd = "end";

const IconInfo = getIcon("infoCir");
const IconWarn = getIcon("exTri");
const IconError = getIcon("timesCir");
const IconOk = getIcon("checkCir");
const IconStart = getIcon("refresh");

const colorInfo = getIconColor("infoCir");
const colorWarn = getIconColor("exTri");
const colorError = getIconColor("timesCir");
const colorOk = getIconColor("checkCir");
const colorStart = getIconColor("refresh");

const classFadeIn = "log-fade-in";
const classHidden = "log-hidden";
const styleStr = `
  .log .${classFadeIn} {
    opacity: 1;
    margin-left: 0.5rem;
    padding: 0.25rem 0.5rem;
    transition: opacity 0.3s, margin-left 0.3s, padding 0.3s;
  }

  .log .${classHidden} {
    opacity: 0;
    margin-left: 0rem;
    padding: 0;
    transition: opacity 0.3s, margin-left 0.3s, padding 0.3s;
  }

  .log a {
    color: #2980b9;
    transition: color 0.3s;
    text-decoration: none;
  }

  .log a:hover {
    color: #3498db;
    transition: color 0.3s;
    text-decoration: none;
  }
`;

const wait = 5000;
const logSlotLen = 2;
const getEmptyLog = () => ({
  className: classHidden,
  msg: "",
  status: statusNull
});

const getLogIcon = status => {
  switch (status) {
    case statusInfo:
      return (
        <IconInfo
          size={16}
          style={{ marginRight: "0.25rem", color: colorInfo }}
        />
      );
    case statusWarn:
      return (
        <IconWarn
          size={16}
          style={{ marginRight: "0.25rem", color: colorWarn }}
        />
      );
    case statusError:
      return (
        <IconError
          size={16}
          style={{ marginRight: "0.25rem", color: colorError }}
        />
      );
    case statusOk:
      return (
        <IconOk size={16} style={{ marginRight: "0.25rem", color: colorOk }} />
      );
    case statusStart:
      return (
        <IconStart
          size={16}
          className={"anm-rotate"}
          style={{ marginRight: "0.25rem", color: colorStart }}
        />
      );
    case statusEnd:
      return (
        <IconOk size={16} style={{ marginRight: "0.25rem", color: colorOk }} />
      );
    default:
      return <span />;
  }
};

export class Log extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      logs: Array(logSlotLen).fill(getEmptyLog())
    };
    this.id = 0;
  }

  genId = () => {
    return this.id++ % logSlotLen;
  };

  addLog = (status, msg) => {
    const id = this.genId();
    const nextLogs = [
      ...this.state.logs.slice(0, id),
      {
        className: classFadeIn,
        msg,
        status
      },
      ...this.state.logs.slice(id + 1)
    ];

    this.setState({ logs: nextLogs });
    this.delayClearLog(id);
    return id;
  };

  delayClearLog = idToDel => {
    setTimeout(this.clearLog, wait, idToDel);
  };

  clearLog = idToDel => {
    // TODO: there may be race condition here
    const nextLogs = [
      ...this.state.logs.slice(0, idToDel),
      getEmptyLog(),
      ...this.state.logs.slice(idToDel + 1)
    ];
    this.setState({ logs: nextLogs });
  };

  info = msg => {
    this.addLog(statusInfo, msg);
  };

  warn = msg => {
    this.addLog(statusWarn, msg);
  };

  error = msg => {
    this.addLog(statusError, msg);
  };

  ok = msg => {
    this.addLog(statusOk, msg);
  };

  start = msg => {
    const id = this.genId();
    const nextLogs = [
      ...this.state.logs.slice(0, id),
      {
        className: classFadeIn,
        msg,
        status: statusStart
      },
      ...this.state.logs.slice(id + 1)
    ];

    this.setState({ logs: nextLogs });
    return id;
  };

  end = (startId, msg) => {
    // remove start log
    this.clearLog(startId);
    this.addLog(statusEnd, msg);
  };

  render() {
    const logList = Object.keys(this.state.logs).map(logId => {
      return (
        <span
          key={logId}
          style={this.props.styleLog}
          className={this.state.logs[logId].className}
        >
          {getLogIcon(this.state.logs[logId].status)}
          {this.state.logs[logId].msg}
        </span>
      );
    });

    return (
      <span className={"log"} style={this.props.style}>
        {logList}
        <style>{styleStr}</style>
      </span>
    );
  }
}

Log.defaultProps = {
  style: {},
  styleLog: {}
};
