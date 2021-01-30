import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map } from "immutable";
import FileSize from "filesize";

import { Layouter } from "./layouter";
import { alertMsg } from "../common/env";
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

export const uploadCheckCycle = 1000;

export interface Item {
  name: string;
  size: number;
  modTime: string;
  isDir: boolean;
  selected: boolean;
}

export interface Props {
  dirPath: List<string>;
  items: List<MetadataResp>;
  uploadings: List<UploadInfo>;

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
    const fileList = List<File>();
    for (let i = 0; i < event.target.files.length; i++) {
      fileList.push(event.target.files[i]);
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

    const layoutChildren = [
      <button
        type="button"
        onClick={() => this.delete()}
        className="red0-bg white-font margin-t-m margin-b-m"
      >
        Delete Selected
      </button>,
      <button
        type="button"
        onClick={() => this.moveHere()}
        className="grey1-bg white-font margin-t-m margin-b-m"
      >
        Paste
      </button>,
      <span className="inline-block margin-t-m margin-b-m">
        <input
          type="text"
          onChange={this.onInputChange}
          value={this.state.inputValue}
          className="black0-font margin-r-m"
          placeholder="folder name"
        />
        <button onClick={this.onMkDir} className="grey1-bg white-font">
          Create Folder
        </button>
      </span>,
      <span className="inline-block margin-t-m margin-b-m">
        <button onClick={this.onClickUpload} className="green0-bg white-font">
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
      </span>,
    ];

    const ops = (
      <Layouter isHorizontal={false} elements={layoutChildren}></Layouter>
    );

    const itemList = this.props.items.map((item: MetadataResp) => {
      const isSelected = this.state.selectedItems.has(item.name);
      const dirPath = this.props.dirPath.join("/");
      const itemPath = dirPath.endsWith("/")
        ? `${dirPath}${item.name}`
        : `${dirPath}/${item.name}`;

      return item.isDir ? (
        <tr
          key={item.name}
          // className={`${isSelected ? "white0-bg selected" : ""}`}
        >
          <td>
            <span className="dot yellow0-bg"></span>
          </td>
          <td>
            <span
              className={nameCellClass}
              onClick={() => this.gotoChild(item.name)}
            >
              {item.name}
            </span>
          </td>
          <td className={sizeCellClass}>--</td>
          <td className={modTimeCellClass}>
            {item.modTime.slice(0, item.modTime.indexOf("T"))}
          </td>
          <td>
            <span className="item-op">
              <button
                onClick={() => this.select(item.name)}
                className={`white-font ${isSelected ? "blue0-bg" : ""}`}
                style={{ width: "8rem", display: "inline-block" }}
              >
                {isSelected ? "Deselect" : "Select"}
              </button>
            </span>
          </td>
        </tr>
      ) : (
        <tr
          key={item.name}
          // className={`${isSelected ? "white0-bg selected" : ""}`}
        >
          <td>
            <span className="dot green0-bg"></span>
          </td>
          <td>
            <a
              className={nameCellClass}
              href={`/v1/fs/files?fp=${itemPath}`}
              target="_blank"
            >
              {item.name}
            </a>
          </td>
          <td className={sizeCellClass}>{FileSize(item.size, { round: 0 })}</td>
          <td className={modTimeCellClass}>
            {item.modTime.slice(0, item.modTime.indexOf("T"))}
          </td>
          <td>
            <span className="item-op">
              <button
                type="button"
                onClick={() => this.select(item.name)}
                className={`white-font ${isSelected ? "blue0-bg" : ""}`}
                style={{ width: "8rem", display: "inline-block" }}
              >
                {isSelected ? "Deselect" : "Select"}
              </button>
            </span>
          </td>
        </tr>
      );
    });

    const uploadingList = this.props.uploadings.map((uploading: UploadInfo) => {
      const pathParts = uploading.realFilePath.split("/");
      const fileName = pathParts[pathParts.length - 1];

      return (
        <tr key={fileName}>
          <td>
            <span className="dot blue0-bg"></span>
          </td>
          <td>
            <div className={nameCellClass}>{fileName}</div>
            <div className="item-op">
              <button
                onClick={() => this.stopUploading(uploading.realFilePath)}
                className="white-font margin-r-m"
              >
                Stop
              </button>
              <button
                onClick={() => this.deleteUpload(uploading.realFilePath)}
                className="white-font"
              >
                Delete
              </button>
            </div>
          </td>
          <td>{FileSize(uploading.uploaded, { round: 0 })}</td>
          <td>{FileSize(uploading.size, { round: 0 })}</td>
        </tr>
      );
    });

    return (
      <div>
        <div id="op-bar" className="op-bar">
          <div className="margin-l-m margin-r-m">{ops}</div>
        </div>

        <div id="item-list">
          <div className="margin-b-l">{breadcrumb}</div>

          {this.props.uploadings.size === 0 ? null : (
            <div className="container">
              <table>
                <thead style={{ fontWeight: "bold" }}>
                  <tr>
                    <td>
                      <span className="dot black-bg"></span>
                    </td>
                    <td>Name</td>
                    <td className={sizeCellClass}>Uploaded</td>
                    <td className={modTimeCellClass}>Size</td>
                  </tr>
                </thead>
                <tbody>{uploadingList}</tbody>
                <tfoot>
                  <tr>
                    <td></td>
                    <td></td>
                    <td className={sizeCellClass}></td>
                    <td className={modTimeCellClass}></td>
                  </tr>
                </tfoot>
              </table>
            </div>
          )}

          <div className="container">
            <table>
              <thead style={{ fontWeight: "bold" }}>
                <tr>
                  <td>
                    <span className="dot black-bg"></span>
                  </td>
                  <td>Name</td>
                  <td className={sizeCellClass}>File Size</td>
                  <td className={modTimeCellClass}>Mod Time</td>
                  <td>
                    <button
                      onClick={() => this.selectAll()}
                      className={`white-font`}
                      style={{ width: "8rem", display: "inline-block" }}
                    >
                      Select All
                    </button>
                  </td>
                </tr>
              </thead>
              <tbody>{itemList}</tbody>
              <tfoot>
                <tr>
                  <td></td>
                  <td></td>
                  <td className={sizeCellClass}></td>
                  <td className={modTimeCellClass}></td>
                  <td></td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      </div>
    );
  }
}
