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
import { UploadEntry, UploadState } from "../worker/interface";
import { Flexbox } from "./layout/flexbox";

// export interface Item {
//   name: string;
//   size: number;
//   modTime: string;
//   isDir: boolean;
//   selected: boolean;
//   sha1: string;
// }

// export interface BrowserProps {
//   tab: string;

//   dirPath: List<string>;
//   isSharing: boolean;
//   items: List<MetadataResp>;
  
//   uploadings: List<UploadEntry>;
//   uploadFiles: List<File>;
//   uploadValue: string;
  
//   sharings: List<string>;
// }

// export interface Props {
//   browser: BrowserProps;
//   msg: MsgProps;
//   login: LoginProps;
//   ui: UIProps;
//   update?: (updater: (prevState: ICoreState) => ICoreState) => void;
// }

export function getItemPath(dirPath: string, itemName: string): string {
  return dirPath.endsWith("/")
    ? `${dirPath}${itemName}`
    : `${dirPath}/${itemName}`;
}

// export interface State {
//   newFolderName: string;
//   selectedSrc: string;
//   selectedItems: Map<string, boolean>;
//   showDetail: Set<string>;
// }

// export class Browser extends React.Component<Props, State, {}> {
//   private uploadInput: Element | Text;
//   private assignInput: (input: Element) => void;
//   private onClickUpload: () => void;

//   constructor(p: Props) {
//     super(p);
//     this.state = {
//       newFolderName: "",
//       selectedSrc: "",
//       selectedItems: Map<string, boolean>(),
//       showDetail: Set<string>(),
//     };

//     Up().setStatusCb(this.updateProgress);
//     this.uploadInput = undefined;
//     this.assignInput = (input) => {
//       this.uploadInput = ReactDOM.findDOMNode(input);
//     };
//     this.onClickUpload = () => {
//       const uploadInput = this.uploadInput as HTMLButtonElement;
//       uploadInput.click();
//     };
//   }

//   onNewFolderNameChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
//     this.setState({ newFolderName: ev.target.value });
//   };

//   addUploads = (event: React.ChangeEvent<HTMLInputElement>) => {
//     if (event.target.files.length > 1000) {
//       alertMsg(this.props.msg.pkg.get("err.tooManyUploads"));
//       return;
//     }

//     let fileList = List<File>();
//     for (let i = 0; i < event.target.files.length; i++) {
//       fileList = fileList.push(event.target.files[i]);
//     }
//     updater().addUploads(fileList);
//     this.props.update(updater().updateBrowser);
//   };

//   deleteUpload = (filePath: string): Promise<void> => {
//     return updater()
//       .deleteUpload(filePath)
//       .then((ok: boolean) => {
//         if (!ok) {
//           alertMsg(this.props.msg.pkg.get("browser.upload.del.fail"));
//         }
//         return updater().refreshUploadings();
//       })
//       .then(() => {
//         return updater().self();
//       })
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//         this.props.update(updater().updateLogin);
//       });
//   };

//   stopUploading = (filePath: string) => {
//     updater().stopUploading(filePath);
//     this.props.update(updater().updateBrowser);
//   };

//   onMkDir = () => {
//     if (this.state.newFolderName === "") {
//       alertMsg(this.props.msg.pkg.get("browser.folder.add.fail"));
//       return;
//     }

//     const dirPath = getItemPath(
//       this.props.browser.dirPath.join("/"),
//       this.state.newFolderName
//     );
//     updater()
//       .mkDir(dirPath)
//       .then(() => {
//         this.setState({ newFolderName: "" });
//         return updater().setItems(this.props.browser.dirPath);
//       })
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//       });
//   };

//   delete = () => {
//     // TODO: selected should be cleaned after change the cwd
//     if (this.props.browser.dirPath.join("/") !== this.state.selectedSrc) {
//       alertMsg(this.props.msg.pkg.get("browser.del.fail"));
//       this.setState({
//         selectedSrc: this.props.browser.dirPath.join("/"),
//         selectedItems: Map<string, boolean>(),
//       });
//       return;
//     } else {
//       const filesToDel = this.state.selectedItems.keySeq().join(", ");
//       if (!confirmMsg(`${this.props.msg.pkg.get("op.confirm")} [${this.state.selectedItems.size}]: ${filesToDel}`)) {
//         return;
//       }
//     }

