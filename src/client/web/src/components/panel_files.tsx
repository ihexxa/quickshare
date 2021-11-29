import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiHomeSmileFill } from "@react-icons/all-files/ri/RiHomeSmileFill";
import { RiFile2Fill } from "@react-icons/all-files/ri/RiFile2Fill";

import { alertMsg, confirmMsg } from "../common/env";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { MetadataResp, roleVisitor, roleAdmin } from "../client";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { Up } from "../worker/upload_mgr";
import { UploadEntry, UploadState } from "../worker/interface";
import { getIcon } from "./visual/icons";

export interface Item {
  name: string;
  size: number;
  modTime: string;
  isDir: boolean;
  selected: boolean;
  sha1: string;
}

export interface FilesProps {
  dirPath: List<string>;
  isSharing: boolean;
  items: List<MetadataResp>;
}

export interface Props {
  filesInfo: FilesProps;
  msg: MsgProps;
  login: LoginProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export function getItemPath(dirPath: string, itemName: string): string {
  return dirPath.endsWith("/")
    ? `${dirPath}${itemName}`
    : `${dirPath}/${itemName}`;
}

export interface State {
  newFolderName: string;
  selectedSrc: string;
  selectedItems: Map<string, boolean>;
  showDetail: Set<string>;
  uploadFiles: string;
}

export class FilesPanel extends React.Component<Props, State, {}> {
  private uploadInput: Element | Text;
  private assignInput: (input: Element) => void;
  private onClickUpload: () => void;

  constructor(p: Props) {
    super(p);
    this.state = {
      newFolderName: "",
      selectedSrc: "",
      selectedItems: Map<string, boolean>(),
      showDetail: Set<string>(),
      uploadFiles: "",
    };

    Up().setStatusCb(this.updateProgress);
    this.uploadInput = undefined;
    this.assignInput = (input) => {
      this.uploadInput = ReactDOM.findDOMNode(input);
    };
    this.onClickUpload = () => {
      const uploadInput = this.uploadInput as HTMLButtonElement;
      uploadInput.click();
    };
  }

  updateProgress = async (
    infos: Map<string, UploadEntry>,
    refresh: boolean
  ) => {
    updater().setUploadings(infos);
    let errCount = 0;
    infos.valueSeq().forEach((entry: UploadEntry) => {
      errCount += entry.state === UploadState.Error ? 1 : 0;
    });

    if (infos.size === 0 || infos.size === errCount) {
      // refresh used space
      updater()
        .self()
        .then(() => {
          this.props.update(updater().updateLogin);
        });
    }

    if (refresh) {
      updater()
        .setItems(this.props.filesInfo.dirPath)
        .then(() => {
          this.props.update(updater().updateFilesInfo);
        });
    } else {
      this.props.update(updater().updateFilesInfo);
    }
  };

  addUploads = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files.length > 1000) {
      alertMsg(this.props.msg.pkg.get("err.tooManyUploads"));
      return;
    }

