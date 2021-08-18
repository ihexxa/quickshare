import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map } from "immutable";
import FileSize from "filesize";

import { Layouter } from "./layouter";
import { alertMsg, comfirmMsg } from "../common/env";
import { updater } from "./browser.updater";
import { ICoreState } from "./core_state";
import {
  IUsersClient,
  IFilesClient,
  MetadataResp,
  UploadInfo,
} from "../client";
import { Up } from "../worker/upload_mgr";
import { UploadEntry } from "../worker/interface";

export interface Item {
  name: string;
  size: number;
  modTime: string;
  isDir: boolean;
  selected: boolean;
}

export interface Props {
  dirPath: List<string>;
  isSharing: boolean;
  items: List<MetadataResp>;
  uploadings: List<UploadInfo>;
  sharings: List<string>;

  uploadFiles: List<File>;
  uploadValue: string;

  isVertical: boolean;

  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export function getItemPath(dirPath: string, itemName: string): string {
  return dirPath.endsWith("/")
    ? `${dirPath}${itemName}`
    : `${dirPath}/${itemName}`;
}

export interface State {
  inputValue: string;
  selectedSrc: string;
  selectedItems: Map<string, boolean>;
}

export class Browser extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  private uploadInput: Element | Text;
  private assignInput: (input: Element) => void;
  private onClickUpload: () => void;

  constructor(p: Props) {
    super(p);
    this.update = p.update;
    this.state = {
      inputValue: "",
      selectedSrc: "",
      selectedItems: Map<string, boolean>(),
    };

    Up().setStatusCb(this.updateProgress);
    this.uploadInput = undefined;
    this.assignInput = (input) => {
      this.uploadInput = ReactDOM.findDOMNode(input);
    };
    this.onClickUpload = () => {
      // TODO: check if the re-upload file is same as previous upload
      const uploadInput = this.uploadInput as HTMLButtonElement;
      uploadInput.click();
    };
  }

  onInputChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ inputValue: ev.target.value });
  };

  addUploads = (event: React.ChangeEvent<HTMLInputElement>) => {
    let fileList = List<File>();
    for (let i = 0; i < event.target.files.length; i++) {
      fileList = fileList.push(event.target.files[i]);
    }
    updater().addUploads(fileList);
    this.update(updater().setBrowser);
  };

  deleteUpload = (filePath: string): Promise<void> => {
    return updater()
      .deleteUpload(filePath)
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(`Failed to delete uploading ${filePath}`);
        }
        return updater().refreshUploadings();
      })
      .then(() => {
        this.update(updater().setBrowser);
      });
  };

  stopUploading = (filePath: string) => {
    updater().stopUploading(filePath);
    this.update(updater().setBrowser);
  };

  onMkDir = () => {
    if (this.state.inputValue === "") {
      alertMsg("folder name can not be empty");
      return;
    }

    const dirPath = getItemPath(
      this.props.dirPath.join("/"),
      this.state.inputValue
    );
    updater()
      .mkDir(dirPath)
      .then(() => {
        this.setState({ inputValue: "" });
        return updater().setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(updater().setBrowser);
      });
  };

  delete = () => {
    if (this.props.dirPath.join("/") !== this.state.selectedSrc) {
      alertMsg("please select file or folder to delete at first");
      this.setState({
        selectedSrc: this.props.dirPath.join("/"),
        selectedItems: Map<string, boolean>(),
      });
      return;
    } else {
      const filesToDel = this.state.selectedItems.keySeq().join(", ");
      if (!comfirmMsg(`do you want to delete ${filesToDel}?`)) {
        return;
      }
    }

    updater()
      .delete(this.props.dirPath, this.props.items, this.state.selectedItems)
      .then(() => {
        this.update(updater().setBrowser);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  moveHere = () => {
    const oldDir = this.state.selectedSrc;
    const newDir = this.props.dirPath.join("/");
    if (oldDir === newDir) {
      alertMsg("source directory is same as destination directory");
      return;
    }

    updater()
      .moveHere(
        this.state.selectedSrc,
        this.props.dirPath.join("/"),
        this.state.selectedItems
      )
      .then(() => {
        this.update(updater().setBrowser);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  gotoChild = (childDirName: string) => {
    this.chdir(this.props.dirPath.push(childDirName));
  };

  chdir = (dirPath: List<string>) => {
    if (dirPath === this.props.dirPath) {
      return;
    }

    updater()
      .setItems(dirPath)
      .then(() => {
        updater().listSharings();
      })
      .then(() => {
        updater().setSharing(dirPath.join("/"));
      })
      .then(() => {
        this.update(updater().setBrowser);
      });
  };

  updateProgress = (infos: Map<string, UploadEntry>) => {
    updater().setUploadings(infos);
    updater()
      .setItems(this.props.dirPath)
      .then(() => {
        this.update(updater().setBrowser);
      });
  };

  select = (itemName: string) => {
    const selectedItems = this.state.selectedItems.has(itemName)
      ? this.state.selectedItems.delete(itemName)
      : this.state.selectedItems.set(itemName, true);

    this.setState({
      selectedSrc: this.props.dirPath.join("/"),
      selectedItems: selectedItems,
    });
  };

  selectAll = () => {
    let newSelected = Map<string, boolean>();
    const someSelected = this.state.selectedItems.size === 0 ? true : false;
    if (someSelected) {
      this.props.items.forEach((item) => {
        newSelected = newSelected.set(item.name, true);
      });
    } else {
      this.props.items.forEach((item) => {
        newSelected = newSelected.delete(item.name);
      });
    }

    this.setState({
      selectedSrc: this.props.dirPath.join("/"),
      selectedItems: newSelected,
    });
  };

  addSharing = () => {
    updater()
      .addSharing()
      .then((ok) => {
        if (!ok) {
          alert("failed to enable sharing");
        } else {
          this.listSharings();
        }
      });
  };

  deleteSharing = (dirPath: string) => {
    updater()
      .deleteSharing(dirPath)
      .then((ok) => {
        if (!ok) {
          alert("failed to disable sharing");
        } else {
          this.listSharings();
        }
      });
  };

  listSharings = () => {
    updater()
      .listSharings()
      .then((ok) => {
        if (ok) {
          this.update(updater().setBrowser);
        }
      });
  };

  render() {
    const breadcrumb = this.props.dirPath.map(
      (pathPart: string, key: number) => {
        return (
          <span key={pathPart}>
            <button
              type="button"
              onClick={() => this.chdir(this.props.dirPath.slice(0, key + 1))}
              className="white-font margin-r-m"
              style={{ backgroundColor: "rgba(0, 0, 0, 0.7)" }}
            >
              {pathPart}
            </button>
          </span>
        );
      }
    );

    const nameCellClass = `item-name item-name-${
      this.props.isVertical ? "vertical" : "horizontal"
    } pointer`;
    const sizeCellClass = this.props.isVertical ? `hidden margin-s` : ``;
    const modTimeCellClass = this.props.isVertical ? `hidden margin-s` : ``;

    const ops = (
      <div>
        <div>
          <span className="inline-block margin-t-m margin-b-m">
            <input
              type="text"
              onChange={this.onInputChange}
              value={this.state.inputValue}
              className="black0-font margin-r-m"
              placeholder="folder name"
            />
            <button
              onClick={this.onMkDir}
              className="grey1-bg white-font margin-r-m"
            >
              Create Folder
            </button>
          </span>
          <span className="inline-block margin-t-m margin-b-m">
            <button
              onClick={this.onClickUpload}
              className="green0-bg white-font"
            >
              Upload Files
            </button>
            <input
              type="file"
              onChange={this.addUploads}
              multiple={true}
              value={this.props.uploadValue}
              ref={this.assignInput}
              className="black0-font hidden"
            />
          </span>
        </div>

        <div className="hr white0-bg margin-t-m margin-b-m"></div>

        <div>
          <button
            type="button"
            onClick={() => this.delete()}
            className="red0-bg white-font margin-t-m margin-b-m margin-r-m"
          >
            Delete Selected
          </button>
          <button
            type="button"
            onClick={() => this.moveHere()}
            className="grey1-bg white-font margin-t-m margin-b-m margin-r-m"
          >
            Paste
          </button>
          {this.props.isSharing ? (
            <button
              type="button"
              onClick={() => {
                this.deleteSharing(this.props.dirPath.join("/"));
              }}
              className="red0-bg white-font margin-t-m margin-b-m"
            >
              Stop Sharing
            </button>
          ) : (
            <button
              type="button"
              onClick={this.addSharing}
              className="green0-bg white-font margin-t-m margin-b-m"
            >
              Share Folder
            </button>
          )}
        </div>
      </div>
    );

    const itemList = this.props.items.map((item: MetadataResp) => {
      const isSelected = this.state.selectedItems.has(item.name);
      const dirPath = this.props.dirPath.join("/");
      const itemPath = dirPath.endsWith("/")
        ? `${dirPath}${item.name}`
        : `${dirPath}/${item.name}`;

      return item.isDir ? (
        <div key={item.name} className="flex-list-container">
          <span className="flex-list-item-l">
            <span className="vbar yellow2-bg"></span>
            <span className={nameCellClass}>
              <span className="bold" onClick={() => this.gotoChild(item.name)}>
                {item.name}
              </span>
              <div className="grey1-font">
                <span>{item.modTime.slice(0, item.modTime.indexOf("T"))}</span>
              </div>
            </span>
          </span>
          <span className="flex-list-item-r padding-r-m">
            <span className="item-op">
              <button
                onClick={() => this.select(item.name)}
                className={`white-font ${isSelected ? "blue0-bg" : "grey1-bg"}`}
                style={{ width: "8rem", display: "inline-block" }}
              >
                {isSelected ? "Deselect" : "Select"}
              </button>
            </span>
          </span>
        </div>
      ) : (
        <div key={item.name} className="flex-list-container">
          <span className="flex-list-item-l">
            <span className="vbar green2-bg"></span>
            <span className={nameCellClass}>
              <a
                className="bold"
                href={`/v1/fs/files?fp=${itemPath}`}
                target="_blank"
              >
                {item.name}
              </a>
              <div className="grey1-font">
                <span>{FileSize(item.size, { round: 0 })}</span>&nbsp;/&nbsp;
                <span>{item.modTime.slice(0, item.modTime.indexOf("T"))}</span>
              </div>
            </span>
          </span>
          <span className="flex-list-item-r padding-r-m">
            <span className="item-op">
              <button
                type="button"
                onClick={() => this.select(item.name)}
                className={`white-font ${isSelected ? "blue0-bg" : "grey1-bg"}`}
                style={{ width: "8rem", display: "inline-block" }}
              >
                {isSelected ? "Deselect" : "Select"}
              </button>
            </span>
          </span>
        </div>
      );
    });

    const uploadingList = this.props.uploadings.map((uploading: UploadInfo) => {
      const pathParts = uploading.realFilePath.split("/");
      const fileName = pathParts[pathParts.length - 1];

      return (
        <div key={fileName} className="flex-list-container">
          <span className="flex-list-item-l">
            <span className="vbar blue2-bg"></span>
            <div className={nameCellClass}>
              <span className="bold">{fileName}</span>
              <div className="grey1-font">
                {FileSize(uploading.uploaded, { round: 0 })}
                &nbsp;/&nbsp;{FileSize(uploading.size, { round: 0 })}
              </div>
            </div>
          </span>
          <span className="flex-list-item-r padding-r-m">
            <div className="item-op">
              <button
                onClick={() => this.stopUploading(uploading.realFilePath)}
                className="grey1-bg white-font margin-r-m"
              >
                Stop
              </button>
              <button
                onClick={() => this.deleteUpload(uploading.realFilePath)}
                className="grey1-bg white-font"
              >
                Delete
              </button>
            </div>
          </span>
        </div>
      );
    });

    const sharingList = this.props.sharings.map((dirPath: string) => {
      <div key={dirPath} className="flex-list-container">
        <span className="flex-list-item-l">{dirPath}</span>
        <span className="flex-list-item-r">
          <button
            onClick={() => {
              this.deleteSharing(dirPath);
            }}
            className="grey1-bg white-font"
          >
            Delete
          </button>
        </span>
      </div>;
    });

    console.log("browser", this.props.sharings, this.props.isSharing);

    return (
      <div>
        <div id="op-bar" className="op-bar">
          <div className="margin-l-m margin-r-m">{ops}</div>
        </div>

        <div id="item-list">
          <div className="breadcrumb margin-b-l">
            <span
              className="white-font margin-r-m"
              style={{
                backgroundColor: "rgba(0, 0, 0, 0.7)",
                padding: "0.8rem 1rem",
                fontWeight: "bold",
                borderRadius: "0.5rem",
              }}
            >
              Location:
            </span>
            {breadcrumb}
          </div>

          {this.props.uploadings.size === 0 ? null : (
            <div className="container">
              <div className="flex-list-container bold">
                <span className="flex-list-item-l">
                  <span className="dot black-bg"></span>
                  <span>Uploading Files</span>
                </span>
                <span className="flex-list-item-r padding-r-m"></span>
              </div>
              {uploadingList}
            </div>
          )}

          {this.props.sharings.size === 0 ? null : (
            <div className="container">
              <div className="flex-list-container bold">
                <span className="flex-list-item-l">
                  <span className="dot black-bg"></span>
                  <span>Uploading Files</span>
                </span>
                <span className="flex-list-item-r padding-r-m"></span>
              </div>
              {sharingList}
            </div>
          )}

          <div className="container">
            <div className="flex-list-container bold">
              <span className="flex-list-item-l">
                <span className="dot black-bg"></span>
                <span>Name</span>
                {/* <span>File Size</span>
                  <span>Mod Time</span> */}
              </span>
              <span className="flex-list-item-r padding-r-m">
                <button
                  onClick={() => this.selectAll()}
                  className={`grey1-bg white-font`}
                  style={{ width: "8rem", display: "inline-block" }}
                >
                  Select All
                </button>
              </span>
            </div>
            {itemList}
          </div>
        </div>
      </div>
    );
  }
}