//     updater()
//       .delete(
//         this.props.browser.dirPath,
//         this.props.browser.items,
//         this.state.selectedItems
//       )
//       .then(() => {
//         return updater().self();
//       })
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//         this.props.update(updater().updateLogin);
//         this.setState({
//           selectedSrc: "",
//           selectedItems: Map<string, boolean>(),
//         });
//       });
//   };

//   moveHere = () => {
//     const oldDir = this.state.selectedSrc;
//     const newDir = this.props.browser.dirPath.join("/");
//     if (oldDir === newDir) {
//       alertMsg(this.props.msg.pkg.get("browser.move.fail"));
//       return;
//     }

//     updater()
//       .moveHere(
//         this.state.selectedSrc,
//         this.props.browser.dirPath.join("/"),
//         this.state.selectedItems
//       )
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//         this.setState({
//           selectedSrc: "",
//           selectedItems: Map<string, boolean>(),
//         });
//       });
//   };

//   gotoChild = async (childDirName: string) => {
//     return this.chdir(this.props.browser.dirPath.push(childDirName));
//   };

//   chdir = async (dirPath: List<string>) => {
//     if (dirPath === this.props.browser.dirPath) {
//       return;
//     } else if (this.props.login.userRole !== roleAdmin && dirPath.size <= 1) {
//       alertMsg(this.props.msg.pkg.get("unauthed"));
//       return;
//     }

//     return updater()
//       .setItems(dirPath)
//       .then(() => {
//         return updater().listSharings();
//       })
//       .then(() => {
//         return updater().isSharing(dirPath.join("/"));
//       })
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//       });
//   };

//   updateProgress = async (
//     infos: Map<string, UploadEntry>,
//     refresh: boolean
//   ) => {
//     updater().setUploadings(infos);
//     let errCount = 0;
//     infos.valueSeq().forEach((entry: UploadEntry) => {
//       errCount += entry.state === UploadState.Error ? 1 : 0;
//     });

//     if (infos.size === 0 || infos.size === errCount) {
//       // refresh used space
//       updater()
//         .self()
//         .then(() => {
//           this.props.update(updater().updateLogin);
//         });
//     }

//     if (refresh) {
//       updater()
//         .setItems(this.props.browser.dirPath)
//         .then(() => {
//           this.props.update(updater().updateBrowser);
//         });
//     } else {
//       this.props.update(updater().updateBrowser);
//     }
//   };

//   select = (itemName: string) => {
//     const selectedItems = this.state.selectedItems.has(itemName)
//       ? this.state.selectedItems.delete(itemName)
//       : this.state.selectedItems.set(itemName, true);

//     this.setState({
//       selectedSrc: this.props.browser.dirPath.join("/"),
//       selectedItems: selectedItems,
//     });
//   };

//   selectAll = () => {
//     let newSelected = Map<string, boolean>();
//     const someSelected = this.state.selectedItems.size === 0 ? true : false;
//     if (someSelected) {
//       this.props.browser.items.forEach((item) => {
//         newSelected = newSelected.set(item.name, true);
//       });
//     } else {
//       this.props.browser.items.forEach((item) => {
//         newSelected = newSelected.delete(item.name);
//       });
//     }

//     this.setState({
//       selectedSrc: this.props.browser.dirPath.join("/"),
//       selectedItems: newSelected,
//     });
//   };

//   addSharing = async () => {
//     return updater()
//       .addSharing()
//       .then((ok) => {
//         if (!ok) {
//           alertMsg(this.props.msg.pkg.get("browser.share.add.fail"));
//         } else {
//           updater().setSharing(true);
//           return this.listSharings();
//         }
//       })
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//       });
//   };

//   deleteSharing = async (dirPath: string) => {
//     return updater()
//       .deleteSharing(dirPath)
//       .then((ok) => {
//         if (!ok) {
//           alertMsg(this.props.msg.pkg.get("browser.share.del.fail"));
//         } else {
//           updater().setSharing(false);
//           return this.listSharings();
//         }
//       })
//       .then(() => {
//         this.props.update(updater().updateBrowser);
//       });
//   };

//   listSharings = async () => {
//     return updater()
//       .listSharings()
//       .then((ok) => {
//         if (ok) {
//           this.props.update(updater().updateBrowser);
//         }
//       });
//   };

//   setTab = (tabName: string) => {
//     updater().setTab(tabName);
//     this.props.update(updater().updateBrowser);
//   };

