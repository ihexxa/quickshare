import React from "react";
import { FileBoxDetail } from "./file_box_detail";
import { Button } from "../control/button";

import { config } from "../../config";
import { getIcon, getIconColor } from "../display/icon";
import { getFileExt } from "../../libs/file_type";
import { del, publishId, shadowId, setDownLimit } from "../../libs/api_share";

const msgUploadOk = "Uploading is stopped and file is deleted";
const msgUploadNok = "Fail to delete file";

const styleLeft = {
  float: "left",
  padding: "1rem 0 1rem 1rem"
};

const styleRight = {
  float: "right",
  textAlign: "right",
  padding: "0rem"
};

const clear = <div style={{ clear: "both" }} />;

const iconDesStyle = {
  display: "inline-block",
  fontSize: "0.875rem",
  lineHeight: "1rem",
  marginBottom: "0.25rem",
  maxWidth: "12rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  textDecoration: "none",
  verticalAlign: "middle",
  whiteSpace: "nowrap"
};

const descStyle = {
  fontSize: "0.75rem",
  padding: "0.75rem"
};

const otherStyle = `
.main-pane {
  background-color: rgba(255, 255, 255, 1);
  transition: background-color 0.1s;
}
.main-pane:hover {
  background-color: rgba(255, 255, 255, 0.85);
  transition: background-color 0.1s;
}

.show-detail {
  opacity: 1;
  height: auto;
  transition: opacity 0.15s, height 0.5s;
}

.hide-detail {
  opacity: 0;
  height: 0;
  overflow: hidden;
  transition: opacity 0.15s, height 0.5s;
}

.main-pane a {
    color: #333;
    transition: color 1s;
}
.main-pane a:hover {
    color: #3498db;
    transition: color 1s;
}
`;

const IconMore = getIcon("bars");
const IconTimesCir = getIcon("timesCir");
const styleIconTimesCir = {
  color: getIconColor("timesCir")
};

const iconMoreStyleStr = `
.file-box-more {
  color: #333;
  background-color: #fff;
  transition: color 0.4s, background-color 0.4s;
}

.file-box-more:hover {
    color: #000;
    background-color: #ccc;
    transition: color 0.4s, background-color 0.4s;
}
`;

let styleFileBox = {
  textAlign: "left",
  margin: "1px 0px",
  fontSize: "0.75rem"
};

const styleButtonContainer = {
  width: "1rem",
  height: "1rem",
  padding: "1.5rem 1rem"
};

const styleButtonIcon = {
  lineHeight: "1rem",
  height: "1rem",
  margin: "0"
};

export class FileBox extends React.PureComponent {
  constructor(props) {
    super(props);
  }

  onToggleDetail = () => {
    this.props.onToggleDetail(this.props.id);
  };

  onDelete = () => {
    del(this.props.id).then(ok => {
      if (ok) {
        this.props.onOk(msgUploadOk);
        this.props.onRefresh();
      } else {
        this.props.onError(msgUploadNok);
      }
    });
  };

  render() {
    const ext = getFileExt(this.props.name);
    const IconFile = getIcon(ext);
    const IconSpinner = getIcon("spinner");

    const styleIcon = {
      color: this.props.isLoading ? "#34495e" : getIconColor(ext)
    };

    styleFileBox = {
      ...styleFileBox,
      width: this.props.width
    };

    const fileIcon = this.props.isLoading ? (
      <IconSpinner
        size={config.rootSize * 2}
        style={styleIcon}
        className="anm-rotate"
      />
    ) : (
      <IconFile size={config.rootSize * 2} style={styleIcon} />
    );

    const opIcon = this.props.isLoading ? (
      <Button
        icon={<IconTimesCir size={config.rootSize} style={styleIconTimesCir} />}
        label=""
        styleContainer={styleButtonContainer}
        styleIcon={styleButtonIcon}
        onClick={this.onDelete}
      />
    ) : (
      <Button
        icon={<IconMore size={config.rootSize} />}
        className={"file-box-more"}
        label=""
        styleContainer={styleButtonContainer}
        styleIcon={styleButtonIcon}
        styleStr={iconMoreStyleStr}
        onClick={this.onToggleDetail}
      />
    );

    const downloadLink = (
      <a href={this.props.href} style={iconDesStyle} target="_blank">
        {this.props.name}
      </a>
    );

    const classDetailPane =
      this.props.showDetailId === this.props.id &&
      this.props.uploadState === "done"
        ? "show-detail"
        : "hide-detail";

    return (
      <div key={this.props.id} style={styleFileBox} className="file-box">
        <div className="main-pane">
          <div style={styleLeft}>{fileIcon}</div>
          <div style={styleLeft}>
            {downloadLink}
            <div
              style={{
                color: "#999",
                lineHeight: "0.75rem",
                height: "0.75rem"
              }}
            >{`${this.props.size} ${this.props.modTime}`}</div>
          </div>
          <div style={styleRight}>{opIcon}</div>
          {clear}
          <style>{otherStyle}</style>
        </div>

        <div style={{ position: "relative" }}>
          <FileBoxDetail
            id={this.props.id}
            name={this.props.name}
            size={this.props.size}
            modTime={this.props.modTime}
            href={this.props.href}
            downLimit={this.props.downLimit}
            width={this.props.width}
            className={classDetailPane}
            onRefresh={this.props.onRefresh}
            onError={this.props.onError}
            onOk={this.props.onOk}
            onDel={del}
            onPublishId={publishId}
            onShadowId={shadowId}
            onSetDownLimit={setDownLimit}
          />
        </div>
      </div>
    );
  }
}

FileBox.defaultProps = {
  id: "",
  name: "",
  isLoading: false,
  modTime: "unknown",
  uploadState: "",
  href: "",
  width: "320px",
  showDetailId: "",
  downLimit: -3,
  size: "unknown",
  onToggleDetail: () => console.error("undefined"),
  onRefresh: () => console.error("undefined"),
  onError: () => console.error("undefined"),
  onOk: () => console.error("undefined")
};
