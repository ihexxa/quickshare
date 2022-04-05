import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiFile2Fill } from "@react-icons/all-files/ri/RiFile2Fill";
import { RiCheckboxFill } from "@react-icons/all-files/ri/RiCheckboxFill";
import { RiMenuUnfoldFill } from "@react-icons/all-files/ri/RiMenuUnfoldFill";
import { RiRestartFill } from "@react-icons/all-files/ri/RiRestartFill";
import { RiCheckboxBlankLine } from "@react-icons/all-files/ri/RiCheckboxBlankLine";

import { ErrorLogger } from "../common/log_error";
import { Env } from "../common/env";
import { getErrMsg } from "../common/utils";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { MetadataResp, roleVisitor, roleAdmin } from "../client";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { BtnList } from "./control/btn_list";
import { Segments } from "./layout/segments";
import { Columns } from "./layout/columns";
import { Up } from "../worker/upload_mgr";
import { UploadEntry, UploadState } from "../worker/interface";
import {
  ctrlOff,
  ctrlOn,
  sharingCtrl,
  filesViewCtrl,
  loadingCtrl,
} from "../common/controls";
import { HotkeyHandler } from "../common/hotkeys";
import { CronTable } from "../common/cron";
import { Title } from "./visual/title";
import { NotFoundBanner } from "./visual/banner_notfound";

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
  orderBy: string;
  order: boolean;
}