//   toggleDetail = (name: string) => {
//     const showDetail = this.state.showDetail.has(name)
//       ? this.state.showDetail.delete(name)
//       : this.state.showDetail.add(name);
//     this.setState({ showDetail });
//   };

//   generateHash = async (filePath: string): Promise<boolean> => {
//     alertMsg(this.props.msg.pkg.get("refresh-hint"));
//     return updater().generateHash(filePath);
//   };

//   render() {
//     const showOp = this.props.login.userRole === roleVisitor ? "hidden" : "";
//     const breadcrumb = this.props.browser.dirPath.map(
//       (pathPart: string, key: number) => {
//         return (
//           <button
//             key={pathPart}
//             onClick={() =>
//               this.chdir(this.props.browser.dirPath.slice(0, key + 1))
//             }
//             className="item"
//           >
//             {pathPart}
//           </button>
//         );
//       }
//     );

//     const nameWidthClass = `item-name item-name-${
//       this.props.ui.isVertical ? "vertical" : "horizontal"
//     } pointer`;

//     const ops = (
//       <div id="upload-op">
//         <div className="float">
//           <input
//             type="text"
//             onChange={this.onNewFolderNameChange}
//             value={this.state.newFolderName}
//             placeholder={this.props.msg.pkg.get("browser.folder.name")}
//             className="float"
//           />
//           <button onClick={this.onMkDir} className="float">
//             {this.props.msg.pkg.get("browser.folder.add")}
//           </button>
//         </div>

//         <div className="float">
//           <button onClick={this.onClickUpload}>
//             {this.props.msg.pkg.get("browser.upload")}
//           </button>
//           <input
//             type="file"
//             onChange={this.addUploads}
//             multiple={true}
//             value={this.props.browser.uploadValue}
//             ref={this.assignInput}
//             className="hidden"
//           />
//         </div>
//       </div>
//     );

//     const sortedItems = this.props.browser.items.sort(
//       (item1: MetadataResp, item2: MetadataResp) => {
//         if (item1.isDir && !item2.isDir) {
//           return -1;
//         } else if (!item1.isDir && item2.isDir) {
//           return 1;
//         }
//         return 0;
//       }
//     );

//     const itemList = sortedItems.map((item: MetadataResp) => {
//       const isSelected = this.state.selectedItems.has(item.name);
//       const dirPath = this.props.browser.dirPath.join("/");
//       const itemPath = dirPath.endsWith("/")
//         ? `${dirPath}${item.name}`
//         : `${dirPath}/${item.name}`;

//       return item.isDir ? (
//         <Flexbox
//           key={item.name}
//           children={List([
//             <span className="padding-m">
//               <Flexbox
//                 children={List([
//                   <RiFolder2Fill
//                     size="3rem"
//                     className="yellow0-font margin-r-m"
//                   />,

//                   <span className={`${nameWidthClass}`}>
//                     <span
//                       className="title-m"
//                       onClick={() => this.gotoChild(item.name)}
//                     >
//                       {item.name}
//                     </span>
//                     <div className="desc-m grey0-font">
//                       <span>
//                         {item.modTime.slice(0, item.modTime.indexOf("T"))}
//                       </span>
//                     </div>
//                   </span>,
//                 ])}
//                 childrenStyles={List([
//                   { flex: "0 0 auto" },
//                   { flex: "0 0 auto" },
//                 ])}
//               />
//             </span>,

//             <span className={`item-op padding-m ${showOp}`}>
//               <button
//                 onClick={() => this.select(item.name)}
//                 className={`${
//                   isSelected ? "cyan0-bg white-font" : "grey2-bg grey3-font"
//                 }`}
//                 style={{ width: "8rem", display: "inline-block" }}
//               >
//                 {isSelected
//                   ? this.props.msg.pkg.get("browser.deselect")
//                   : this.props.msg.pkg.get("browser.select")}
//               </button>
//             </span>,
//           ])}
//           childrenStyles={List([
//             { flex: "0 0 auto", width: "60%" },
//             { flex: "0 0 auto", justifyContent: "flex-end", width: "40%" },
//           ])}
//         />
//       ) : (
//         <div key={item.name}>
//           <Flexbox
//             key={item.name}
//             children={List([
//               <Flexbox
//                 children={List([
//                   <RiFile2Fill
//                     size="3rem"
//                     className="cyan0-font margin-l-m margin-r-m"
//                   />,

