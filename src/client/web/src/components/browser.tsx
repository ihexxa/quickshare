import * as React from "react";
import * as ReactDOM from "react-dom";
import { List } from "immutable";

import { ICoreState } from "./core_state";
import { filesClient } from "../client";
import { MetadataResp } from "../client/files";
import { FileUploader } from "../client/uploader";

export const uploadCheckCycle = 1000;

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
  selectedItems: List<string>;
}

export class Browser extends React.Component<Props, State, {}> {
  private update: (updater: (prevState: ICoreState) => ICoreState) => void;

  constructor(p: Props) {
    super(p);
    Updater.init(p);
    this.update = p.update;
    this.state = {
      inputValue: "",
      selectedSrc: "",
      selectedItems: List<string>([]),
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

  delete = (itemName: string) => {
    const dirPath = this.props.dirPath.join("/");
    const itemPath = dirPath.endsWith("/")
      ? `${dirPath}${itemName}`
      : `${dirPath}/${itemName}`;

    Updater.delete(itemPath)
      .then(() => {
        return Updater.setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(Updater.setBrowser);
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
    this.setState((prevState: State, _: Props) => {
      return {
        ...prevState,
        selectedSrc: this.props.dirPath.join("/"),
        selectedItems: prevState.selectedItems.push(itemName),
      };
    });
  };

  moveHere = () => {
    // TODO: Add checks and clean selected when change the dir
    const movePromises = this.state.selectedItems.map(
      (selectedItem: string): Promise<void> => {
        const oldPath = this.state.selectedSrc.endsWith("/")
          ? `${this.state.selectedSrc}${selectedItem}`
          : `${this.state.selectedSrc}/${selectedItem}`;
        const newPath = this.props.dirPath.join("/").endsWith("/")
          ? `${this.props.dirPath.join("/")}${selectedItem}`
          : `${this.props.dirPath.join("/")}/${selectedItem}`;
        return Updater.move(oldPath, newPath);
      }
    );

    Promise.all(movePromises)
      .then(() => {
        return Updater.setItems(this.props.dirPath);
      })
      .then(() => {
        this.update(Updater.setBrowser);
        this.setState({
          ...this.state,
          selectedSrc: "",
          selectedItems: List<string>([]),
        });
      });
  };

  // copyHere = () => {};

  render() {
    const breadcrumb = this.props.dirPath.map(
      (pathPart: string, key: number) => {
        return (
          <span key={pathPart}>
            <button
              type="button"
              onClick={() => this.chdir(this.props.dirPath.slice(0, key + 1))}
              className="grey1-bg white-font margin-r-m"
            >
              {pathPart}
            </button>
          </span>
        );
      }
    );
    const ops = (
      <div>
        <button
          type="button"
          onClick={() => this.moveHere()}
          className="grey1-bg white-font margin-r-m"
        >
          Move here
        </button>
        <input
          type="text"
          onChange={this.onInputChange}
          value={this.state.inputValue}
          className="margin-r-m black0-font"
        />
        <button
          onClick={this.onMkDir}
          className="grey1-bg white-font margin-r-m"
        >
          MkDir
        </button>
        <input
          type="file"
          onChange={this.onUpload}
          multiple={true}
          value={this.props.uploadValue}
          className="black0-font"
        />
      </div>
    );
    const itemList = this.props.items.map((item: MetadataResp) => {
      return (
        <span key={item.name} className="grid margin-r-m margin-t-m white0-bg">
          <span className="dot margin-r-m green0-bg"></span>
          <span className="margin-r-m">
            {item.name}
          </span>
          <button
            type="button"
            onClick={() => this.delete(item.name)}
            className="grey1-bg white-font margin-r-m"
          >
            Del
          </button>
          <button
            type="button"
            onClick={() => this.select(item.name)}
            className="grey1-bg white-font margin-r-m"
          >
            Select
          </button>
          {item.isDir ? (
            <button
              type="button"
              onClick={() => this.goto(item.name)}
              className="grey1-bg white-font margin-r-m"
            >
              goto
            </button>
          ) : null}
        </span>
      );
    });
    return (
      <div>
        <div className="flex-2col-parent padding-l">
          <div className="flex-2col">{breadcrumb}</div>
          <div className="flex-2col text-right ">{ops}</div>
        </div>
        <div className="padding-l">{itemList}</div>
      </div>
    );
  }
}
