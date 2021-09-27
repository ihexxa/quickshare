import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiHomeSmileFill } from "@react-icons/all-files/ri/RiHomeSmileFill";
import { RiFile2Fill } from "@react-icons/all-files/ri/RiFile2Fill";
import { RiShareBoxLine } from "@react-icons/all-files/ri/RiShareBoxLine";
import { RiFolderSharedFill } from "@react-icons/all-files/ri/RiFolderSharedFill";
import { RiUploadCloudFill } from "@react-icons/all-files/ri/RiUploadCloudFill";
import { RiUploadCloudLine } from "@react-icons/all-files/ri/RiUploadCloudLine";
import { RiEmotionSadLine } from "@react-icons/all-files/ri/RiEmotionSadLine";

import { alertMsg, confirmMsg } from "../common/env";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { MetadataResp, roleVisitor, roleAdmin } from "../client";
import { Up } from "../worker/upload_mgr";
import { UploadEntry } from "../worker/interface";
import { Flexbox } from "./layout/flexbox";

export interface Item {
  name: string;
  size: number;
  modTime: string;
  isDir: boolean;
  selected: boolean;
  sha1: string;
}

export interface BrowserProps {
  dirPath: List<string>;
  isSharing: boolean;
  items: List<MetadataResp>;
  uploadings: List<UploadEntry>;
  sharings: List<string>;

  uploadFiles: List<File>;
  uploadValue: string;

  tab: string;
}

export interface Props {
  browser: BrowserProps;
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
  inputValue: string;
  selectedSrc: string;
  selectedItems: Map<string, boolean>;
  showDetail: Set<string>;
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
      showDetail: Set<string>(),
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
    if (event.target.files.length > 1000) {
      alertMsg(this.props.msg.pkg.get("err.tooManyUploads"));
      return;
    }