//                   <span className={`${nameWidthClass}`}>
//                     <a
//                       className="title-m"
//                       href={`/v1/fs/files?fp=${itemPath}`}
//                       target="_blank"
//                     >
//                       {item.name}
//                     </a>
//                     <div className="desc-m grey0-font">
//                       <span>
//                         {item.modTime.slice(0, item.modTime.indexOf("T"))}
//                       </span>
//                       &nbsp;/&nbsp;
//                       <span>{FileSize(item.size, { round: 0 })}</span>
//                     </div>
//                   </span>,
//                 ])}
//                 childrenStyles={List([
//                   { flex: "0 0 auto" },
//                   { flex: "0 0 auto" },
//                 ])}
//               />,

//               <span className={`item-op padding-m ${showOp}`}>
//                 <button
//                   onClick={() => this.toggleDetail(item.name)}
//                   style={{ width: "8rem" }}
//                   className="float-input"
//                 >
//                   {this.props.msg.pkg.get("detail")}
//                 </button>

//                 <button
//                   type="button"
//                   onClick={() => this.select(item.name)}
//                   className={`float-input ${
//                     isSelected ? "cyan0-bg white-font " : "grey2-bg grey3-font "
//                   }`}
//                   style={{ width: "8rem" }}
//                 >
//                   {isSelected
//                     ? this.props.msg.pkg.get("browser.deselect")
//                     : this.props.msg.pkg.get("browser.select")}
//                 </button>
//               </span>,
//             ])}
//             childrenStyles={List([
//               { flex: "0 0 auto", width: "60%" },
//               { flex: "0 0 auto", justifyContent: "flex-end", width: "40%" },
//             ])}
//           />

//           <div
//             className={`${
//               this.state.showDetail.has(item.name) ? "" : "hidden"
//             }`}
//           >
//             <Flexbox
//               children={List([
//                 <span>
//                   <b>SHA1:</b>
//                   {` ${item.sha1}`}
//                 </span>,
//                 <button
//                   onClick={() => this.generateHash(itemPath)}
//                   className="black-bg white-font margin-l-m"
//                   style={{ display: "inline-block" }}
//                 >
//                   {this.props.msg.pkg.get("refresh")}
//                 </button>,
//               ])}
//               className={`grey2-bg grey3-font detail margin-r-m`}
//               childrenStyles={List([{}, { justifyContent: "flex-end" }])}
//             />
//           </div>
//         </div>
//       );
//     });

//     const usedSpace = FileSize(parseInt(this.props.login.usedSpace, 10), {
//       round: 0,
//     });
//     const spaceLimit = FileSize(
//       parseInt(this.props.login.quota.spaceLimit, 10),
//       {
//         round: 0,
//       }
//     );

//     const itemListPane =
//       this.props.browser.tab === "" || this.props.browser.tab === "item" ? (
//         <div id="item-list">
//           <div className={`container ${showOp}`}>{ops}</div>

//           <div className="container">
//             <div id="browser-op" className={`${showOp}`}>
//               <Flexbox
//                 children={List([
//                   <span>
//                     {this.props.browser.isSharing ? (
//                       <button
//                         type="button"
//                         onClick={() => {
//                           this.deleteSharing(
//                             this.props.browser.dirPath.join("/")
//                           );
//                         }}
//                         className="red-btn"
//                       >
//                         {this.props.msg.pkg.get("browser.share.del")}
//                       </button>
//                     ) : (
//                       <button
//                         type="button"
//                         onClick={this.addSharing}
//                         className="cyan-btn"
//                       >
//                         {this.props.msg.pkg.get("browser.share.add")}
//                       </button>
//                     )}
//                   </span>,

//                   <span>
//                     {this.state.selectedItems.size > 0 ? (
//                       <span>
//                         <button
//                           type="button"
//                           onClick={() => this.delete()}
//                           className="red-btn"
//                         >
//                           {this.props.msg.pkg.get("browser.delete")}
//                         </button>

//                         <button type="button" onClick={() => this.moveHere()}>
//                           {this.props.msg.pkg.get("browser.paste")}
//                         </button>
//                       </span>
//                     ) : null}
//                   </span>,