export interface Props {
  filesInfo: FilesProps;
  msg: MsgProps;
  login: LoginProps;
  ui: UIProps;
  enabled: boolean;
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
  private hotkeyHandler: HotkeyHandler;

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
      if (!this.props.enabled) {
        return;
      }
      const uploadInput = this.uploadInput as HTMLButtonElement;
      uploadInput.click();
    };
  }

  componentDidMount(): void {
    CronTable().setInterval("refreshFileList", {
      func: updater().refreshFiles,
      args: [],
      delay: 5000,
    });

    this.hotkeyHandler = new HotkeyHandler();
    this.hotkeyHandler.add({ key: "a", ctrl: true }, this.selectAll);
    this.hotkeyHandler.add({ key: "q", ctrl: true }, this.onClickUpload);
    this.hotkeyHandler.add({ key: "Delete" }, this.delete);
    this.hotkeyHandler.add({ key: "v", ctrl: true }, this.moveHere);

    document.addEventListener("keyup", this.hotkeyHandler.handle);
  }

  componentWillUnmount() {
    CronTable().clearInterval("refreshFileList");

    document.removeEventListener("keyup", this.hotkeyHandler.handle);
  }

  onNewFolderNameChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newFolderName: ev.target.value });
  };

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  updateProgress = async (infos: Map<string, UploadEntry>) => {
    updater().setUploads(infos);
    let errCount = 0;
    infos.valueSeq().forEach((entry: UploadEntry) => {
      errCount += entry.state === UploadState.Error ? 1 : 0;
    });

    if (infos.size === 0 || infos.size === errCount) {
      // refresh used space
      const status = await updater().self();
      if (status !== "") {
        Env().alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        return;
      }
      this.props.update(updater().updateLogin);
      this.props.update(updater().updateUploadingsInfo);
    }
  };

  addUploads = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files.length > 200) {
      Env().alertMsg(this.props.msg.pkg.get("err.tooManyUploads"));
      return;
    }

    let fileList = List<File>();
    for (let i = 0; i < event.target.files.length; i++) {
      fileList = fileList.push(event.target.files[i]);
    }

    const status = updater().addUploads(fileList);
    if (status !== "") {
      Env().alertMsg(getErrMsg(this.props.msg.pkg, "upload.add.fail", status));
    }
    this.props.update(updater().updateUploadingsInfo);
  };

  mkDir = async () => {
    if (this.state.newFolderName === "") {
      Env().alertMsg(this.props.msg.pkg.get("browser.folder.add.fail"));
      return;
    }

    const dirPath = getItemPath(
      this.props.filesInfo.dirPath.join("/"),
      this.state.newFolderName
    );

    this.setLoading(true);

    try {
      const mkDirStatus = await updater().mkDir(dirPath);
      if (mkDirStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", mkDirStatus.toString())
        );
        return;
      }

      const setItemsStatus = await updater().setItems(
        this.props.filesInfo.dirPath
      );
      if (setItemsStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", setItemsStatus.toString())
        );
        return;
      }

      this.setState({ newFolderName: "" });
      this.props.update(updater().updateFilesInfo);
      this.props.update(updater().updateSharingsInfo);
    } finally {
      this.setLoading(false);
    }
  };

  delete = async () => {
    if (!this.props.enabled) {
      return;
    }

    // TODO: selected should be cleaned after change the cwd
    if (this.props.filesInfo.dirPath.join("/") !== this.state.selectedSrc) {
      Env().alertMsg(this.props.msg.pkg.get("browser.del.fail"));
      this.setState({
        selectedSrc: this.props.filesInfo.dirPath.join("/"),
        selectedItems: Map<string, boolean>(),
      });
      return;
    }
    const filesToDel = this.state.selectedItems.keySeq().join(", ");
    if (
      !Env().confirmMsg(
        `${this.props.msg.pkg.get("op.confirm")} [${
          this.state.selectedItems.size
        }]: ${filesToDel}`
      )
    ) {
      return;
    }

    this.setLoading(true);

    try {
      const cwd = this.props.filesInfo.dirPath.join("/");
      const itemsToDel = this.props.filesInfo.items
        .filter((item) => {
          return this.state.selectedItems.has(item.name);
        })
        .map((selectedItem: MetadataResp): string => {
          return getItemPath(cwd, selectedItem.name);
        })
        .toArray();

      const deleteStatus = await updater().deleteInArray(itemsToDel);
      if (deleteStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", deleteStatus.toString())
        );
        return deleteStatus;
      }

      const refreshStatus = await updater().setItems(
        this.props.filesInfo.dirPath
      );
      if (refreshStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", refreshStatus.toString())
        );
        return refreshStatus;
      }

      const selfStatus = await updater().self();
      if (selfStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", selfStatus.toString())
        );
        return selfStatus;
      }

      this.props.update(updater().updateFilesInfo);
      this.props.update(updater().updateSharingsInfo);
      this.props.update(updater().updateLogin);
      this.setState({
        selectedSrc: "",
        selectedItems: Map<string, boolean>(),
      });
    } finally {
      this.setLoading(false);
    }
  };

  moveHere = async () => {
    if (!this.props.enabled) {
      return;
    }

    const oldDir = this.state.selectedSrc;
    const newDir = this.props.filesInfo.dirPath.join("/");
    if (oldDir === newDir) {
      Env().alertMsg(this.props.msg.pkg.get("browser.move.fail"));
      return;
    }

    this.setLoading(true);

    try {
      const moveStatus = await updater().moveHere(
        this.state.selectedSrc,
        this.props.filesInfo.dirPath.join("/"),
        this.state.selectedItems
      );
      if (moveStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", moveStatus.toString())
        );
        return;
      }

      this.props.update(updater().updateFilesInfo);
      this.props.update(updater().updateSharingsInfo);
      this.setState({
        selectedSrc: "",
        selectedItems: Map<string, boolean>(),
      });
    } finally {
      this.setLoading(false);
    }
  };

  gotoChild = async (childDirName: string) => {
    return this.chdir(this.props.filesInfo.dirPath.push(childDirName));
  };

  goHome = async () => {
    this.setLoading(true);

    try {
      const status = await updater().setHomeItems();
      if (status !== "") {
        Env().alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        return;
      }
      this.props.update(updater().updateFilesInfo);
    } finally {
      this.setLoading(false);
    }
  };

  chdir = async (dirPath: List<string>) => {
    if (dirPath === this.props.filesInfo.dirPath) {
      return;
    } else if (this.props.login.userRole !== roleAdmin && dirPath.size <= 1) {
      Env().alertMsg(this.props.msg.pkg.get("unauthed"));
      return;
    }

    this.setLoading(true);
    try {
      const status = await updater().setItems(dirPath);
      if (status !== "") {
        Env().alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        return;
      }

      const isSharingStatus = await updater().syncIsSharing(dirPath.join("/"));
      if (isSharingStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", isSharingStatus)
        );
        return;
      }

      this.props.update(updater().updateFilesInfo);
      this.props.update(updater().updateSharingsInfo);
    } finally {
      this.setLoading(false);
    }
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
    if (!this.props.enabled) {
      return;
    }

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

  generateHash = async (filePath: string) => {
    Env().alertMsg(this.props.msg.pkg.get("refresh-hint"));
    updater().generateHash(filePath);
  };

  addSharing = async () => {
    this.setLoading(true);

    try {
      const addStatus = await updater().addSharing();
      if (addStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", addStatus.toString())
        );
        return;
      }

      updater().setSharing(true);
      const listStatus = await updater().listSharings();
      if (listStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", listStatus.toString())
        );
        return;
      }

      this.props.update(updater().updateSharingsInfo);
      this.props.update(updater().updateFilesInfo);
    } finally {
      this.setLoading(false);
    }
  };

  deleteSharing = async (dirPath: string) => {
    this.setLoading(true);

    try {
      const delStatus = await updater().deleteSharing(dirPath);
      if (delStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", delStatus.toString())
        );
        return;
      }

      updater().setSharing(false);
      const listStatus = await updater().listSharings();
      if (listStatus !== "") {
        Env().alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", listStatus.toString())
        );
        return;
      }

      this.props.update(updater().updateSharingsInfo);
      this.props.update(updater().updateFilesInfo);
    } finally {
      this.setLoading(false);
    }
  };

  updateItems = (items: Object) => {
    const metadataResps = items as List<MetadataResp>;
    updater().updateItems(metadataResps);
    this.props.update(updater().updateFilesInfo);
  };

  prepareColumns = (
    sortedItems: List<MetadataResp>,
    showOp: string
  ): React.ReactNode => {
    const shareModeClass =
      this.props.ui.control.controls.get(sharingCtrl) === ctrlOn
        ? "hidden"
        : "";

    const rows = sortedItems.map((item: MetadataResp): React.ReactNode => {
      const isSelected = this.state.selectedItems.has(item.name);
      const dirPath = this.props.filesInfo.dirPath.join("/");
      const itemPath = dirPath.endsWith("/")
        ? `${dirPath}${item.name}`
        : `${dirPath}/${item.name}`;

      const selectedIconColor = isSelected ? "focus-font" : "minor-font";
      const descIconColor = this.state.showDetail.has(item.name)
        ? "focus-font"
        : "major-font";
      const icon = item.isDir ? (
        <RiFolder2Fill size="2rem" className="yellow0-font margin-r-m" />
      ) : (
        <RiFile2Fill size="2rem" className="focus-font margin-r-m" />
      );

      const modTimeDate = new Date(item.modTime);
      const modTimeFormatted = `${modTimeDate.getFullYear()}-${
        modTimeDate.getMonth() + 1
      }-${modTimeDate.getDate()}`;
      const downloadPath = `/v1/fs/files?fp=${itemPath}`;
      const name = item.isDir ? (
        <span className="title-m-wrap">
          <span className="clickable" onClick={() => this.gotoChild(item.name)}>
            {item.name}
          </span>
          <span className="major-font">{` - ${modTimeFormatted}`}</span>
        </span>
      ) : (
        <span className="title-m-wrap">
          <a className="clickable" href={downloadPath} target="_blank">
            {item.name}
          </a>
          <span className="major-font">{` - ${modTimeFormatted}`}</span>
        </span>
      );

      const checkIcon = isSelected ? (
        <RiCheckboxFill
          size="2rem"
          className={`${selectedIconColor} ${shareModeClass}`}
          onClick={() => this.select(item.name)}
        />
      ) : (
        <RiCheckboxBlankLine
          size="2rem"
          className={`${selectedIconColor} ${shareModeClass}`}
          onClick={() => this.select(item.name)}
        />
      );

      const op = item.isDir ? (
        <div className={`txt-align-r icon-s ${showOp}`}>{checkIcon}</div>
      ) : (
        <div className={`txt-align-r icon-s ${showOp}`}>
          <RiMenuUnfoldFill
            size="2rem"
            className={`${descIconColor} margin-r-m`}
            onClick={() => this.toggleDetail(item.name)}
          />
          {checkIcon}
        </div>
      );

      const absDownloadURL = `${document.location.protocol}//${document.location.hostname}:${document.location.port}${downloadPath}`;
      const pathTitle = this.props.msg.pkg.get("item.downloadURL");
      const modTimeTitle = this.props.msg.pkg.get("item.modTime");
      const sizeTitle = this.props.msg.pkg.get("item.size");
      const itemSize = FileSize(item.size, { round: 0 });

      const descStateClass = this.state.showDetail.has(item.name)
        ? "margin-t-m padding-m"
        : "no-height";
      const desc = (
        <div className={`${descStateClass} major-font major-bg`}>
          <div className="column">
            <div className="card">
              <span className="title-m minor-font">{pathTitle}</span>
              <span className="font-s work-break-all">{absDownloadURL}</span>
            </div>
          </div>

          <div className="column">
            <div className="card">
              <span className="title-m minor-font">{modTimeTitle}</span>
              <span className="font-s work-break-all">{modTimeFormatted}</span>
            </div>
            <div className="card">
              <span className="title-m minor-font">{sizeTitle}</span>
              <span className="font-s work-break-all">{itemSize}</span>
            </div>
          </div>

          <div className="fix">
            <div className="card">
              <Flexbox
                children={List([
                  <span className="title-m minor-font">SHA1</span>,
                  <RiRestartFill
                    onClick={() => this.generateHash(itemPath)}
                    size={"2rem"}
                    className={`minor-font ${shareModeClass}`}
                  />,
                ])}
                childrenStyles={List([{}, { justifyContent: "flex-end" }])}
              />
              <div className="info minor-bg">{item.sha1}</div>
            </div>
          </div>
        </div>
      );

      const cells = List<React.ReactNode>([
        <div className="icon-s">{icon}</div>,
        <div>{name}</div>,
        <div className="title-m major-font padding-l-s">{itemSize}</div>,
        <div className="txt-align-r">{op}</div>,
      ]);

      const tableCols = (
        <Columns
          rows={List([cells])}
          widths={List(["3rem", "calc(100% - 18rem)", "8rem", "7rem"])}
          childrenClassNames={List(["", "", "", ""])}
        />
      );

      return (
        <div>
          {tableCols}
          {desc}
          <div className="hr"></div>
        </div>
      );
    });

    return <div>{rows}</div>;
  };

  setView = (opt: string) => {
    if (opt !== "rows" && opt !== "table") {
      ErrorLogger().error(`FilesPanel:setView: unknown view ${opt}`);
      return;
    }
    updater().setControlOption(filesViewCtrl, opt);
    this.props.update(updater().updateUI);
  };

  orderBy = (columnName: string) => {
    const order =
      this.props.filesInfo.orderBy === columnName
        ? !this.props.filesInfo.order
        : true;

    updater().sortFiles(columnName, order);
    this.props.update(updater().updateFilesInfo);
  };

  render() {
    const showEndpoints =
      this.props.login.userRole === roleAdmin ? "" : "hidden";
    const gotoRoot = () => this.chdir(List(["/"]));
    const endPoints = (
      <div className={showEndpoints}>
        <Container>
          <Title
            title={this.props.msg.pkg.get("endpoints")}
            iconColor="normal"
            iconName="RiGridFill"
          />
          <div className="hr"></div>

          <button onClick={gotoRoot} className="button-default margin-r-m">
            {this.props.msg.pkg.get("endpoints.root")}
          </button>
          <button onClick={this.goHome} className="button-default">
            {this.props.msg.pkg.get("endpoints.home")}
          </button>
        </Container>
      </div>
    );
    const shareModeClass =
      this.props.ui.control.controls.get(sharingCtrl) === ctrlOn
        ? "hidden"
        : "";

    const showOp = this.props.login.userRole === roleVisitor ? "hidden" : "";
    const breadcrumb = this.props.filesInfo.dirPath.map(
      (pathPart: string, key: number) => {
        return (
          <span key={`${pathPart}-${key}`}>
            <a
              onClick={() =>
                this.chdir(this.props.filesInfo.dirPath.slice(0, key + 1))
              }
              className="item clickable"
            >
              <span className="content">
                {pathPart === "/" ? "~" : pathPart}
              </span>
            </a>
            <span className="item">
              <span className="content">{"/"}</span>
            </span>
          </span>
        );
      }
    );

    const ops = (
      <div>
        <Flexbox
          children={List([
            <div>
              <button
                onClick={this.mkDir}
                className="inline-block focus-bg white-font margin-r-m"
              >
                {this.props.msg.pkg.get("browser.folder.add")}
              </button>
              <input
                type="text"
                onChange={this.onNewFolderNameChange}
                value={this.state.newFolderName}
                placeholder={this.props.msg.pkg.get("browser.folder.name")}
                className="inline-block"
              />
            </div>,

            <div>
              <button
                onClick={this.onClickUpload}
                className="focus-bg white-font"
              >
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
            </div>,
          ])}
          childrenStyles={List([
            { flex: "0 0 70%" },
            { flex: "0 0 30%", justifyContent: "flex-end" },
          ])}
        />
      </div>
    );

    const orderByCallbacks = List([
      () => {
        this.orderBy(this.props.msg.pkg.get("item.name"));
      },
      () => {
        this.orderBy(this.props.msg.pkg.get("item.type"));
      },
      () => {
        this.orderBy(this.props.msg.pkg.get("item.modTime"));
      },
    ]);
    const orderByButtons = (
      <BtnList
        titleIcon="BiSortUp"
        btnNames={List([
          this.props.msg.pkg.get("item.name"),
          this.props.msg.pkg.get("item.type"),
          this.props.msg.pkg.get("item.modTime"),
        ])}
        btnCallbacks={orderByCallbacks}
      />
    );
    const viewType = this.props.ui.control.controls.get(filesViewCtrl);
    const view =
      this.props.filesInfo.items.size > 0 ? (
        <div>
          {orderByButtons}
          <div className="margin-t-l">
            {this.prepareColumns(this.props.filesInfo.items, showOp)}
          </div>
        </div>
      ) : (
        <NotFoundBanner title={this.props.msg.pkg.get("terms.nothingHere")} />
      ); // TODO: support better views in the future

    const usedSpace = FileSize(
      // TODO: this a work around before transaction is introduced
      Math.trunc(
        parseInt(this.props.login.extInfo.usedSpace, 10) / (1024 * 1024)
      ) *
        (1024 * 1024),
      {
        round: 0,
      }
    );
    const spaceLimit = FileSize(
      // TODO: this a work around before transaction is introduced
      Math.trunc(
        parseInt(this.props.login.quota.spaceLimit, 10) / (1024 * 1024)
      ) *
        (1024 * 1024),
      {
        round: 0,
      }
    );

    const rowsViewColorClass =
      this.props.ui.control.controls.get(filesViewCtrl) === "rows"
        ? "focus-font"
        : "major-font";
    const tableViewColorClass =
      this.props.ui.control.controls.get(filesViewCtrl) === "table"
        ? "focus-font"
        : "major-font";

    const itemListPane = (
      <div>
        {endPoints}

        <div className={showOp}>
          <Container>{ops}</Container>
        </div>

        <Container>
          <div className={`${showOp} margin-b-m`}>
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
                      className="red0-bg white-font margin-r-m"
                    >
                      {this.props.msg.pkg.get("browser.share.del")}
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={this.addSharing}
                      className="focus-bg white-font margin-r-m"
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
                        className="red0-bg white-font margin-r-m"
                      >
                        {this.props.msg.pkg.get("browser.delete")}
                      </button>

                      <button
                        type="button"
                        className="button-default"
                        onClick={() => this.moveHere()}
                      >
                        {this.props.msg.pkg.get("browser.paste")}
                      </button>
                    </span>
                  ) : null}
                </span>,

                <Flexbox
                  children={List([
                    // <BiListUl
                    //   size="2rem"
                    //   className={`${rowsViewColorClass} margin-r-s`}
                    //   onClick={() => {
                    //     this.setView("rows");
                    //   }}
                    // />,
                    // <BiTable
                    //   size="2rem"
                    //   className={`${tableViewColorClass} margin-r-s`}
                    //   onClick={() => {
                    //     this.setView("table");
                    //   }}
                    // />,

                    <span className={`${showOp}`}>
                      <button
                        onClick={() => this.selectAll()}
                        className="inline-block button-default"
                      >
                        {this.props.msg.pkg.get("browser.selectAll")}
                      </button>
                    </span>,
                  ])}
                />,
              ])}
              childrenStyles={List([
                { flex: "0 0 auto" },
                { flex: "0 0 auto" },
                { justifyContent: "flex-end" },
              ])}
            />
          </div>

          <Segments
            id="breadcrumb"
            children={List([
              <div>
                <span className="location-item">
                  <span className="content">
                    {`${this.props.msg.pkg.get("breadcrumb.loc")}:`}
                  </span>
                </span>
                {breadcrumb}
              </div>,
              <div
                id="space-used"
                className={`grey0-font ${shareModeClass}`}
              >{`${this.props.msg.pkg.get(
                "browser.used"
              )} ${usedSpace} / ${spaceLimit}`}</div>,
            ])}
            ratios={List([60, 40])}
            dir={true}
          />

          <div className="hr grey0-bg"></div>

          {view}
        </Container>
      </div>
    );

    return <div id="browser">{itemListPane}</div>;
  }
}
