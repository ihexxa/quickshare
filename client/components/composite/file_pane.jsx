import axios from "axios";
import byteSize from "byte-size";
import React from "react";
import ReactDOM from "react-dom";
import throttle from "lodash.throttle";
import { Grids } from "../../components/layout/grids";
import { Uploader } from "../../components/composite/uploader";
import { FileBox } from "./file_box";
import { TimeGrids } from "./time_grids";

import { config } from "../../config";

const msgSynced = "Synced";
const msgSyncFailed = "Syncing failed";
const interval = 250;

export class FilePane extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      infos: [],
      showDetailId: -1
    };
    this.onRefresh = throttle(this.onRefreshImp, interval);
    this.onUpdateProgress = throttle(this.onUpdateProgressImp, interval);
  }

  componentWillMount() {
    return this.onRefreshImp();
  }

  onRefreshImp = () => {
    return this.props
      .onList()
      .then(infos => {
        if (infos != null) {
          this.setState({ infos });
          this.props.onOk(msgSynced);
        } else {
          this.props.onError(msgSyncFailed);
        }
      })
      .catch(err => {
        console.error(err);
        this.props.onError(msgSyncFailed);
      });
  };

  onUpdateProgressImp = (shareId, progress) => {
    const updatedInfos = this.state.infos.map(shareInfo => {
      return shareInfo.Id === shareId ? { ...shareInfo, progress } : shareInfo;
    });

    this.setState({ infos: updatedInfos });
  };

  onToggleDetail = id => {
    this.setState({
      showDetailId: this.state.showDetailId === id ? -1 : id
    });
  };

  getByteSize = size => {
    const sizeObj = byteSize(size);
    return `${sizeObj.value} ${sizeObj.unit}`;
  };

  getInfos = filterName => {
    const filteredInfos = this.state.infos.filter(shareInfo => {
      return shareInfo.PathLocal.includes(filterName);
    });

    return filteredInfos.map(shareInfo => {
      const isLoading = shareInfo.State === "uploading";
      const timestamp = shareInfo.ModTime / 1000000;
      const modTime = new Date(timestamp).toLocaleString();
      const href = `${config.serverAddr}/download?shareid=${shareInfo.Id}`;
      const progress = isNaN(shareInfo.progress) ? 0 : shareInfo.progress;
      const name = isLoading
        ? `${Math.floor(progress * 100)}% ${shareInfo.PathLocal}`
        : shareInfo.PathLocal;

      return {
        key: shareInfo.Id,
        timestamp,
        component: (
          <FileBox
            key={shareInfo.Id}
            id={shareInfo.Id}
            name={name}
            size={this.getByteSize(shareInfo.Uploaded)}
            uploadState={shareInfo.State}
            isLoading={isLoading}
            modTime={modTime}
            href={href}
            downLimit={shareInfo.DownLimit}
            width={`${this.props.colWidth}rem`}
            onRefresh={this.onRefresh}
            onOk={this.props.onOk}
            onError={this.props.onError}
            showDetailId={this.state.showDetailId}
            onToggleDetail={this.onToggleDetail}
          />
        )
      };
    });
  };

  render() {
    const styleUploaderContainer = {
      width: `${this.props.colWidth}rem`,
      margin: "auto"
    };

    const containerStyle = {
      width: this.props.width,
      margin: "auto",
      marginTop: "0",
      marginBottom: "10rem"
    };

    return (
      <div className="file-pane" style={containerStyle}>
        <TimeGrids
          items={this.getInfos(this.props.filterName)}
          styleContainer={{
            width:
              this.props.width === "auto"
                ? config.rootSize * config.colWidth
                : this.props.width,
            margin: "auto"
          }}
        />
        <div style={styleUploaderContainer}>
          <Uploader
            onRefresh={this.onRefresh}
            onUpdateProgress={this.onUpdateProgress}
            onOk={this.props.onOk}
            onError={this.props.onError}
          />
        </div>
      </div>
    );
  }
}

FilePane.defaultProps = {
  width: "100%",
  colWidth: 20,
  filterName: "",
  onList: () => console.error("undefined"),
  onOk: () => console.error("undefined"),
  onError: () => console.error("undefined")
};
