import * as React from "react";
import { createRef, RefObject } from "react";
import { List, Map, Set } from "immutable";
import FileSize from "filesize";

import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiFile2Fill } from "@react-icons/all-files/ri/RiFile2Fill";
import { RiCheckboxFill } from "@react-icons/all-files/ri/RiCheckboxFill";
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
import { CronJobs } from "../common/cron";
import { Title } from "./visual/title";
import { NotFoundBanner } from "./visual/banner_notfound";
import { getIconWithProps } from "./visual/icons";

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
  searchResults: List<string>;
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
  searchKeywords: string;
}

export class FilesPanel extends React.Component<Props, State, {}> {
  private hotkeyHandler: HotkeyHandler;
  private uploadInput: RefObject<HTMLInputElement> = createRef();

  constructor(p: Props) {
    super(p);
    this.state = {
      newFolderName: "",
      selectedSrc: "",
      selectedItems: Map<string, boolean>(),
      showDetail: Set<string>(),
      uploadFiles: "",
      searchKeywords: "",
    };

    Up().setStatusCb(this.updateProgress);
  }

  componentDidMount(): void {
    CronJobs().setInterval("refreshFileList", {
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
    CronJobs().clearInterval("refreshFileList");

    document.removeEventListener("keyup", this.hotkeyHandler.handle);
  }

  onClickUpload = () => {
    if (!this.props.enabled) {
      return;
    }
    this.uploadInput.current.click();
  };

  onNewFolderNameChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ newFolderName: ev.target.value });
  };

  onSearchKeywordsChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ searchKeywords: ev.target.value });
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

  addFileList = (originalFileList: FileList) => {
    if (originalFileList.length > 200) {
      Env().alertMsg(this.props.msg.pkg.get("err.tooManyUploads"));
      return;
    }

    let fileList = List<File>();
    for (let i = 0; i < originalFileList.length; i++) {
      fileList = fileList.push(originalFileList[i]);
    }

    const status = updater().addUploads(fileList);
    if (status !== "") {
      Env().alertMsg(getErrMsg(this.props.msg.pkg, "upload.add.fail", status));
    }
    this.props.update(updater().updateUploadingsInfo);
  };

  addUploads = (event: React.ChangeEvent<HTMLInputElement>) => {
    this.addFileList(event.target.files);
  };

  mkDirFromKb = async (
    event: React.KeyboardEvent<HTMLInputElement>
  ): Promise<void> => {
    if (event.key === "Enter") {
      return await this.mkDir();
    }
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

    const rows = sortedItems.map(
      (item: MetadataResp, i: number): React.ReactNode => {
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
          <RiFolder2Fill
            size="1.6rem"
            className="yellow0-font mr-8 inline-block h-[3.2rem]"
          />
        ) : (
          <RiFile2Fill
            size="1.6rem"
            className="focus-font mr-8 inline-block h-[3.2rem]"
          />
        );

        const modTimeDate = new Date(item.modTime);
        const modTimeFormatted = `${modTimeDate.getFullYear()}-${
          modTimeDate.getMonth() + 1
        }-${modTimeDate.getDate()}`;
        const downloadPath = `/v1/fs/files?fp=${itemPath}`;
        const name = item.isDir ? (
          <span className="overflow-hidden break-words block">
            <span
              className="clickable leading-default"
              onClick={() => this.gotoChild(item.name)}
            >
              {item.name}
            </span>
            <span className="major-font">{` - ${modTimeFormatted}`}</span>
          </span>
        ) : (
          <span className="overflow-hidden break-words block">
            <a
              className="clickable leading-default"
              href={downloadPath}
              target="_blank"
            >
              {item.name}
            </a>
            <span className="major-font">{` - ${modTimeFormatted}`}</span>
          </span>
        );

        const checkIcon = isSelected ? (
          <RiCheckboxFill
            size="1.6rem"
            className={`${selectedIconColor} ${shareModeClass} inline-block h-[3.2rem]`}
            onClick={() => this.select(item.name)}
          />
        ) : (
          <RiCheckboxBlankLine
            size="1.6rem"
            className={`${selectedIconColor} ${shareModeClass} inline-block h-[3.2rem]`}
            onClick={() => this.select(item.name)}
          />
        );

        const detailIcon = getIconWithProps("GoUnfold", {
          size: "1.6rem",
          className: `${descIconColor} inline-block h-[3.2rem]`,
          onClick: () => this.toggleDetail(item.name),
        });

        const absDownloadURL = `${document.location.protocol}//${document.location.host}${downloadPath}`;
        const pathTitle = this.props.msg.pkg.get("item.downloadURL");
        const modTimeTitle = this.props.msg.pkg.get("item.modTime");
        const sizeTitle = this.props.msg.pkg.get("item.size");
        const itemSize = FileSize(item.size, { round: 0 });

        const descStateClass = this.state.showDetail.has(item.name)
          ? "margin-t-m padding-m"
          : "no-height";
        const desc = (
          <div className={`${descStateClass}  major-bg`}>
            <div className="column">
              <div className="card">
                <span className="title-m ">{pathTitle}</span>
                <span className="work-break-all">{absDownloadURL}</span>
              </div>
            </div>

            <div className="column">
              <div className="card">
                <span className="title-m ">{modTimeTitle}</span>
                <span className="work-break-all">{modTimeFormatted}</span>
              </div>
              <div className="card">
                <span className="title-m ">{sizeTitle}</span>
                <span className="work-break-all">{itemSize}</span>
              </div>
            </div>

            <div className="fix">
              <div className="card">
                <Flexbox
                  children={List([
                    <span className="title-m ">SHA1</span>,
                    <RiRestartFill
                      onClick={() => this.generateHash(itemPath)}
                      size={"2rem"}
                      className={` ${shareModeClass}`}
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
          <Flexbox children={List([checkIcon])} childrenStyles={List([{}])} />,
          <Flexbox children={List([icon])} childrenStyles={List([{}])} />,
          name,
          <div className="major-font padding-l-s leading-default">
            {itemSize}
          </div>,
          detailIcon,
        ]);

        const tableCols = (
          <Columns
            rows={List([cells])}
            widths={List([
              "3rem",
              "3rem",
              "calc(100% - 18rem)",
              "8rem",
              "3rem",
            ])}
            childrenClassNames={List(["", "", "", "", "text-right"])}
            colKey={`filesPanel-${i}`}
          />
        );

        return (
          <div key={`filesPanel-row-${i}`}>
            <div className="h-[3.6rem]">{tableCols}</div>
            {desc}
          </div>
        );
      }
    );

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

  search = async () => {
    if (this.state.searchKeywords.trim() === "") {
      Env().alertMsg(this.props.msg.pkg.get("hint.keywords"));
      return;
    }
    const keywords = this.state.searchKeywords.split(" ");
    if (keywords.length === 0) {
      Env().alertMsg(this.props.msg.pkg.get("hint.keywords"));
      return;
    }

    const status = await updater().search(keywords);
    if (status !== "") {
      Env().alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
      return;
    } else if (!updater().hasResult()) {
      Env().alertMsg(this.props.msg.pkg.get("term.noResult"));
      return;
    }
    this.props.update(updater().updateFilesInfo);
  };

  searchKb = async (
    event: React.KeyboardEvent<HTMLInputElement>
  ): Promise<void> => {
    if (event.key === "Enter") {
      return await this.search();
    }
  };

  gotoSearchResult = async (pathname: string) => {
    this.setLoading(true);
    try {
      const status = await updater().gotoSearchResult(pathname);
      if (status !== "") {
        Env().alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
      }
    } finally {
      this.setLoading(false);
    }
  };

  truncateSearchResults = async () => {
    updater().truncateSearchResults();
    this.setState({ searchKeywords: "" });
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
          <Flexbox
            children={List([
              <>
                <button onClick={gotoRoot} className="mr-8">
                  {this.props.msg.pkg.get("endpoints.root")}
                </button>
                <button onClick={this.goHome} className="">
                  {this.props.msg.pkg.get("endpoints.home")}
                </button>
              </>,

              <>
                <input
                  type="text"
                  onChange={this.onSearchKeywordsChange}
                  onKeyUp={this.searchKb}
                  value={this.state.searchKeywords}
                  placeholder={this.props.msg.pkg.get("hint.keywords")}
                  className="inline-block mr-8"
                />
                <button onClick={this.search} className="mr-8">
                  {this.props.msg.pkg.get("term.search")}
                </button>
                <button onClick={this.truncateSearchResults}>
                  {this.props.msg.pkg.get("reset")}
                </button>
              </>,
            ])}
            childrenStyles={List([
              { flex: "0 0 50%" },
              { flex: "0 0 50%", justifyContent: "flex-end" },
            ])}
          />
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
          <span
            key={`${pathPart}-${key}`}
            className="inline-block leading-default"
          >
            <button
              onClick={() =>
                this.chdir(this.props.filesInfo.dirPath.slice(0, key + 1))
              }
              className="item clickable"
            >
              <span className="content">
                {pathPart === "/" ? "~" : pathPart}
              </span>
            </button>
            <span className="item">
              <span className="content">{"/"}</span>
            </span>
          </span>
        );
      }
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

    const ops = (
      <div>
        <Flexbox
          children={List([
            <div>
              <button onClick={this.mkDir} className="inline-block mr-8">
                {this.props.msg.pkg.get("browser.folder.add")}
              </button>
              <input
                type="text"
                onChange={this.onNewFolderNameChange}
                onKeyUp={this.mkDirFromKb}
                value={this.state.newFolderName}
                placeholder={this.props.msg.pkg.get("browser.folder.name")}
                className="inline-block"
              />
            </div>,

            <div>
              <span
                id="space-used"
                className={`minor-font mr-8 text-right leading-default ${shareModeClass}`}
              >
                {`${this.props.msg.pkg.get(
                  "browser.used"
                )} ${usedSpace} / ${spaceLimit}`}
              </span>

              <button onClick={this.onClickUpload}>
                {this.props.msg.pkg.get("browser.upload")}
              </button>
              <input
                type="file"
                onChange={this.addUploads}
                multiple={true}
                value={this.state.uploadFiles}
                ref={this.uploadInput}
                className="hidden"
              />
            </div>,
          ])}
          childrenStyles={List([
            { flex: "0 0 60%" },
            { flex: "0 0 40%", justifyContent: "flex-end" },
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
        titleIcon="TbSortAscending2"
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
        <div className={`${showOp}`}>
          <Flexbox
            children={List([
              <div>
                <span className={`${showOp} mr-8`}>
                  <button onClick={() => this.selectAll()}>
                    {this.props.msg.pkg.get("browser.selectAll")}
                  </button>
                </span>
                <span>
                  {this.state.selectedItems.size > 0 ? (
                    <span>
                      <button
                        type="button"
                        onClick={() => this.delete()}
                        className="mr-8"
                      >
                        {this.props.msg.pkg.get("browser.delete")}
                      </button>

                      <button type="button" onClick={() => this.moveHere()}>
                        {this.props.msg.pkg.get("browser.paste")}
                      </button>
                    </span>
                  ) : null}
                </span>
              </div>,

              orderByButtons,
            ])}
            childrenStyles={List([{}, { justifyContent: "flex-end" }])}
          />

          <div className="my-8">
            {this.prepareColumns(this.props.filesInfo.items, showOp)}
          </div>
        </div>
      ) : (
        <NotFoundBanner title={this.props.msg.pkg.get("terms.nothingHere")} />
      ); // TODO: support better views in the future

    const rowsViewColorClass =
      this.props.ui.control.controls.get(filesViewCtrl) === "rows"
        ? "focus-font"
        : "major-font";
    const tableViewColorClass =
      this.props.ui.control.controls.get(filesViewCtrl) === "table"
        ? "focus-font"
        : "major-font";

    const showSearchResults =
      this.props.filesInfo.searchResults.size > 0 ? "" : "hidden";
    const searchResultPane = this.props.filesInfo.searchResults.map(
      (searchResult: string) => {
        return (
          <>
            <Flexbox
              children={List([
                <span>{searchResult}</span>,
                <button
                  type="button"
                  onClick={() => {
                    this.gotoSearchResult(searchResult);
                  }}
                >
                  {this.props.msg.pkg.get("action.go")}
                </button>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            <div className="hr"></div>
          </>
        );
      }
    );

    const itemListPane = (
      <div>
        {endPoints}

        <div className={showSearchResults}>
          <Container>
            <Title
              title={this.props.msg.pkg.get("term.results")}
              iconColor="normal"
              iconName="RiFileSearchFill"
            />
            <div className="hr"></div>
            {searchResultPane}
          </Container>
        </div>

        <div className={showOp}>
          <Container>{ops}</Container>
        </div>

        <Container>
          <div id="breadcrumb" className="leading-default">
            <span className="location-item">
              <span className="content">
                {this.props.filesInfo.isSharing ? (
                  <button
                    type="button"
                    onClick={() => {
                      this.deleteSharing(
                        this.props.filesInfo.dirPath.join("/")
                      );
                    }}
                  >
                    {this.props.msg.pkg.get("browser.share.del")}
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={this.addSharing}
                    className="mr-4"
                  >
                    {this.props.msg.pkg.get("browser.share.add")}
                  </button>
                )}
                {/* {`${this.props.msg.pkg.get("breadcrumb.loc")}:`} */}
              </span>
            </span>
            {breadcrumb}
          </div>
          {view}
        </Container>
      </div>
    );

    return <div id="browser">{itemListPane}</div>;
  }
}