//                   <span>
//                     <span
//                       id="space-used"
//                       className="desc-m grey0-font"
//                     >{`${this.props.msg.pkg.get(
//                       "browser.used"
//                     )} ${usedSpace} / ${spaceLimit}`}</span>
//                   </span>,
//                 ])}
//                 childrenStyles={List([
//                   { flex: "0 0 auto" },
//                   { flex: "0 0 auto" },
//                   { justifyContent: "flex-end" },
//                 ])}
//               />
//             </div>

//             <Flexbox
//               children={List([
//                 <span id="breadcrumb">
//                   <Flexbox
//                     children={List([
//                       <RiHomeSmileFill size="3rem" id="icon-home" />,
//                       <Flexbox children={breadcrumb} />,
//                     ])}
//                     childrenStyles={List([
//                       { flex: "0 0 auto" },
//                       { flex: "0 0 auto" },
//                     ])}
//                   />
//                 </span>,

//                 <span className={`${showOp}`}>
//                   <button
//                     onClick={() => this.selectAll()}
//                     className="select-btn"
//                   >
//                     {this.props.msg.pkg.get("browser.selectAll")}
//                   </button>
//                 </span>,
//               ])}
//               childrenStyles={List([{}, { justifyContent: "flex-end" }])}
//             />

//             {itemList}
//           </div>
//         </div>
//       ) : null;

//     const uploadingList = this.props.browser.uploadings.map(
//       (uploading: UploadEntry) => {
//         const pathParts = uploading.filePath.split("/");
//         const fileName = pathParts[pathParts.length - 1];

//         return (
//           <div key={uploading.filePath}>
//             <Flexbox
//               children={List([
//                 <span className="padding-m">
//                   <Flexbox
//                     children={List([
//                       <RiUploadCloudLine
//                         size="3rem"
//                         id="icon-upload"
//                         className="margin-r-m blue0-font"
//                       />,

//                       <div className={`${nameWidthClass}`}>
//                         <span className="title-m">{fileName}</span>
//                         <div className="desc-m grey0-font">
//                           {FileSize(uploading.uploaded, { round: 0 })}
//                           &nbsp;/&nbsp;{FileSize(uploading.size, { round: 0 })}
//                         </div>
//                       </div>,
//                     ])}
//                   />
//                 </span>,

//                 <div className="item-op">
//                   <button
//                     onClick={() => this.stopUploading(uploading.filePath)}
//                     className="float-input"
//                   >
//                     {this.props.msg.pkg.get("browser.stop")}
//                   </button>
//                   <button
//                     onClick={() => this.deleteUpload(uploading.filePath)}
//                     className="float-input"
//                   >
//                     {this.props.msg.pkg.get("browser.delete")}
//                   </button>
//                 </div>,
//               ])}
//               childrenStyles={List([{}, { justifyContent: "flex-end" }])}
//             />
//             {uploading.err.trim() === "" ? null : (
//               <div className="error">{uploading.err.trim()}</div>
//             )}
//           </div>
//         );
//       }
//     );

//     const uploadingListPane =
//       this.props.browser.tab === "uploading" ? (
//         this.props.browser.uploadings.size === 0 ? (
//           <div className="container">
//             <Flexbox
//               children={List([
//                 <RiEmotionSadLine
//                   size="4rem"
//                   className="margin-r-m red0-font"
//                 />,
//                 <span>
//                   <h3 className="title-l">
//                     {this.props.msg.pkg.get("upload.404.title")}
//                   </h3>
//                   <span className="desc-l grey0-font">
//                     {this.props.msg.pkg.get("upload.404.desc")}
//                   </span>
//                 </span>,
//               ])}
//               childrenStyles={List([
//                 { flex: "auto", justifyContent: "flex-end" },
//                 { flex: "auto" },
//               ])}
//               className="padding-l"
//             />
//           </div>
//         ) : (
//           <div className="container">
//             <Flexbox
//               children={List([
//                 <span className="upload-item">
//                   <Flexbox
//                     children={List([
//                       <RiUploadCloudFill
//                         size="3rem"
//                         className="margin-r-m black-font"
//                       />,

//                       <span>
//                         <span className="title-m bold">
//                           {this.props.msg.pkg.get("browser.upload.title")}
//                         </span>
//                         <span className="desc-m grey0-font">
//                           {this.props.msg.pkg.get("browser.upload.desc")}
//                         </span>
//                       </span>,
//                     ])}
//                   />
//                 </span>,

//                 <span></span>,
//               ])}
//             />

//             {uploadingList}
//           </div>
//         )
//       ) : null;

