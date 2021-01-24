import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map } from "immutable";
import FileSize from "filesize";

import { ICoreState } from "./core_state";
import {
  IUsersClient,
  IFilesClient,
  MetadataResp,
  UploadInfo,
} from "../client";
import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";
import { UploadMgr } from "../worker/upload_mgr";
import { UploadEntry } from "../worker/interface";
import { FileUploader } from "../worker/uploader";

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

  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

function getItemPath(dirPath: string, itemName: string): string {
  return dirPath.endsWith("/")
    ? `${dirPath}${itemName}`
    : `${dirPath}/${itemName}`;
}

export class Updater {
  private static props: Props;
  private static usersClient: IUsersClient;
  private static filesClient: IFilesClient;

  static init = (props: Props) => (Updater.props = { ...props });
  static setClients(usersClient: IUsersClient, filesClient: IFilesClient) {
    Updater.usersClient = usersClient;
    Updater.filesClient = filesClient;
  }

  static setUploadings = (infos: Map<string, UploadInfo>) => {
    Updater.props.uploadings = List<UploadInfo>(infos.values());
  };

  static setItems = async (dirParts: List<string>): Promise<void> => {
    const dirPath = dirParts.join("/");
    const listResp = await Updater.filesClient.list(dirPath);

    Updater.props.dirPath = dirParts;
    Updater.props.items =
      listResp.status === 200
        ? List<MetadataResp>(listResp.data.metadatas)
        : Updater.props.items;
  };

  static refreshUploadings = async (): Promise<boolean> => {
    const luResp = await Updater.filesClient.listUploadings();

    Updater.props.uploadings =
      luResp.status === 200
        ? List<UploadInfo>(luResp.data.uploadInfos)
        : Updater.props.uploadings;
    return luResp.status === 200;
  };

  static deleteUploading = async (filePath: string): Promise<boolean> => {
    UploadMgr.delete(filePath);
    const resp = await Updater.filesClient.deleteUploading(filePath);
    return resp.status === 200;
  };

  static stopUploading = (filePath: string) => {
    UploadMgr.stop(filePath);
  };

  static mkDir = async (dirPath: string): Promise<void> => {
    let resp = await Updater.filesClient.mkdir(dirPath);
    if (resp.status !== 200) {
      alert(`failed to make dir ${dirPath}`);
    }
  };

  static delete = async (
    dirParts: List<string>,
    items: List<MetadataResp>,
    selectedItems: Map<string, boolean>
  ): Promise<void> => {
    const delRequests = items
      .filter((item) => {
        return selectedItems.has(item.name);
      })
      .map(
        async (selectedItem: MetadataResp): Promise<string> => {
          const itemPath = getItemPath(dirParts.join("/"), selectedItem.name);
          const resp = await Updater.filesClient.delete(itemPath);
          return resp.status === 200 ? "" : selectedItem.name;
        }
      );

    const failedFiles = await Promise.all(delRequests);
    failedFiles.forEach((failedFile) => {
      if (failedFile !== "") {
        alert(`failed to delete ${failedFile}`);
      }
    });
    return Updater.setItems(dirParts);
  };

  static moveHere = async (
    srcDir: string,
    dstDir: string,
    selectedItems: Map<string, boolean>
  ): Promise<void> => {
    const moveRequests = List<string>(selectedItems.keys()).map(
      async (itemName: string): Promise<string> => {
        const oldPath = getItemPath(srcDir, itemName);
        const newPath = getItemPath(dstDir, itemName);
        const resp = await Updater.filesClient.move(oldPath, newPath);
        return resp.status === 200 ? "" : itemName;
      }
    );

    const failedFiles = await Promise.all(moveRequests);
    failedFiles.forEach((failedItem) => {
      if (failedItem !== "") {
        alert(`failed to move ${failedItem}`);
      }
    });

    return Updater.setItems(List<string>(dstDir.split("/")));
  };

  static setPwd = async (oldPwd: string, newPwd: string): Promise<boolean> => {
    const resp = await Updater.usersClient.setPwd(oldPwd, newPwd);
    return resp.status === 200;
  };

  static addUploadFiles = (fileList: FileList, len: number) => {
    for (let i = 0; i < len; i++) {
      // do not wait for the promise
      UploadMgr.add(fileList[i], fileList[i].name);
    }
    Updater.setUploadings(
      UploadMgr.list().map((entry: UploadEntry) => {
        return {
          realFilePath: entry.filePath,
          size: entry.size,
          uploaded: entry.uploaded,
        };
      })
    );
  };

  static setBrowser = (prevState: ICoreState): ICoreState => {
    prevState.panel.browser = { ...prevState.panel, ...Updater.props };
    return prevState;
  };
}

export interface State {
  inputValue: string;
  selectedSrc: string;
  selectedItems: Map<string, boolean>;

  show: boolean;
  oldPwd: string;
  newPwd1: string;
  newPwd2: string;
}