    let fileList = List<File>();
    for (let i = 0; i < event.target.files.length; i++) {
      fileList = fileList.push(event.target.files[i]);
    }
    updater().addUploads(fileList);
    this.update(updater().updateBrowser);
  };

  deleteUpload = (filePath: string): Promise<void> => {
    return updater()
      .deleteUpload(filePath)
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.upload.del.fail"));
        }
        return updater().refreshUploadings();
      })
      .then(() => {
        return updater().self();
      })
      .then(() => {
        this.update(updater().updateBrowser);
        this.update(updater().updateLogin);
      });
  };

  stopUploading = (filePath: string) => {
    updater().stopUploading(filePath);
    this.update(updater().updateBrowser);
  };

  onMkDir = () => {
    if (this.state.inputValue === "") {
      alertMsg(this.props.msg.pkg.get("browser.folder.add.fail"));
      return;
    }

    const dirPath = getItemPath(
      this.props.browser.dirPath.join("/"),
      this.state.inputValue
    );
    updater()
      .mkDir(dirPath)
      .then(() => {
        this.setState({ inputValue: "" });
        return updater().setItems(this.props.browser.dirPath);
      })
      .then(() => {
        this.update(updater().updateBrowser);
      });
  };

  delete = () => {
    if (this.props.browser.dirPath.join("/") !== this.state.selectedSrc) {
      alertMsg(this.props.msg.pkg.get("browser.del.fail"));
      this.setState({
        selectedSrc: this.props.browser.dirPath.join("/"),
        selectedItems: Map<string, boolean>(),
      });
      return;
    } else {
      const filesToDel = this.state.selectedItems.keySeq().join(", ");
      if (!confirmMsg(`do you want to delete ${filesToDel}?`)) {
        return;
      }
    }

    updater()
      .delete(
        this.props.browser.dirPath,
        this.props.browser.items,
        this.state.selectedItems
      )
      .then(() => {
        return updater().self();
      })
      .then(() => {
        this.update(updater().updateBrowser);
        this.update(updater().updateLogin);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  moveHere = () => {
    const oldDir = this.state.selectedSrc;
    const newDir = this.props.browser.dirPath.join("/");
    if (oldDir === newDir) {
      alertMsg(this.props.msg.pkg.get("browser.move.fail"));
      return;
    }

    updater()
      .moveHere(
        this.state.selectedSrc,
        this.props.browser.dirPath.join("/"),
        this.state.selectedItems
      )
      .then(() => {
        this.update(updater().updateBrowser);
        this.setState({
          selectedSrc: "",
          selectedItems: Map<string, boolean>(),
        });
      });
  };

  gotoChild = (childDirName: string) => {
    this.chdir(this.props.browser.dirPath.push(childDirName));
  };

  chdir = async (dirPath: List<string>) => {
    if (dirPath === this.props.browser.dirPath) {
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
        this.update(updater().updateBrowser);
      });
  };

  updateProgress = (infos: Map<string, UploadEntry>, refresh: boolean) => {
    updater().setUploadings(infos);
    if (refresh) {
      updater()
        .setItems(this.props.browser.dirPath)
        .then((ok: boolean) => {
          this.update(updater().updateBrowser);
        });
    } else {
      this.update(updater().updateBrowser);
    }
  };

  select = (itemName: string) => {
    const selectedItems = this.state.selectedItems.has(itemName)
      ? this.state.selectedItems.delete(itemName)
      : this.state.selectedItems.set(itemName, true);

    this.setState({
      selectedSrc: this.props.browser.dirPath.join("/"),
      selectedItems: selectedItems,
    });
  };

  selectAll = () => {
    let newSelected = Map<string, boolean>();
    const someSelected = this.state.selectedItems.size === 0 ? true : false;
    if (someSelected) {
      this.props.browser.items.forEach((item) => {
        newSelected = newSelected.set(item.name, true);
      });
    } else {
      this.props.browser.items.forEach((item) => {
        newSelected = newSelected.delete(item.name);
      });
    }

    this.setState({
      selectedSrc: this.props.browser.dirPath.join("/"),
      selectedItems: newSelected,
    });
  };

  addSharing = async () => {
    return updater()
      .addSharing()
      .then((ok) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.share.add.fail"));
        } else {
          updater().setSharing(true);
          return this.listSharings();
        }
      })
      .then(() => {
        this.props.update(updater().updateBrowser);
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
          return this.listSharings();
        }
      })
      .then(() => {
        this.props.update(updater().updateBrowser);
      });
  };

  listSharings = async () => {
    return updater()
      .listSharings()
      .then((ok) => {
        if (ok) {
          this.update(updater().updateBrowser);
        }
      });
  };

  setTab = (tabName: string) => {
    updater().setTab(tabName);
    this.props.update(updater().updateBrowser);
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

  render() {
    const showOp = this.props.login.userRole === roleVisitor ? "hidden" : "";

    let breadcrumb = this.props.browser.dirPath.map(
      (pathPart: string, key: number) => {
        return (
          <button
            key={pathPart}
            onClick={() =>
              this.chdir(this.props.browser.dirPath.slice(0, key + 1))
            }
            className="grey3-bg grey4-font margin-r-m"
          >
            {pathPart}
          </button>
        );
      }
    );

    const nameCellClass = `item-name item-name-${
      this.props.ui.isVertical ? "vertical" : "horizontal"
    } pointer`;

    const ops = (
      <div>
        <div>
          <span className="inline-block margin-t-m margin-b-m">
            <input
              type="text"
              onChange={this.onInputChange}
              value={this.state.inputValue}
              className="black0-font margin-r-m"
              placeholder={this.props.msg.pkg.get("browser.folder.name")}
            />
            <button onClick={this.onMkDir} className="margin-r-m">
              {this.props.msg.pkg.get("browser.folder.add")}
            </button>
          </span>
          <span className="inline-block margin-t-m margin-b-m">
            <button onClick={this.onClickUpload}>
              {this.props.msg.pkg.get("browser.upload")}
            </button>
            <input
              type="file"
              onChange={this.addUploads}
              multiple={true}
              value={this.props.browser.uploadValue}
              ref={this.assignInput}
              className="black0-font hidden"
            />
          </span>
        </div>
      </div>
    );

    const sortedItems = this.props.browser.items.sort(
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
      const dirPath = this.props.browser.dirPath.join("/");
      const itemPath = dirPath.endsWith("/")
        ? `${dirPath}${item.name}`
        : `${dirPath}/${item.name}`;

      return item.isDir ? (
        <Flexbox
          key={item.name}
          children={List([
            <span className="padding-m">
              <Flexbox
                children={List([
                  <RiFolder2Fill
                    size="3rem"
                    className="yellow0-font margin-r-m"
                  />,

                  <span className="">
                    <span className={nameCellClass}>
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
                    </span>
                  </span>,
                ])}
              />
            </span>,

            <span className={`padding-m ${showOp}`}>
              <button
                onClick={() => this.select(item.name)}
                className={`${
                  isSelected ? "blue0-bg white-font" : "grey2-bg grey3-font"
                }`}
                style={{ width: "8rem", display: "inline-block" }}
              >
                {isSelected
                  ? this.props.msg.pkg.get("browser.deselect")
                  : this.props.msg.pkg.get("browser.select")}
              </button>
            </span>,
          ])}
          childrenStyles={List([{}, { justifyContent: "flex-end" }])}
        />
      ) : (
        <div key={item.name}>
          <Flexbox
            children={List([
              <span className="padding-m">
                <Flexbox
                  children={List([
                    <RiFile2Fill
                      size="3rem"
                      className="cyan0-font margin-r-m"
                    />,

                    <span className={`${nameCellClass}`}>
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
                />
              </span>,

              <span className={`item-op padding-m ${showOp}`}>
                <button
                  onClick={() => this.toggleDetail(item.name)}
                  className="grey2-bg grey3-font margin-r-m"
                  style={{ width: "8rem", display: "inline-block" }}
                >
                  {this.props.msg.pkg.get("detail")}
                </button>

                <button
                  type="button"
                  onClick={() => this.select(item.name)}
                  className={`${
                    isSelected ? "blue0-bg white-font" : "grey2-bg grey3-font"
                  }`}
                  style={{ width: "8rem", display: "inline-block" }}
                >
                  {isSelected
                    ? this.props.msg.pkg.get("browser.deselect")
                    : this.props.msg.pkg.get("browser.select")}
                </button>
              </span>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />
          <div
            className={`${
              this.state.showDetail.has(item.name) ? "" : "hidden"
            }`}
          >
            <Flexbox
              children={List([
                <span>
                  <b>SHA1:</b>
                  {` ${item.sha1}`}
                </span>,
                <button
                  onClick={() => this.generateHash(itemPath)}
                  className="black-bg white-font margin-l-m"
                  style={{ display: "inline-block" }}
                >
                  {this.props.msg.pkg.get("refresh")}
                </button>,
              ])}
              className={`grey2-bg grey3-font detail margin-r-m`}
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

    const itemListPane =
      this.props.browser.tab === "" || this.props.browser.tab === "item" ? (
        <div>
          <div id="op-bar" className={`op-bar ${showOp}`}>
            <div className="margin-l-m margin-r-m margin-b-m">{ops}</div>
          </div>

          <div className="container" style={{ paddingBottom: "1rem" }}>
            <div className={`${showOp}`}>
              <Flexbox
                children={List([
                  <span>
                    {this.props.browser.isSharing ? (
                      <button
                        type="button"
                        onClick={() => {
                          this.deleteSharing(
                            this.props.browser.dirPath.join("/")
                          );
                        }}
                        className="red0-bg white-font margin-r-m"
                      >
                        {this.props.msg.pkg.get("browser.share.del")}
                      </button>
                    ) : (
                      <button
                        type="button"
                        onClick={this.addSharing}
                        className="green0-bg white-font margin-r-m"
                      >
                        {this.props.msg.pkg.get("browser.share.add")}
                      </button>
                    )}

                    {this.state.selectedItems.size > 0 ? (
                      <span>
                        <button
                          type="button"
                          onClick={() => this.delete()}
                          className="red0-bg white-font margin-r-m"
                        >
                          {this.props.msg.pkg.get("browser.delete")}
                        </button>

                        <button
                          type="button"
                          onClick={() => this.moveHere()}
                          className="margin-r-m"
                        >
                          {this.props.msg.pkg.get("browser.paste")}
                        </button>
                      </span>
                    ) : null}
                  </span>,

                  <span>
                    <span className="desc-m grey0-font">{`${this.props.msg.pkg.get(
                      "browser.used"
                    )} ${usedSpace} / ${spaceLimit}`}</span>
                  </span>,
                ])}
                className="padding-m"
                childrenStyles={List([{}, { justifyContent: "flex-end" }])}
              />
            </div>

            <Flexbox
              children={List([
                <span className="padding-m">
                  <Flexbox
                    children={List([
                      <RiHomeSmileFill
                        size="3rem"
                        className="margin-r-m black-font"
                      />,

                      <Flexbox children={breadcrumb} />,
                    ])}
                  />
                </span>,

                <span className={`padding-m ${showOp}`}>
                  <button
                    onClick={() => this.selectAll()}
                    className={`grey1-bg white-font`}
                    style={{ width: "8rem", display: "inline-block" }}
                  >
                    {this.props.msg.pkg.get("browser.selectAll")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />

            {itemList}
          </div>
        </div>
      ) : null;

    const uploadingList = this.props.browser.uploadings.map(
      (uploading: UploadEntry) => {
        const pathParts = uploading.filePath.split("/");
        const fileName = pathParts[pathParts.length - 1];

        return (
          <div key={uploading.filePath}>
            <Flexbox
              children={List([
                <span className="padding-m">
                  <Flexbox
                    children={List([
                      <RiUploadCloudLine
                        size="3rem"
                        className="margin-r-m blue0-font"
                      />,

                      <div className={`${nameCellClass}`}>
                        <span className="title-m">{fileName}</span>
                        <div className="desc-m grey0-font">
                          {FileSize(uploading.uploaded, { round: 0 })}
                          &nbsp;/&nbsp;{FileSize(uploading.size, { round: 0 })}
                        </div>
                      </div>,
                    ])}
                  />
                </span>,

                <div className="item-op padding-m">
                  <button
                    onClick={() => this.stopUploading(uploading.filePath)}
                    className="grey3-bg grey4-font margin-r-m"
                  >
                    {this.props.msg.pkg.get("browser.stop")}
                  </button>
                  <button
                    onClick={() => this.deleteUpload(uploading.filePath)}
                    className="grey3-bg grey4-font"
                  >
                    {this.props.msg.pkg.get("browser.delete")}
                  </button>
                </div>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            {uploading.err.trim() === "" ? null : (
              <div className="alert-red margin-s">
                <span className="padding-m">{uploading.err.trim()}</span>
              </div>
            )}
          </div>
        );
      }
    );

    const uploadingListPane =
      this.props.browser.tab === "uploading" ? (
        this.props.browser.uploadings.size === 0 ? (
          <div className="container">
            <Flexbox
              children={List([
                <RiEmotionSadLine
                  size="4rem"
                  className="margin-r-m red0-font"
                />,
                <span>
                  <h3 className="title-l">
                    {this.props.msg.pkg.get("upload.404.title")}
                  </h3>
                  <span className="desc-l grey0-font">
                    {this.props.msg.pkg.get("upload.404.desc")}
                  </span>
                </span>,
              ])}
              childrenStyles={List([
                { flex: "auto", justifyContent: "flex-end" },
                { flex: "auto" },
              ])}
              className="padding-l"
            />
          </div>
        ) : (
          <div className="container padding-b-m">
            <Flexbox
              children={List([
                <span className="padding-m">
                  <Flexbox
                    children={List([
                      <RiUploadCloudFill
                        size="3rem"
                        className="margin-r-m black-font"
                      />,

                      <span>
                        <span className="title-m bold">
                          {this.props.msg.pkg.get("browser.upload.title")}
                        </span>
                        <span className="desc-m grey0-font">
                          {this.props.msg.pkg.get("browser.upload.desc")}
                        </span>
                      </span>,
                    ])}
                  />
                </span>,

                <span></span>,
              ])}
            />

            {uploadingList}
          </div>
        )
      ) : null;

    const sharingList = this.props.browser.sharings.map((dirPath: string) => {
      return (
        <div key={dirPath} className="padding-m">
          <Flexbox
            children={List([
              <Flexbox
                children={List([
                  <RiFolderSharedFill
                    size="3rem"
                    className="purple0-font margin-r-m"
                  />,
                  <span>{dirPath}</span>,
                ])}
              />,

              <span>
                <input
                  type="text"
                  readOnly
                  className="margin-r-m"
                  value={`${
                    document.location.href.split("?")[0]
                  }?dir=${encodeURIComponent(dirPath)}`}
                />
                <button
                  onClick={() => {
                    this.deleteSharing(dirPath);
                  }}
                  className="grey3-bg grey4-bg"
                >
                  {this.props.msg.pkg.get("browser.share.del")}
                </button>
              </span>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />
        </div>
      );
    });

    const sharingListPane =
      this.props.browser.tab === "sharing" ? (
        this.props.browser.sharings.size === 0 ? (
          <div className="container">
            <Flexbox
              children={List([
                <RiEmotionSadLine
                  size="4rem"
                  className="margin-r-m red0-font"
                />,
                <span>
                  <h3 className="title-l">
                    {this.props.msg.pkg.get("share.404.title")}
                  </h3>
                  <span className="desc-l grey0-font">
                    {this.props.msg.pkg.get("share.404.desc")}
                  </span>
                </span>,
              ])}
              childrenStyles={List([
                { flex: "auto", justifyContent: "flex-end" },
                { flex: "auto" },
              ])}
              className="padding-l"
            />
          </div>
        ) : (
          <div className="container">
            <Flexbox
              children={List([
                <span className="padding-m">
                  <Flexbox
                    children={List([
                      <RiShareBoxLine
                        size="3rem"
                        className="margin-r-m black-font"
                      />,

                      <span>
                        <span className="title-m bold">
                          {this.props.msg.pkg.get("browser.share.title")}
                        </span>
                        <span className="desc-m grey0-font">
                          {this.props.msg.pkg.get("browser.share.desc")}
                        </span>
                      </span>,
                    ])}
                  />
                </span>,

                <span></span>,
              ])}
            />

            {sharingList}
          </div>
        )
      ) : null;

    const showTabs = this.props.login.userRole === roleVisitor ? "hidden" : "";
    return (
      <div>
        <div id="item-list">
          <div className={`container ${showTabs}`}>
            <div className="padding-m">
              <button
                onClick={() => {
                  this.setTab("item");
                }}
                className="margin-r-m"
              >
                <Flexbox
                  children={List([
                    <RiFolder2Fill
                      size="1.6rem"
                      className="margin-r-s cyan0-font"
                    />,
                    <span>{this.props.msg.pkg.get("browser.item.title")}</span>,
                  ])}
                  childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
                />
              </button>
              <button
                onClick={() => {
                  this.setTab("uploading");
                }}
                className="margin-r-m"
              >
                <Flexbox
                  children={List([
                    <RiUploadCloudFill
                      size="1.6rem"
                      className="margin-r-s blue0-font"
                    />,
                    <span>
                      {this.props.msg.pkg.get("browser.upload.title")}
                    </span>,
                  ])}
                  childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
                />
              </button>
              <button
                onClick={() => {
                  this.setTab("sharing");
                }}
                className="margin-r-m"
              >
                <Flexbox
                  children={List([
                    <RiShareBoxLine
                      size="1.6rem"
                      className="margin-r-s purple0-font"
                    />,
                    <span>
                      {this.props.msg.pkg.get("browser.share.title")}
                    </span>,
                  ])}
                  childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
                />
              </button>
            </div>
          </div>
          <div>{sharingListPane}</div>
          <div>{uploadingListPane}</div>
          {itemListPane}
        </div>
      </div>
    );
  }
}
