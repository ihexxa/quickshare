import React from "react";
import ReactDOM from "react-dom";

import { config } from "../../config";
import { Button } from "../control/button";
import { getIcon } from "../display/icon";
import { FileUploader } from "../../libs/api_upload";

const msgFileNotFound = "File not found";
const msgFileUploadOk = "is uploaded";
const msgChromeLink = "https://www.google.com/chrome/";
const msgFirefoxLink = "https://www.mozilla.org/";

export const checkQueueCycle = 1000;

const IconPlus = getIcon("cirUp");
const IconThiList = getIcon("thList");

const styleContainer = {
  position: "fixed",
  bottom: "0.5rem",
  margin: "auto",
  zIndex: 1
};

const styleButtonContainer = {
  backgroundColor: "#2ecc71",
  width: "20rem",
  height: "auto",
  textAlign: "center"
};

const styleDefault = {
  color: "#fff"
};

const styleLabel = {
  display: "inline-block",
  verticalAlign: "middle",
  marginLeft: "0.5rem"
};

const styleUploadQueue = {
  backgroundColor: "#000",
  opacity: 0.85,
  color: "#fff",
  fontSize: "0.75rem",
  lineHeight: "1.25rem"
};

const styleUploadItem = {
  width: "18rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
  padding: "0.5rem 1rem"
};

const styleUnsupported = {
  backgroundColor: "#e74c3c",
  color: "#fff",
  overflow: "hidden",
  padding: "0.5rem 1rem",
  width: "18rem",
  textAlign: "center"
};

const styleStr = `
  a {
    color: white;
    margin: auto 0.5rem auto 0.5rem;
  }
`;

export class Uploader extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      uploadQueue: [],
      uploadValue: ""
    };

    this.input = undefined;
    this.assignInput = input => {
      this.input = ReactDOM.findDOMNode(input);
    };
  }

  componentDidMount() {
    // will polling uploadQueue like a worker
    this.checkQueue();
  }

  checkQueue = () => {
    // TODO: using web worker to avoid lagging UI
    if (this.state.uploadQueue.length > 0) {
      this.upload(this.state.uploadQueue[0]).then(() => {
        this.setState({ uploadQueue: this.state.uploadQueue.slice(1) });
        setTimeout(this.checkQueue, checkQueueCycle);
      });
    } else {
      setTimeout(this.checkQueue, checkQueueCycle);
    }
  };

  upload = file => {
    const fileUploader = new FileUploader(
      this.onStart,
      this.onProgress,
      this.onFinish,
      this.onError
    );

    return fileUploader.uploadFile(file);
  };

  onStart = () => {
    this.props.onRefresh();
  };

  onProgress = (shareId, progress) => {
    this.props.onUpdateProgress(shareId, progress);
  };

  onFinish = () => {
    this.props.onRefresh();
  };

  onError = err => {
    this.props.onError(err);
  };

  onUpload = event => {
    if (event.target.files == null || event.target.files.length === 0) {
      this.props.onError(msgFileNotFound);
      this.setState({ uploadValue: "" });
    } else {
      this.setState({
        uploadQueue: [...this.state.uploadQueue, ...event.target.files],
        uploadValue: ""
      });
    }
  };

  onChooseFile = () => {
    this.input.click();
  };

  render() {
    if (
      window.FormData == null ||
      window.FileReader == null ||
      window.Blob == null
    ) {
      return (
        <div style={{ ...styleUnsupported, ...styleContainer }}>
          Unsupported Browser. Try
          <a href={msgFirefoxLink} target="_blank">
            Firefox
          </a>
          or
          <a href={msgChromeLink} target="_blank">
            Chrome
          </a>
          <style>{styleStr}</style>
        </div>
      );
    }

    const hiddenInput = (
      <input
        type="file"
        onChange={this.onUpload}
        style={{ display: "none" }}
        ref={this.assignInput}
        multiple={true}
        value={this.state.uploadValue}
      />
    );

    const uploadQueue = this.state.uploadQueue.map(file => {
      return (
        <div key={file.name} style={styleUploadItem}>
          <IconThiList
            size={config.rootSize * 0.75}
            style={{ marginRight: "0.5rem" }}
          />
          {file.name}
        </div>
      );
    });

    return (
      <div className="uploader" style={styleContainer}>
        <div style={styleUploadQueue}>{uploadQueue}</div>
        <Button
          onClick={this.onChooseFile}
          label="UPLOAD"
          icon={<IconPlus size={config.rootSize} style={styleDefault} />}
          styleDefault={styleDefault}
          styleContainer={styleButtonContainer}
          styleLabel={styleLabel}
        />
        {hiddenInput}
      </div>
    );
  }
}

Uploader.defaultProps = {
  onRefresh: () => console.error("undefined"),
  onUpdateProgress: () => console.error("undefined"),
  onOk: () => console.error("undefined"),
  onError: () => console.error("undefined")
};
