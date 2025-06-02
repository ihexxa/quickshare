import * as React from "react";
import { Map, List } from "immutable";
import { throttle } from "throttle-debounce";

import { ICoreState, MsgProps, UIProps } from "./core_state";
import { FilesPanel, FilesProps } from "./panel_files";
import { UploadingsPanel, UploadingsProps } from "./panel_uploadings";
import { SharingsPanel, SharingsProps } from "./panel_sharings";
import { IconProps, iconSize } from "./visual/icons";
import { Tabs } from "./control/tabs";
import { LoginProps } from "./pane_login";
import { Layers } from "./layers";
import { AdminProps } from "./pane_admin";
import { TopBar } from "./topbar";
import { CronJobs } from "../common/cron";
import { updater } from "./state_updater";
import { dropAreaCtrl, ctrlOn, ctrlOff } from "../common/controls";
import { Container } from "./layout/container";

export const controlName = "panelTabs";
const dragOverthrottlePeriod = 200;
export interface Props {
  filesInfo: FilesProps;
  uploadingsInfo: UploadingsProps;
  sharingsInfo: SharingsProps;
  admin: AdminProps;
  login: LoginProps;
  msg: MsgProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {
  lastDragOverTime: number;
}
export class RootFrame extends React.Component<Props, State, {}> {
  private filesPanelRef: FilesPanel;
  constructor(p: Props) {
    super(p);
    this.state = {
      lastDragOverTime: 0,
    };
  }

  componentDidMount(): void {
    CronJobs().setInterval("autoSwitchTheme", {
      func: updater().autoSwitchTheme,
      args: [],
      delay: 60 * 1000,
    });

    CronJobs().setInterval("endDrag", {
      func: this.endDrag,
      args: [],
      delay: dragOverthrottlePeriod * 2,
    });
  }

  componentWillUnmount() {
    CronJobs().clearInterval("autoSwitchTheme");
  }

  private setFilesPanelRef = (ref: FilesPanel) => {
    this.filesPanelRef = ref;
  };

  makeBgStyle = (): Object => {
    if (this.props.ui.clientCfg.allowSetBg) {
      if (
        this.props.login.preferences != null &&
        this.props.login.preferences.bg.url !== ""
      ) {
        const bgConfig = this.props.login.preferences.bg;
        return {
          background: `url("${bgConfig.url}") ${bgConfig.repeat} ${bgConfig.position} ${bgConfig.align}`,
        };
      }

      if (this.props.login.preferences.bg.bgColor !== "") {
        return {
          backgroundColor: this.props.login.preferences.bg.bgColor,
        };
      }
    }

    if (this.props.ui.clientCfg.bg.url !== "") {
      return {
        background: `url("${this.props.ui.clientCfg.bg.url}") ${this.props.ui.clientCfg.bg.repeat} ${this.props.ui.clientCfg.bg.position} ${this.props.ui.clientCfg.bg.align}`,
      };
    }

    if (this.props.ui.clientCfg.bg.bgColor !== "") {
      return {
        backgroundColor: this.props.ui.clientCfg.bg.bgColor,
      };
    }

    return {};
  };

  onDragOver = (ev: React.DragEvent<HTMLDivElement>) => {
    this.onDragOverImp();
    ev.preventDefault();
  };

  onDragOverImp = throttle(dragOverthrottlePeriod, () => {
    updater().setControlOption(dropAreaCtrl, ctrlOn);
    this.props.update(updater().updateUI);
    this.setState({ lastDragOverTime: Date.now() });
  });

  onDrop = (ev: React.DragEvent<HTMLDivElement>) => {
    if (ev.dataTransfer?.files?.length > 0) {
      this.filesPanelRef.addFileList(ev.dataTransfer.files);
    }
    ev.preventDefault();
  };

  endDrag = () => {
    const now = Date.now();
    const isDragOverOff =
      this.props.ui.control.controls.get(dropAreaCtrl) === ctrlOff;
    if (
      now - this.state.lastDragOverTime < dragOverthrottlePeriod * 1.5 ||
      isDragOverOff
    ) {
      return;
    }

    updater().setControlOption(dropAreaCtrl, ctrlOff);
    this.props.update(updater().updateUI);
  };

  render() {
    const bgStyle = this.makeBgStyle();
    const autoTheme =
      updater().getCurrentTheme() === "light" ? "theme-default" : "theme-dark";
    const fixedTheme =
      this.props.login.preferences.theme === "light"
        ? "theme-default"
        : "theme-dark";
    const theme = this.props.ui.clientCfg.autoTheme ? autoTheme : fixedTheme;

    const fontSizeClass = "font-xs";

    const displaying = this.props.ui.control.controls.get(controlName);
    const filesPanelClass = displaying === "filesPanel" ? "" : "hidden";
    const uploadingsPanelClass =
      displaying === "uploadingsPanel" ? "" : "hidden";
    const sharingsPanelClass = displaying === "sharingsPanel" ? "" : "hidden";

    return (
      <div
        id="root-frame"
        className={`${theme} ${fontSizeClass}`}
        onDragOver={this.onDragOver}
        onDrop={this.onDrop}
      >
        <div id="bg" style={bgStyle}>
          <div id="custom">
            <Layers
              login={this.props.login}
              admin={this.props.admin}
              ui={this.props.ui}
              msg={this.props.msg}
              filesInfo={this.props.filesInfo}
              update={this.props.update}
            />

            <div className="container-center pt-8">
              <Container>
                <TopBar
                  login={this.props.login}
                  msg={this.props.msg}
                  ui={this.props.ui}
                  update={this.props.update}
                />
                <div className="hr"></div>
                <Tabs
                  targetControl={controlName}
                  tabIcons={Map<string, IconProps>({
                    filesPanel: {
                      name: "RiFolder2Fill",
                      size: iconSize("s"),
                      color: "focus",
                    },
                    uploadingsPanel: {
                      name: "RiUploadCloudFill",
                      size: iconSize("s"),
                      color: "focus",
                    },
                    sharingsPanel: {
                      name: "FaSquareShareNodes",
                      size: iconSize("s"),
                      color: "focus",
                    },
                  })}
                  ui={this.props.ui}
                  msg={this.props.msg}
                  update={this.props.update}
                />
              </Container>

              <span className={filesPanelClass}>
                <FilesPanel
                  filesInfo={this.props.filesInfo}
                  msg={this.props.msg}
                  login={this.props.login}
                  ui={this.props.ui}
                  enabled={displaying === "filesPanel"}
                  update={this.props.update}
                  ref={this.setFilesPanelRef}
                />
              </span>

              <span className={uploadingsPanelClass}>
                <UploadingsPanel
                  uploadingsInfo={this.props.uploadingsInfo}
                  msg={this.props.msg}
                  login={this.props.login}
                  ui={this.props.ui}
                  update={this.props.update}
                />
              </span>

              <span className={sharingsPanelClass}>
                <SharingsPanel
                  sharingsInfo={this.props.sharingsInfo}
                  msg={this.props.msg}
                  login={this.props.login}
                  ui={this.props.ui}
                  update={this.props.update}
                />
              </span>
            </div>

            <div id="footer" className="container-center">
              <a href="https://github.com/ihexxa/quickshare">Quickshare</a> -
              Quick and simple file sharing.
            </div>
          </div>
        </div>
      </div>
    );
  }
}
