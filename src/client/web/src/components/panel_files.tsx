import * as React from "react";
import * as ReactDOM from "react-dom";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiFile2Fill } from "@react-icons/all-files/ri/RiFile2Fill";
import { RiFileList2Fill } from "@react-icons/all-files/ri/RiFileList2Fill";
import { RiCheckboxFill } from "@react-icons/all-files/ri/RiCheckboxFill";
import { RiMore2Fill } from "@react-icons/all-files/ri/RiMore2Fill";
import { BiTable } from "@react-icons/all-files/bi/BiTable";
import { BiListUl } from "@react-icons/all-files/bi/BiListUl";
import { RiRestartFill } from "@react-icons/all-files/ri/RiRestartFill";
import { RiCheckboxBlankLine } from "@react-icons/all-files/ri/RiCheckboxBlankLine";

import { ErrorLogger } from "../common/log_error";
import { alertMsg, confirmMsg } from "../common/env";
import { getErrMsg } from "../common/utils";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { MetadataResp, roleVisitor, roleAdmin } from "../client";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { Table, Cell, Head } from "./layout/table";
import { BtnList } from "./control/btn_list";
import { Segments } from "./layout/segments";
import { Rows } from "./layout/rows";
import { Up } from "../worker/upload_mgr";
import { UploadEntry, UploadState } from "../worker/interface";
import { getIcon } from "./visual/icons";
import {
  ctrlOff,
  ctrlOn,
  sharingCtrl,
  filesViewCtrl,
  loadingCtrl,
} from "../common/controls";
import { HotkeyHandler } from "../common/hotkeys";
import { CronTable } from "../common/cron";

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

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  updateProgress = async (
    infos: Map<string, UploadEntry>,
    refresh: boolean
  ) => {
    updater().setUploads(infos);
    let errCount = 0;
    infos.valueSeq().forEach((entry: UploadEntry) => {
      errCount += entry.state === UploadState.Error ? 1 : 0;
    });

    if (infos.size === 0 || infos.size === errCount) {
      // refresh used space
      const status = await updater().self();
      if (status !== "") {
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        return;
      }
      this.props.update(updater().updateLogin);
      this.props.update(updater().updateUploadingsInfo);
    }

    if (refresh) {
      const status = await updater().setItems(this.props.filesInfo.dirPath);
      if (status !== "") {
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        return;
      }
    }
    this.props.update(updater().updateFilesInfo);
    this.props.update(updater().updateUploadingsInfo);
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

    const status = updater().addUploads(fileList);
    if (status !== "") {
      alertMsg(getErrMsg(this.props.msg.pkg, "upload.add.fail", status));
    }
    this.props.update(updater().updateUploadingsInfo);
  };

  onNewFolderNameChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newFolderName: ev.target.value });
  };

  onMkDir = async () => {
    if (this.state.newFolderName === "") {
      alertMsg(this.props.msg.pkg.get("browser.folder.add.fail"));
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
        alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", mkDirStatus.toString())
        );
        return;
      }

      const setItemsStatus = await updater().setItems(
        this.props.filesInfo.dirPath
      );
      if (setItemsStatus !== "") {
        alertMsg(
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

    this.setLoading(true);

    try {
      const deleteStatus = await updater().delete(
        this.props.filesInfo.dirPath,
        this.props.filesInfo.items,
        this.state.selectedItems
      );
      if (deleteStatus !== "") {
        alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", deleteStatus.toString())
        );
        return deleteStatus;
      }

      const selfStatus = await updater().self();
      if (selfStatus !== "") {
        alertMsg(
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
    const oldDir = this.state.selectedSrc;
    const newDir = this.props.filesInfo.dirPath.join("/");
    if (oldDir === newDir) {
      alertMsg(this.props.msg.pkg.get("browser.move.fail"));
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
        alertMsg(
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
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
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
      alertMsg(this.props.msg.pkg.get("unauthed"));
      return;
    }

    this.setLoading(true);
    try {
      const status = await updater().setItems(dirPath);
      if (status !== "") {
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        return;
      }

      const isSharingStatus = await updater().syncIsSharing(dirPath.join("/"));
      if (isSharingStatus !== "") {
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", isSharingStatus));
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
    alertMsg(this.props.msg.pkg.get("refresh-hint"));
    updater().generateHash(filePath);
  };

  addSharing = async () => {
    this.setLoading(true);

    try {
      const addStatus = await updater().addSharing();
      if (addStatus !== "") {
        alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", addStatus.toString())
        );
        return;
      }

      updater().setSharing(true);
      const listStatus = await updater().listSharings();
      if (listStatus !== "") {
        alertMsg(
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
        alertMsg(
          getErrMsg(this.props.msg.pkg, "op.fail", delStatus.toString())
        );
        return;
      }

      updater().setSharing(false);
      const listStatus = await updater().listSharings();
      if (listStatus !== "") {
        alertMsg(
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

  prepareTable = (
    sortedItems: List<MetadataResp>,
    showOp: string
  ): React.ReactNode => {
    const items = sortedItems.map((item: MetadataResp) => {
      const isSelected = this.state.selectedItems.has(item.name);
      const dirPath = this.props.filesInfo.dirPath.join("/");
      const itemPath = dirPath.endsWith("/")
        ? `${dirPath}${item.name}`
        : `${dirPath}/${item.name}`;

      const icon = item.isDir ? (
        <div className="v-mid item-cell">
          <RiFolder2Fill size="3rem" className="yellow0-font" />
        </div>
      ) : (
        <div className="v-mid item-cell">
          <RiFile2Fill size="3rem" className="cyan1-font" />
        </div>
      );

      const modTimeTitle = this.props.msg.pkg.get("item.modTime");
      const sizeTitle = this.props.msg.pkg.get("item.size");
      const itemSize = FileSize(item.size, { round: 0 });

      const content = item.isDir ? (
        <div className={`v-mid item-cell`}>
          <div className="full-width">
            <div
              className="title-m clickable"
              onClick={() => this.gotoChild(item.name)}
            >
              {item.name}
            </div>
            <div className="desc-m grey0-font">
              <span>
                <span className="grey3-font">{`${modTimeTitle}: `}</span>
                <span>{`${item.modTime}`}</span>
              </span>
            </div>
          </div>
        </div>
      ) : (
        <div>
          <div className={`v-mid item-cell`}>
            <div className="full-width">
              <a
                className="title-m clickable"
                href={`/v1/fs/files?fp=${itemPath}`}
                target="_blank"
              >
                {item.name}
              </a>
              <div className="desc-m grey0-font">
                <span className="grey3-font">{`${sizeTitle}: `}</span>
                <span>{`${itemSize} | `}</span>
                <span className="grey3-font">{`SHA1: `}</span>
                <span>{item.sha1}</span>
              </div>
            </div>
          </div>
        </div>
      );

      const op = item.isDir ? (
        <div className={`v-mid item-cell item-op ${showOp}`}>
          <span onClick={() => this.select(item.name)} className="float-l">
            {isSelected
              ? getIcon("RiCheckboxFill", "1.8rem", "cyan1")
              : getIcon("RiCheckboxBlankLine", "1.8rem", "black1")}
          </span>
        </div>
      ) : (
        <div className={`v-mid item-cell item-op ${showOp}`}>
          <span onClick={() => this.select(item.name)} className="float-l">
            {isSelected
              ? getIcon("RiCheckboxFill", "1.8rem", "cyan1")
              : getIcon("RiCheckboxBlankLine", "1.8rem", "black1")}
          </span>
        </div>
      );

      return {
        val: item,
        cells: List<Cell>([
          { elem: icon, val: item.isDir ? "d" : "f" },
          { elem: content, val: itemPath },
          { elem: op, val: "" },
        ]),
      };
    });

    const tableTitles = List<Head>([
      {
        elem: (
          <div className="font-s grey0-font">
            <RiFileList2Fill size="3rem" className="black-font" />
          </div>
        ),
        sortable: true,
      },
      {
        elem: <div className="font-s grey0-font">Name</div>,
        sortable: true,
      },
      {
        elem: <div className="font-s grey0-font">Action</div>,
        sortable: false,
      },
    ]);

    return (
      <Table
        colStyles={List([
          { width: "3rem", paddingRight: "1rem" },
          { width: "calc(100% - 12rem)", textAlign: "left" },
          { width: "8rem", textAlign: "right" },
        ])}
        id="item-table"
        head={tableTitles}
        foot={List()}
        rows={items}
        updateRows={this.updateItems}
      />
    );
  };

  prepareRows = (
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

      const selectedIconColor = isSelected ? "cyan1-font" : "black1-font";

      const descIconColor = this.state.showDetail.has(item.name)
        ? "cyan1-font"
        : "grey0-font";
      const icon = item.isDir ? (
        <RiFolder2Fill size="1.8rem" className="yellow0-font" />
      ) : (
        <RiFile2Fill size="1.8rem" className="cyan1-font" />
      );
      const fileType = item.isDir
        ? this.props.msg.pkg.get("item.type.folder")
        : this.props.msg.pkg.get("item.type.file");

      const downloadPath = `/v1/fs/files?fp=${itemPath}`;
      const name = item.isDir ? (
        <span className="clickable" onClick={() => this.gotoChild(item.name)}>
          {item.name}
        </span>
      ) : (
        <a className="title-m clickable" href={downloadPath} target="_blank">
          {item.name}
        </a>
      );

      const checkIcon = isSelected ? (
        <RiCheckboxFill
          size="1.8rem"
          className={`${selectedIconColor} ${shareModeClass}`}
          onClick={() => this.select(item.name)}
        />
      ) : (
        <RiCheckboxBlankLine
          size="1.8rem"
          className={`${selectedIconColor} ${shareModeClass}`}
          onClick={() => this.select(item.name)}
        />
      );

      const op = item.isDir ? (
        <div className={`v-mid item-op ${showOp}`}>{checkIcon}</div>
      ) : (
        <div className={`v-mid item-op ${showOp}`}>
          <RiMore2Fill
            size="1.8rem"
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
      const fileTypeTitle = this.props.msg.pkg.get("item.type");
      const itemSize = FileSize(item.size, { round: 0 });

      const compact = item.isDir ? (
        <span>
          <span className="grey3-font">{`${modTimeTitle}: `}</span>
          <span>{item.modTime}</span>
        </span>
      ) : (
        <span>
          <span className="grey3-font">{`${pathTitle}: `}</span>
          <span>{`${absDownloadURL} | `}</span>
          <span className="grey3-font">{`${modTimeTitle}: `}</span>
          <span>{`${item.modTime} | `}</span>
          <span className="grey3-font">{`${sizeTitle}: `}</span>
          <span>{`${itemSize} | `}</span>
          <span className="grey3-font">{`SHA1: `}</span>
          <span>{item.sha1}</span>
        </span>
      );
      const details = (
        <div>
          <div className="column">
            <div className="card">
              <span className="title-m black-font">{pathTitle}</span>
              <span>{absDownloadURL}</span>
            </div>
          </div>

          <div className="column">
            <div className="card margin-l-m">
              <span className="title-m black-font">{modTimeTitle}</span>
              <span>{item.modTime}</span>
            </div>
            <div className="card margin-l-m">
              <span className="title-m black-font">{sizeTitle}</span>
              <span>{itemSize}</span>
            </div>
          </div>

          <div className="fix">
            <div className="card">
              <Flexbox
                children={List([
                  <span className="title-m black-font">SHA1</span>,
                  <RiRestartFill
                    onClick={() => this.generateHash(itemPath)}
                    size={"2rem"}
                    className={`grey3-font ${shareModeClass}`}
                  />,
                ])}
                childrenStyles={List([{}, { justifyContent: "flex-end" }])}
              />
              <div className="info">{item.sha1}</div>
            </div>
          </div>
        </div>
      );
      const desc = this.state.showDetail.has(item.name) ? details : compact;

      const elem = (
        <div>
          <Flexbox
            children={List([
              <div>
                <div className="v-mid">
                  {icon}
                  <span className="margin-l-m desc-l">{`${fileTypeTitle}`}</span>
                  &nbsp;
                  <span className="desc-l grey0-font">{`- ${fileType}`}</span>
                </div>
              </div>,
              <div>{op}</div>,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />
          <div className="name">{name}</div>
          <div className="desc">{desc}</div>
          <div className="hr"></div>
        </div>
      );

      return elem;
    });

    return <Rows rows={List(rows)} />;
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
          <h5 className="pane-title margin-r-m">
            {this.props.msg.pkg.get("endpoints")}
          </h5>
          <div className="hr"></div>

          <button onClick={gotoRoot} className="margin-r-m">
            {this.props.msg.pkg.get("endpoints.root")}
          </button>
          <button onClick={this.goHome}>
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
      <div id="upload-op">
        <Flexbox
          children={List([
            <div>
              <button onClick={this.onMkDir} className="float cyan-btn">
                {this.props.msg.pkg.get("browser.folder.add")}
              </button>
              <input
                type="text"
                onChange={this.onNewFolderNameChange}
                value={this.state.newFolderName}
                placeholder={this.props.msg.pkg.get("browser.folder.name")}
                className="float"
              />
            </div>,

            <div>
              <button onClick={this.onClickUpload} className="cyan-btn">
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
      viewType === "rows" ? (
        <div id="item-rows">
          {this.prepareRows(this.props.filesInfo.items, showOp)}
        </div>
      ) : (
        this.prepareTable(this.props.filesInfo.items, showOp)
      );

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
        ? "cyan1-font"
        : "black-font";
    const tableViewColorClass =
      this.props.ui.control.controls.get(filesViewCtrl) === "table"
        ? "cyan1-font"
        : "black-font";

    const itemListPane = (
      <div>
        {endPoints}

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
                      className="red-btn left"
                    >
                      {this.props.msg.pkg.get("browser.share.del")}
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={this.addSharing}
                      className="cyan-btn left"
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
                        className="red-btn left"
                      >
                        {this.props.msg.pkg.get("browser.delete")}
                      </button>

                      <button type="button" onClick={() => this.moveHere()}>
                        {this.props.msg.pkg.get("browser.paste")}
                      </button>
                    </span>
                  ) : null}
                </span>,

                <Flexbox
                  children={List([
                    <BiListUl
                      size="2rem"
                      className={`${rowsViewColorClass} margin-r-s`}
                      onClick={() => {
                        this.setView("rows");
                      }}
                    />,
                    <BiTable
                      size="2rem"
                      className={`${tableViewColorClass} margin-r-s`}
                      onClick={() => {
                        this.setView("table");
                      }}
                    />,

                    <span className={`${showOp}`}>
                      <button
                        onClick={() => this.selectAll()}
                        className="select-btn"
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
          {orderByButtons}
          {view}
        </Container>
      </div>
    );

    return <div id="browser">{itemListPane}</div>;
  }
}
