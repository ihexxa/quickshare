import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map } from "immutable";

import { ICoreState } from "./core_state";
import { filesClient, usersClient } from "../client";
import { MetadataResp } from "../client/files";
import { FileUploader } from "../client/uploader";

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

  uploadFiles: List<File>;
  uploadValue: string;

  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export class Updater {
  private static props: Props;

  static init = (props: Props) => (Updater.props = { ...props });

  static setItems = async (dirParts: List<string>): Promise<void> => {
    let dirPath = dirParts.join("/");
    let listResp = await filesClient.list(dirPath);

    Updater.props.dirPath = dirParts;
    Updater.props.items =
      listResp != null
        ? List<MetadataResp>(listResp.metadatas)
        : Updater.props.items;
  };

  static mkDir = async (dirPath: string): Promise<void> => {
    let status = await filesClient.mkdir(dirPath);
    if (status != 200) {
      // set err
    }
  };

  static delete = async (itemPath: string): Promise<void> => {
    let status = await filesClient.delete(itemPath);
    if (status != 200) {
      // set err
    }
  };

  static move = async (oldPath: string, newPath: string): Promise<void> => {
    let status = await filesClient.move(oldPath, newPath);
    if (status != 200) {
      // set err
    }
  };

  static setPwd = async (oldPwd: string, newPwd: string): Promise<boolean> => {
    const status = await usersClient.setPwd(oldPwd, newPwd);
    return status == 200;
  };

  static addUploadFiles = (uploadFiles: List<File>) => {
    Updater.props.uploadFiles = Updater.props.uploadFiles.concat(uploadFiles);
  };

  static setUploadFiles = (uploadFiles: List<File>) => {
    Updater.props.uploadFiles = uploadFiles;
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

  constructor(p: Props) {
    super(p);
    Updater.init(p);
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

    Updater.setItems(p.dirPath).then(() => {
      this.update(Updater.setBrowser);
    });
  }

  componentDidMount() {
    this.startUploadWorker();
  }

  startUploadWorker = () => {
    if (this.props.uploadFiles.size > 0) {
      const file = this.props.uploadFiles.get(0);
      Updater.setUploadFiles(this.props.uploadFiles.slice(0, -1));
      this.update(Updater.setBrowser);

      const uploader = new FileUploader(
        file,
        `${this.props.dirPath.join("/")}/${file.name}`,
        this.startUploadWorker
      );

      uploader.start().then(() => {
        Updater.setItems(this.props.dirPath).then(() => {
          this.update(Updater.setBrowser);
          this.startUploadWorker();
        });
      });
    } else {
      setTimeout(this.startUploadWorker, uploadCheckCycle);
    }
  };

  onClickUpload = () => {
    const uploadInput = this.uploadInput as HTMLButtonElement;
    uploadInput.click();
  };

  onUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files == null || event.target.files.length === 0) {
      // TODO: prompt an alert
    } else {
      let newUploads = List<File>([]);
      for (let i = 0; i < event.target.files.length; i++) {
        newUploads = newUploads.push(event.target.files.item(i));
      }

      Updater.addUploadFiles(newUploads);
      this.update(Updater.setBrowser);
    }
  };

  onInputChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    const val = ev.target.value;
    this.setState((prevState: State, _: Props) => {
      return { ...prevState, inputValue: val };
    });
  };

  onMkDir = () => {
    Updater.mkDir(this.state.inputValue)
      .then(() => {
        return Updater.setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(Updater.setBrowser);
      });
  };

  delete = () => {
    // TODO: Add checks and clean selected when change the dir
    const delPromises = this.props.items
      .filter((item) => {
        return this.state.selectedItems.has(item.name);
      })
      .map(
        (selectedItem: MetadataResp): Promise<void> => {
          const dirPath = this.props.dirPath.join("/");
          const itemPath = dirPath.endsWith("/")
            ? `${dirPath}${selectedItem.name}`
            : `${dirPath}/${selectedItem.name}`;

          return Updater.delete(itemPath);
        }
      );

    // TODO: clean selected
    Promise.all(delPromises)
      .then(() => {
        return Updater.setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(Updater.setBrowser);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  goto = (childDirName: string) => {
    const dirPath = this.props.dirPath.push(childDirName);
    Updater.setItems(dirPath).then(() => {
      this.update(Updater.setBrowser);
    });
  };

  chdir = (dirPath: List<string>) => {
    Updater.setItems(dirPath).then(() => {
      this.update(Updater.setBrowser);
    });
  };

  select = (itemName: string) => {
    const selectedItems = this.state.selectedItems.has(itemName)
      ? this.state.selectedItems.delete(itemName)
      : this.state.selectedItems.set(itemName, true);

    this.setState((prevState: State, _: Props) => {
      return {
        ...prevState,
        selectedSrc: this.props.dirPath.join("/"),
        selectedItems: selectedItems,
      };
    });
  };

  moveHere = () => {
    // TODO: Add checks and clean selected when change the dir
    const movePromises = this.props.items
      .filter((item) => {
        return this.state.selectedItems.has(item.name);
      })
      .map(
        (selectedItem: MetadataResp): Promise<void> => {
          const itemName = selectedItem.name;
          const oldPath = this.state.selectedSrc.endsWith("/")
            ? `${this.state.selectedSrc}${itemName}`
            : `${this.state.selectedSrc}/${itemName}`;
          const newPath = this.props.dirPath.join("/").endsWith("/")
            ? `${this.props.dirPath.join("/")}${itemName}`
            : `${this.props.dirPath.join("/")}/${itemName}`;
          return Updater.move(oldPath, newPath);
        }
      );

    // TODO: clean selected
    Promise.all(movePromises)
      .then(() => {
        return Updater.setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(Updater.setBrowser);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  // copyHere = () => {};

  setPwd = () => {
    if (this.state.newPwd1 !== this.state.newPwd2) {
      // alert
      alert("new pwds not same");
      return;
    }
    if (this.state.newPwd1 == "") {
      alert("new pwds can not be empty");
      return;
    }
    if (this.state.oldPwd == this.state.newPwd1) {
      // alert
      alert("old and new pwds are same");
      return;
    }
    Updater.setPwd(this.state.oldPwd, this.state.newPwd1).then(
      (ok: boolean) => {
        if (ok) {
          // hint
          alert("ok");
        } else {
          // alert
        }
      }
    );
  };

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
            className="red0-bg white-font"
          >
            Delete Selected
          </button>
          <span className="margin-l-l margin-r-l">/</span>
          <button
            type="button"
            onClick={() => this.moveHere()}
            className="grey1-bg white-font"
          >
            Paste
          </button>
          <span className="margin-l-l margin-r-l">/</span>
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
          <span className="margin-l-l margin-r-l">/</span>
          <button onClick={this.onClickUpload} className="green0-bg white-font">
            Upload Files
          </button>
          <input
            type="file"
            onChange={this.onUpload}
            multiple={true}
            value={this.props.uploadValue}
            ref={this.assignInput}
            className="black0-font hidden"
          />
          <span className="margin-l-l margin-r-l">/</span>
          <button onClick={this.showPane} className="grey1-bg white-font">
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
              className="margin-r-m black0-font"
              placeholder="old password"
            />
            <input
              name="new_pwd1"
              type="password"
              onChange={this.changeNewPwd1}
              value={this.state.newPwd1}
              className="margin-r-m black0-font"
              placeholder="new password"
            />
            <input
              name="new_pwd2"
              type="password"
              onChange={this.changeNewPwd2}
              value={this.state.newPwd2}
              className="margin-r-m black0-font"
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
          className={`${isSelected ? "green0-bg white-font" : "white0-bg"}`}
          onClick={() => this.select(item.name)}
        >
          <td className="padding-l-l" style={{width: "3rem"}}>
            <span className="dot green0-bg" ></span>
          </td>
          <td>
            <span onClick={() => this.goto(item.name)}>{item.name}</span>
          </td>
          <td>N/A</td>
          <td>{item.modTime.slice(0, item.modTime.indexOf("T"))}</td>
          <td>Folder</td>
        </tr>
      ) : (
        <tr
          key={item.name}
          className={`${isSelected ? "green0-bg white-font" : "white0-bg"}`}
          onClick={() => this.select(item.name)}
        >
          <td className="padding-l-l" style={{width: "3rem"}}>
            <span className="dot green0-bg" ></span>
          </td>
          <td>
            <a href={`/v1/fs/files?fp=${itemPath}`} target="_blank">
              {item.name}
            </a>
          </td>
          <td>{item.size}</td>
          <td>{item.modTime.slice(0, item.modTime.indexOf("T"))}</td>
          <td>Regular File</td>
        </tr>
      );
    });
    return (
      <div>
        <div id="op-bar" className="op-bar padding-t-m padding-b-m">
          <div className="margin-l-m margin-r-m">{ops}</div>
        </div>

        <div id="item-list" className="">
          <div className="margin-b-l">{breadcrumb}</div>
          <table>
            <thead style={{fontWeight:"bold"}}>
              <tr>
                <td></td>
                <td>Entry Name</td>
                <td>File Size</td>
                <td>Mod Time</td>
                <td>Type</td>
              </tr>
            </thead>
            <tbody>{itemList}</tbody>
          </table>
        </div>
      </div>
    );
  }
}