export class Browser extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;
  private uploadInput: Element | Text;
  private assignInput: (input: Element) => void;
  private onClickUpload: () => void;

  constructor(p: Props) {
    super(p);
    Updater.init(p);
    Updater.setClients(new UsersClient(""), new FilesClient(""));
    this.update = p.update;
    this.state = {
      inputValue: "",
      selectedSrc: "",
      selectedItems: Map<string, boolean>(),
      show: false,
      oldPwd: "",
      newPwd1: "",
      newPwd2: "",
    };

    this.uploadInput = undefined;
    this.assignInput = (input) => {
      this.uploadInput = ReactDOM.findDOMNode(input);
    };
    this.onClickUpload = () => {
      // TODO: check if the re-upload file is same as previous upload
      const uploadInput = this.uploadInput as HTMLButtonElement;
      uploadInput.click();
    };

    // UploadMgr.setStatusCb(this.updateProgress);
    Updater.setItems(p.dirPath)
      .then(() => {
        return Updater.refreshUploadings();
      })
      .then((_: boolean) => {
        this.update(Updater.setBrowser);
      });
  }

  showPane = () => {
    this.setState({ show: !this.state.show });
  };
  changeOldPwd = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ oldPwd: ev.target.value });
  };
  changeNewPwd1 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd1: ev.target.value });
  };
  changeNewPwd2 = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newPwd2: ev.target.value });
  };
  onInputChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ inputValue: ev.target.value });
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

  addUploadFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    Updater.addUploadFiles(event.target.files, event.target.files.length);
    this.update(Updater.setBrowser);
  };

  updateProgress = (infos: Map<string, UploadInfo>) => {
    Updater.setUploadings(infos);
    Updater.setItems(this.props.dirPath).then(() => {
      this.update(Updater.setBrowser);
    });
  };

  onMkDir = () => {
    Updater.mkDir(this.state.inputValue)
      .then(() => {
        this.setState({ inputValue: "" });
        return Updater.setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(Updater.setBrowser);
      });
  };

  deleteUploading = (filePath: string) => {
    Updater.deleteUploading(filePath)
      .then((ok: boolean) => {
        if (!ok) {
          alert(`Failed to delete uploading ${filePath}`);
        }
        return Updater.refreshUploadings();
      })
      .then(() => {
        this.update(Updater.setBrowser);
      });
  };

  stopUploading = (filePath: string) => {
    Updater.stopUploading(filePath);
    this.update(Updater.setBrowser);
  };

  delete = () => {
    if (this.props.dirPath.join("/") !== this.state.selectedSrc) {
      alert("please select file or folder to delete at first");
      this.setState({
        selectedSrc: this.props.dirPath.join("/"),
        selectedItems: Map<string, boolean>(),
      });
      return;
    }

    Updater.delete(
      this.props.dirPath,
      this.props.items,
      this.state.selectedItems
    ).then(() => {
      this.update(Updater.setBrowser);
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

    Updater.setItems(dirPath).then(() => {
      this.update(Updater.setBrowser);
    });
  };

  moveHere = () => {
    const oldDir = this.state.selectedSrc;
    const newDir = this.props.dirPath.join("/");
    if (oldDir === newDir) {
      alert("source directory is same as destination directory");
      return;
    }

    Updater.moveHere(
      this.state.selectedSrc,
      this.props.dirPath.join("/"),
      this.state.selectedItems
    ).then(() => {
      this.update(Updater.setBrowser);
      this.setState({
        selectedSrc: "",
        selectedItems: Map<string, boolean>(),
      });
    });
  };

  setPwd = () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      alert("new passwords are not same");
    } else if (this.state.newPwd1 == "") {
      alert("new passwords can not be empty");
    } else if (this.state.oldPwd == this.state.newPwd1) {
      alert("old and new passwords are same");
    } else {
      Updater.setPwd(this.state.oldPwd, this.state.newPwd1).then(
        (ok: boolean) => {
          if (ok) {
            alert("Password is updated");
          } else {
            alert("fail to update password");
          }
          this.setState({
            oldPwd: "",
            newPwd1: "",
            newPwd2: "",
          });
        }
      );
    }
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

    const ops = (
      <div>
        <div className="grey0-font">
          <button
            type="button"
            onClick={() => this.delete()}
            className="red0-bg white-font margin-m"
          >
            Delete Selected
          </button>
          <span className="margin-s">-</span>
          <button
            type="button"
            onClick={() => this.moveHere()}
            className="grey1-bg white-font margin-m"
          >
            Paste
          </button>
          <span className="margin-s">-</span>
          <button
            onClick={this.onClickUpload}
            className="green0-bg white-font margin-m"
          >
            Upload Files
          </button>
          <span className="margin-s">-</span>
          <span className="margin-m">
            <input
              type="text"
              onChange={this.onInputChange}
              value={this.state.inputValue}
              className="margin-r-m black0-font"
              placeholder="folder name"
            />
            <button onClick={this.onMkDir} className="grey1-bg white-font">
              Create Folder
            </button>
          </span>
          <input
            type="file"
            onChange={this.addUploadFile}
            multiple={true}
            value={this.props.uploadValue}
            ref={this.assignInput}
            className="black0-font hidden margin-m"
          />
          <span className="margin-s">-</span>
          <button
            onClick={this.showPane}
            className="grey1-bg white-font margin-m"
          >
            Settings
          </button>
        </div>
        <div>
          <div
            style={{ display: this.state.show ? "inherit" : "none" }}
            className="margin-t-m"
          >
            <h3 className="padding-l-s grey0-font">Update Password</h3>
            <input
              name="old_pwd"
              type="password"
              onChange={this.changeOldPwd}
              value={this.state.oldPwd}
              className="margin-m black0-font"
              placeholder="old password"
            />
            <input
              name="new_pwd1"
              type="password"
              onChange={this.changeNewPwd1}
              value={this.state.newPwd1}
              className="margin-m black0-font"
              placeholder="new password"
            />
            <input
              name="new_pwd2"
              type="password"
              onChange={this.changeNewPwd2}
              value={this.state.newPwd2}
              className="margin-m black0-font"
              placeholder="new password again"
            />
            <button onClick={this.setPwd} className="grey1-bg white-font">
              Update
            </button>
          </div>
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
        <tr
          key={item.name}
          className={`${isSelected ? "white0-bg selected" : ""}`}
        >
          <td className="padding-l-l" style={{ width: "3rem" }}>
            <span className="dot yellow0-bg"></span>
          </td>
          <td>
            <span
              className="item-name pointer"
              onClick={() => this.gotoChild(item.name)}
            >
              {item.name}
            </span>
          </td>
          <td>--</td>
          <td>{item.modTime.slice(0, item.modTime.indexOf("T"))}</td>

          <td>
            <button
              onClick={() => this.select(item.name)}
              className="white-font margin-t-m margin-b-m"
            >
              {isSelected ? "Unselect" : "Select"}
            </button>
          </td>
        </tr>
      ) : (
        <tr
          key={item.name}
          className={`${isSelected ? "white0-bg selected" : ""}`}
        >
          <td className="padding-l-l" style={{ width: "3rem" }}>
            <span className="dot green0-bg"></span>
          </td>
          <td>
            <a
              className="item-name"
              href={`/v1/fs/files?fp=${itemPath}`}
              target="_blank"
            >
              {item.name}
            </a>
          </td>
          <td>{FileSize(item.size, { round: 0 })}</td>
          <td>{item.modTime.slice(0, item.modTime.indexOf("T"))}</td>

          <td>
            <button
              type="button"
              onClick={() => this.select(item.name)}
              className="white-font margin-t-m margin-b-m"
            >
              {isSelected ? "Unselect" : "Select"}
            </button>
          </td>
        </tr>
      );
    });

    const uploadingList = this.props.uploadings.map((uploading: UploadInfo) => {
      const pathParts = uploading.realFilePath.split("/");
      const fileName = pathParts[pathParts.length - 1];

      return (
        <tr key={fileName}>
          <td className="padding-l-l" style={{ width: "3rem" }}>
            <span className="dot blue0-bg"></span>
          </td>
          <td>
            <span className="item-name pointer">{fileName}</span>
          </td>
          <td>{FileSize(uploading.uploaded, { round: 0 })}</td>
          <td>{FileSize(uploading.size, { round: 0 })}</td>
          <td>
            <button
              onClick={() => this.stopUploading(uploading.realFilePath)}
              className="white-font margin-m"
            >
              Stop
            </button>
            <button
              onClick={() => this.deleteUploading(uploading.realFilePath)}
              className="white-font margin-m"
            >
              Delete
            </button>
          </td>
        </tr>
      );
    });

    return (
      <div>
        <div id="op-bar" className="op-bar">
          <div className="margin-l-m margin-r-m">{ops}</div>
        </div>

        <div id="item-list" className="">
          <div className="margin-b-l">{breadcrumb}</div>

          <table>
            <thead style={{ fontWeight: "bold" }}>
              <tr>
                <td className="padding-l-l" style={{ width: "3rem" }}>
                  <span className="dot black-bg"></span>
                </td>
                <td>Name</td>
                <td>Uploaded</td>
                <td>Size</td>
                <td>Action</td>
              </tr>
            </thead>
            <tbody>{uploadingList}</tbody>
            <tfoot>
              <tr>
                <td></td>
                <td></td>
                <td></td>
                <td></td>
                <td></td>
              </tr>
            </tfoot>
          </table>

          <table>
            <thead style={{ fontWeight: "bold" }}>
              <tr>
                <td className="padding-l-l" style={{ width: "3rem" }}>
                  <span className="dot black-bg"></span>
                </td>
                <td>Name</td>
                <td>File Size</td>
                <td>Mod Time</td>
                <td>Edit</td>
              </tr>
            </thead>
            <tbody>{itemList}</tbody>
            <tfoot>
              <tr>
                <td></td>
                <td></td>
                <td></td>
                <td></td>
                <td></td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    );
  }
}