    let fileList = List<File>();
    for (let i = 0; i < event.target.files.length; i++) {
      fileList = fileList.push(event.target.files[i]);
    }
    updater().addUploads(fileList);
    this.props.update(updater().updateFilesInfo);
    this.props.update(updater().updateSharingsInfo);
  };

  onNewFolderNameChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newFolderName: ev.target.value });
  };

  onMkDir = () => {
    if (this.state.newFolderName === "") {
      alertMsg(this.props.msg.pkg.get("browser.folder.add.fail"));
      return;
    }

    const dirPath = getItemPath(
      this.props.filesInfo.dirPath.join("/"),
      this.state.newFolderName
    );
    updater()
      .mkDir(dirPath)
      .then(() => {
        this.setState({ newFolderName: "" });
        return updater().setItems(this.props.filesInfo.dirPath);
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
      });
  };

  delete = () => {
    // TODO: selected should be cleaned after change the cwd
    if (this.props.filesInfo.dirPath.join("/") !== this.state.selectedSrc) {
      alertMsg(this.props.msg.pkg.get("browser.del.fail"));
      this.setState({
        selectedSrc: this.props.filesInfo.dirPath.join("/"),
        selectedItems: Map<string, boolean>(),
      });
      return;
    } else {
      const filesToDel = this.state.selectedItems.keySeq().join(", ");
      if (
        !confirmMsg(
          `${this.props.msg.pkg.get("op.confirm")} [${
            this.state.selectedItems.size
          }]: ${filesToDel}`
        )
      ) {
        return;
      }
    }

    updater()
      .delete(
        this.props.filesInfo.dirPath,
        this.props.filesInfo.items,
        this.state.selectedItems
      )
      .then(() => {
        return updater().self();
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
        this.props.update(updater().updateLogin);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  moveHere = () => {
    const oldDir = this.state.selectedSrc;
    const newDir = this.props.filesInfo.dirPath.join("/");
    if (oldDir === newDir) {
      alertMsg(this.props.msg.pkg.get("browser.move.fail"));
      return;
    }

    updater()
      .moveHere(
        this.state.selectedSrc,
        this.props.filesInfo.dirPath.join("/"),
        this.state.selectedItems
      )
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  gotoChild = async (childDirName: string) => {
    return this.chdir(this.props.filesInfo.dirPath.push(childDirName));
  };

  chdir = async (dirPath: List<string>) => {
    if (dirPath === this.props.filesInfo.dirPath) {
      return;
    } else if (this.props.login.userRole !== roleAdmin && dirPath.size <= 1) {
      alertMsg(this.props.msg.pkg.get("unauthed"));
      return;
    }

    return updater()
      .setItems(dirPath)
      .then(() => {
        return updater().listSharings();
      })
      .then(() => {
        return updater().isSharing(dirPath.join("/"));
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
      });
  };

  select = (itemName: string) => {
    const selectedItems = this.state.selectedItems.has(itemName)
      ? this.state.selectedItems.delete(itemName)
      : this.state.selectedItems.set(itemName, true);

    this.setState({
      selectedSrc: this.props.filesInfo.dirPath.join("/"),
      selectedItems: selectedItems,
    });
  };

  selectAll = () => {
    let newSelected = Map<string, boolean>();
    const someSelected = this.state.selectedItems.size === 0 ? true : false;
    if (someSelected) {
      this.props.filesInfo.items.forEach((item) => {
        newSelected = newSelected.set(item.name, true);
      });
    } else {
      this.props.filesInfo.items.forEach((item) => {
        newSelected = newSelected.delete(item.name);
      });
    }

    this.setState({
      selectedSrc: this.props.filesInfo.dirPath.join("/"),
      selectedItems: newSelected,
    });
  };

  toggleDetail = (name: string) => {
    const showDetail = this.state.showDetail.has(name)
      ? this.state.showDetail.delete(name)
      : this.state.showDetail.add(name);
    this.setState({ showDetail });
  };

  generateHash = async (filePath: string): Promise<boolean> => {
    alertMsg(this.props.msg.pkg.get("refresh-hint"));
    return updater().generateHash(filePath);
  };

  addSharing = async () => {
    return updater()
      .addSharing()
      .then((ok) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.share.add.fail"));
        } else {
          updater().setSharing(true);
        }
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
      });
  };

  deleteSharing = async (dirPath: string) => {
    return updater()
      .deleteSharing(dirPath)
      .then((ok) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.share.del.fail"));
        } else {
          updater().setSharing(false);
        }
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
      });
  };

  render() {
    const showOp = this.props.login.userRole === roleVisitor ? "hidden" : "";
    const breadcrumb = this.props.filesInfo.dirPath.map(
      (pathPart: string, key: number) => {
        return (
          <button
            key={pathPart}
            onClick={() =>
              this.chdir(this.props.filesInfo.dirPath.slice(0, key + 1))
            }
            className="item"
          >
            {pathPart}
          </button>
        );
      }
    );

    const nameWidthClass = `item-name item-name-${
      this.props.ui.isVertical ? "vertical" : "horizontal"
    } pointer`;

    const ops = (
      <div id="upload-op">
        <div className="float">
          <input
            type="text"
            onChange={this.onNewFolderNameChange}
            value={this.state.newFolderName}
            placeholder={this.props.msg.pkg.get("browser.folder.name")}
            className="float"
          />
          <button onClick={this.onMkDir} className="float">
            {this.props.msg.pkg.get("browser.folder.add")}
          </button>
        </div>

        <div className="float">
          <button onClick={this.onClickUpload}>
            {this.props.msg.pkg.get("browser.upload")}
          </button>
          <input
            type="file"
            onChange={this.addUploads}
            multiple={true}
            value={this.state.uploadFiles}
            ref={this.assignInput}
            className="hidden"
          />
        </div>
      </div>
    );

    const sortedItems = this.props.filesInfo.items.sort(
      (item1: MetadataResp, item2: MetadataResp) => {
        if (item1.isDir && !item2.isDir) {
          return -1;
        } else if (!item1.isDir && item2.isDir) {
          return 1;
        }
        return 0;
      }
    );

    const itemList = sortedItems.map((item: MetadataResp) => {
      const isSelected = this.state.selectedItems.has(item.name);
      const dirPath = this.props.filesInfo.dirPath.join("/");
      const itemPath = dirPath.endsWith("/")
        ? `${dirPath}${item.name}`
        : `${dirPath}/${item.name}`;

      return item.isDir ? (
        <Flexbox
          key={item.name}
          children={List([
            <Flexbox
              children={List([
                <RiFolder2Fill
                  size="3rem"
                  className="yellow0-font margin-r-m"
                />,

                <span className={`${nameWidthClass}`}>
                  <span
                    className="title-m"
                    onClick={() => this.gotoChild(item.name)}
                  >
                    {item.name}
                  </span>
                  <div className="desc-m grey0-font">
                    <span>
                      {item.modTime.slice(0, item.modTime.indexOf("T"))}
                    </span>
                  </div>
                </span>,
              ])}
              childrenStyles={List([
                { flex: "0 0 auto" },
                { flex: "0 0 auto" },
              ])}
            />,
            <span className={`item-op ${showOp}`}>
              <span onClick={() => this.select(item.name)} className="float-l">
                {isSelected
                  ? getIcon("RiCheckboxFill", "1.8rem", "cyan0")
                  : getIcon("RiCheckboxBlankFill", "1.8rem", "grey1")}
              </span>
            </span>,
          ])}
          childrenStyles={List([
            { flex: "0 0 auto", width: "60%" },
            { flex: "0 0 auto", justifyContent: "flex-end", width: "40%" },
          ])}
        />
      ) : (
        <div key={item.name}>
          <Flexbox
            key={item.name}
            children={List([
              <Flexbox
                children={List([
                  <RiFile2Fill size="3rem" className="cyan0-font margin-r-m" />,

                  <span className={`${nameWidthClass}`}>
                    <a
                      className="title-m"
                      href={`/v1/fs/files?fp=${itemPath}`}
                      target="_blank"
                    >
                      {item.name}
                    </a>
                    <div className="desc-m grey0-font">
                      <span>
                        {item.modTime.slice(0, item.modTime.indexOf("T"))}
                      </span>
                      &nbsp;/&nbsp;
                      <span>{FileSize(item.size, { round: 0 })}</span>
                    </div>
                  </span>,
                ])}
                childrenStyles={List([
                  { flex: "0 0 auto" },
                  { flex: "0 0 auto" },
                ])}
              />,

              <span className={`item-op ${showOp}`}>
                <span
                  onClick={() => this.toggleDetail(item.name)}
                  className="float-l"
                >
                  {getIcon("RiInformationFill", "1.8rem", "grey1")}
                </span>

                <span
                  onClick={() => this.select(item.name)}
                  className="float-l"
                >
                  {isSelected
                    ? getIcon("RiCheckboxFill", "1.8rem", "cyan0")
                    : getIcon("RiCheckboxBlankFill", "1.8rem", "grey1")}
                </span>
              </span>,
            ])}
            childrenStyles={List([
              { flex: "0 0 auto", width: "60%" },
              { flex: "0 0 auto", justifyContent: "flex-end", width: "40%" },
            ])}
          />

          <div
            className={`${
              this.state.showDetail.has(item.name) ? "" : "hidden"
            }`}
          >
            <Flexbox
              children={List([
                <span>
                  <div className="label">{`SHA1: `}</div>
                  <input type="text" readOnly={true} value={`${item.sha1}`} />
                </span>,
                <button onClick={() => this.generateHash(itemPath)}>
                  {this.props.msg.pkg.get("refresh")}
                </button>,
              ])}
              className="item-info"
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
          </div>
        </div>
      );
    });

    const usedSpace = FileSize(parseInt(this.props.login.usedSpace, 10), {
      round: 0,
    });
    const spaceLimit = FileSize(
      parseInt(this.props.login.quota.spaceLimit, 10),
      {
        round: 0,
      }
    );

    const itemListPane = (
      <div id="item-list">
        <div className={showOp}>
          <Container>{ops}</Container>
        </div>

        <Container>
          <div id="browser-op" className={`${showOp}`}>
            <Flexbox
              children={List([
                <span>
                  {this.props.filesInfo.isSharing ? (
                    <button
                      type="button"
                      onClick={() => {
                        this.deleteSharing(
                          this.props.filesInfo.dirPath.join("/")
                        );
                      }}
                      className="red-btn"
                    >
                      {this.props.msg.pkg.get("browser.share.del")}
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={this.addSharing}
                      className="cyan-btn"
                    >
                      {this.props.msg.pkg.get("browser.share.add")}
                    </button>
                  )}
                </span>,

                <span>
                  {this.state.selectedItems.size > 0 ? (
                    <span>
                      <button
                        type="button"
                        onClick={() => this.delete()}
                        className="red-btn"
                      >
                        {this.props.msg.pkg.get("browser.delete")}
                      </button>

                      <button type="button" onClick={() => this.moveHere()}>
                        {this.props.msg.pkg.get("browser.paste")}
                      </button>
                    </span>
                  ) : null}
                </span>,

                <span>
                  <span
                    id="space-used"
                    className="desc-m grey0-font"
                  >{`${this.props.msg.pkg.get(
                    "browser.used"
                  )} ${usedSpace} / ${spaceLimit}`}</span>
                </span>,
              ])}
              childrenStyles={List([
                { flex: "0 0 auto" },
                { flex: "0 0 auto" },
                { justifyContent: "flex-end" },
              ])}
            />
          </div>

          <Flexbox
            children={List([
              <span id="breadcrumb">
                <Flexbox
                  children={List([
                    <RiHomeSmileFill size="3rem" id="icon-home" />,
                    <Flexbox children={breadcrumb} />,
                  ])}
                  childrenStyles={List([
                    { flex: "0 0 auto" },
                    { flex: "0 0 auto" },
                  ])}
                />
              </span>,

              <span className={`${showOp}`}>
                <button onClick={() => this.selectAll()} className="select-btn">
                  {this.props.msg.pkg.get("browser.selectAll")}
                </button>
              </span>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          {itemList}
        </Container>
      </div>
    );

    return <div id="browser">{itemListPane}</div>;
  }
}