//     const sharingList = this.props.browser.sharings.map((dirPath: string) => {
//       return (
//         <div id="share-list" key={dirPath}>
//           <Flexbox
//             children={List([
//               <Flexbox
//                 children={List([
//                   <RiFolderSharedFill
//                     size="3rem"
//                     className="purple0-font margin-r-m"
//                   />,
//                   <span>{dirPath}</span>,
//                 ])}
//               />,

//               <span>
//                 <input
//                   type="text"
//                   readOnly
//                   className="float-input"
//                   value={`${
//                     document.location.href.split("?")[0]
//                   }?dir=${encodeURIComponent(dirPath)}`}
//                 />
//                 <button
//                   onClick={() => {
//                     this.deleteSharing(dirPath);
//                   }}
//                   className="float-input"
//                 >
//                   {this.props.msg.pkg.get("browser.share.del")}
//                 </button>
//               </span>,
//             ])}
//             childrenStyles={List([{}, { justifyContent: "flex-end" }])}
//           />
//         </div>
//       );
//     });

//     const sharingListPane =
//       this.props.browser.tab === "sharing" ? (
//         this.props.browser.sharings.size === 0 ? (
//           <div className="container">
//             <Flexbox
//               children={List([
//                 <RiEmotionSadLine
//                   size="4rem"
//                   className="margin-r-m red0-font"
//                 />,
//                 <span>
//                   <h3 className="title-l">
//                     {this.props.msg.pkg.get("share.404.title")}
//                   </h3>
//                   <span className="desc-l grey0-font">
//                     {this.props.msg.pkg.get("share.404.desc")}
//                   </span>
//                 </span>,
//               ])}
//               childrenStyles={List([
//                 { flex: "auto", justifyContent: "flex-end" },
//                 { flex: "auto" },
//               ])}
//               className="padding-l"
//             />
//           </div>
//         ) : (
//           <div className="container">
//             <Flexbox
//               children={List([
//                 <span className="padding-m">
//                   <Flexbox
//                     children={List([
//                       <RiShareBoxLine
//                         size="3rem"
//                         className="margin-r-m black-font"
//                       />,

//                       <span>
//                         <span className="title-m bold">
//                           {this.props.msg.pkg.get("browser.share.title")}
//                         </span>
//                         <span className="desc-m grey0-font">
//                           {this.props.msg.pkg.get("browser.share.desc")}
//                         </span>
//                       </span>,
//                     ])}
//                   />
//                 </span>,

//                 <span></span>,
//               ])}
//             />

//             {sharingList}
//           </div>
//         )
//       ) : null;

//     const showTabs = this.props.login.userRole === roleVisitor ? "hidden" : "";
//     return (
//       <div>
//         <div id="browser">
//           <div className={`container ${showTabs}`}>
//             <div id="tabs">
//               <button
//                 onClick={() => {
//                   this.setTab("item");
//                 }}
//                 className="float"
//               >
//                 <Flexbox
//                   children={List([
//                     <RiFolder2Fill
//                       size="1.6rem"
//                       className="margin-r-s cyan0-font"
//                     />,
//                     <span>{this.props.msg.pkg.get("browser.item.title")}</span>,
//                   ])}
//                   childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
//                 />
//               </button>
//               <button
//                 onClick={() => {
//                   this.setTab("uploading");
//                 }}
//                 className="float"
//               >
//                 <Flexbox
//                   children={List([
//                     <RiUploadCloudFill
//                       size="1.6rem"
//                       className="margin-r-s blue0-font"
//                     />,
//                     <span>
//                       {this.props.msg.pkg.get("browser.upload.title")}
//                     </span>,
//                   ])}
//                   childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
//                 />
//               </button>
//               <button
//                 onClick={() => {
//                   this.setTab("sharing");
//                 }}
//                 className="float"
//               >
//                 <Flexbox
//                   children={List([
//                     <RiShareBoxLine
//                       size="1.6rem"
//                       className="margin-r-s purple0-font"
//                     />,
//                     <span>
//                       {this.props.msg.pkg.get("browser.share.title")}
//                     </span>,
//                   ])}
//                   childrenStyles={List([{ flex: "30%" }, { flex: "70%" }])}
//                 />
//               </button>
//             </div>
//           </div>

//           <div>{sharingListPane}</div>
//           <div>{uploadingListPane}</div>
//           {itemListPane}
//         </div>
//       </div>
//     );
//   }
// }
