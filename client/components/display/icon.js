const IconFile = require("react-icons/lib/fa/file-o");
const IconImg = require("react-icons/lib/md/image");
const IconZip = require("react-icons/lib/md/archive");
const IconVideo = require("react-icons/lib/md/ondemand-video");
const IconAudio = require("react-icons/lib/md/music-video");
const IconText = require("react-icons/lib/md/description");
const IconExcel = require("react-icons/lib/fa/file-excel-o");
const IconPPT = require("react-icons/lib/fa/file-powerpoint-o");
const IconPdf = require("react-icons/lib/md/picture-as-pdf");
const IconWord = require("react-icons/lib/fa/file-word-o");
const IconCode = require("react-icons/lib/md/code");
const IconApk = require("react-icons/lib/md/android");
const IconExe = require("react-icons/lib/fa/cog");

const IconBars = require("react-icons/lib/fa/bars");
const IconSpinner = require("react-icons/lib/md/autorenew");
const IconCirUp = require("react-icons/lib/fa/arrow-circle-up");
const IconSignIn = require("react-icons/lib/fa/sign-in");
const IconSignOut = require("react-icons/lib/fa/sign-out");
const IconAngUp = require("react-icons/lib/fa/angle-up");
const IconAngRight = require("react-icons/lib/fa/angle-right");
const IconAngDown = require("react-icons/lib/fa/angle-down");
const IconAngLeft = require("react-icons/lib/fa/angle-left");
const IconTimesCir = require("react-icons/lib/md/cancel");
const IconPlusSqu = require("react-icons/lib/md/add-box");
const IconPlusCir = require("react-icons/lib/fa/plus-circle");
const IconPlus = require("react-icons/lib/md/add");
const IconSearch = require("react-icons/lib/fa/search");
const IconThList = require("react-icons/lib/fa/th-list");
const IconCalendar = require("react-icons/lib/fa/calendar-o");

const IconCheckCir = require("react-icons/lib/fa/check-circle");
const IconExTri = require("react-icons/lib/fa/exclamation-triangle");
const IconInfoCir = require("react-icons/lib/fa/info-circle");
const IconRefresh = require("react-icons/lib/fa/refresh");

const fileTypeIconMap = {
  // text
  txt: { icon: IconText, color: "#333" },
  rtf: { icon: IconText, color: "#333" },
  htm: { icon: IconText, color: "#333" },
  html: { icon: IconText, color: "#333" },
  xml: { icon: IconText, color: "#333" },
  yml: { icon: IconText, color: "#333" },
  json: { icon: IconText, color: "#333" },
  toml: { icon: IconText, color: "#333" },
  md: { icon: IconText, color: "#333" },
  // office
  ppt: { icon: IconPPT, color: "#e67e22" },
  pptx: { icon: IconPPT, color: "#e67e22" },
  xls: { icon: IconExcel, color: "#16a085" },
  xlsx: { icon: IconExcel, color: "#16a085" },
  xlsm: { icon: IconExcel, color: "#16a085" },
  doc: { icon: IconWord, color: "#2980b9" },
  docx: { icon: IconWord, color: "#2980b9" },
  docx: { icon: IconWord, color: "#2980b9" },
  pdf: { icon: IconPdf, color: "#c0392b" },
  // code
  c: { icon: IconCode, color: "#666" },
  cpp: { icon: IconCode, color: "#666" },
  java: { icon: IconCode, color: "#666" },
  js: { icon: IconCode, color: "#666" },
  py: { icon: IconCode, color: "#666" },
  pyc: { icon: IconCode, color: "#666" },
  rb: { icon: IconCode, color: "#666" },
  php: { icon: IconCode, color: "#666" },
  go: { icon: IconCode, color: "#666" },
  sh: { icon: IconCode, color: "#666" },
  vb: { icon: IconCode, color: "#666" },
  sql: { icon: IconCode, color: "#666" },
  r: { icon: IconCode, color: "#666" },
  swift: { icon: IconCode, color: "#666" },
  oc: { icon: IconCode, color: "#666" },
  // misc
  apk: { icon: IconApk, color: "#2ecc71" },
  exe: { icon: IconExe, color: "#333" },
  deb: { icon: IconExe, color: "#333" },
  rpm: { icon: IconExe, color: "#333" },
  // img
  bmp: { icon: IconImg, color: "#1abc9c" },
  gif: { icon: IconImg, color: "#1abc9c" },
  jpg: { icon: IconImg, color: "#1abc9c" },
  jpeg: { icon: IconImg, color: "#1abc9c" },
  tiff: { icon: IconImg, color: "#1abc9c" },
  psd: { icon: IconImg, color: "#1abc9c" },
  png: { icon: IconImg, color: "#1abc9c" },
  svg: { icon: IconImg, color: "#1abc9c" },
  pcx: { icon: IconImg, color: "#1abc9c" },
  dxf: { icon: IconImg, color: "#1abc9c" },
  wmf: { icon: IconImg, color: "#1abc9c" },
  emf: { icon: IconImg, color: "#1abc9c" },
  eps: { icon: IconImg, color: "#1abc9c" },
  tga: { icon: IconImg, color: "#1abc9c" },
  // compress
  gz: { icon: IconZip, color: "#34495e" },
  zip: { icon: IconZip, color: "#34495e" },
  "7z": { icon: IconZip, color: "#34495e" },
  rar: { icon: IconZip, color: "#34495e" },
  tar: { icon: IconZip, color: "#34495e" },
  gzip: { icon: IconZip, color: "#34495e" },
  cab: { icon: IconZip, color: "#34495e" },
  uue: { icon: IconZip, color: "#34495e" },
  arj: { icon: IconZip, color: "#34495e" },
  bz2: { icon: IconZip, color: "#34495e" },
  lzh: { icon: IconZip, color: "#34495e" },
  jar: { icon: IconZip, color: "#34495e" },
  ace: { icon: IconZip, color: "#34495e" },
  iso: { icon: IconZip, color: "#34495e" },
  z: { icon: IconZip, color: "#34495e" },
  // video
  asf: { icon: IconVideo, color: "#f39c12" },
  avi: { icon: IconVideo, color: "#f39c12" },
  flv: { icon: IconVideo, color: "#f39c12" },
  mkv: { icon: IconVideo, color: "#f39c12" },
  mov: { icon: IconVideo, color: "#f39c12" },
  mp4: { icon: IconVideo, color: "#f39c12" },
  mpeg: { icon: IconVideo, color: "#f39c12" },
  mpg: { icon: IconVideo, color: "#f39c12" },
  ram: { icon: IconVideo, color: "#f39c12" },
  rmvb: { icon: IconVideo, color: "#f39c12" },
  qt: { icon: IconVideo, color: "#f39c12" },
  wmv: { icon: IconVideo, color: "#f39c12" },
  // audio
  cda: { icon: IconAudio, color: "#d35400" },
  cmf: { icon: IconAudio, color: "#d35400" },
  mid: { icon: IconAudio, color: "#d35400" },
  mp1: { icon: IconAudio, color: "#d35400" },
  mp2: { icon: IconAudio, color: "#d35400" },
  mp3: { icon: IconAudio, color: "#d35400" },
  rm: { icon: IconAudio, color: "#d35400" },
  rmi: { icon: IconAudio, color: "#d35400" },
  vqf: { icon: IconAudio, color: "#d35400" },
  wav: { icon: IconAudio, color: "#d35400" }
};

const fileIconMap = {
  ...fileTypeIconMap,
  // other
  spinner: { icon: IconSpinner, color: "#1abc9c" },
  cirup: { icon: IconCirUp, color: "#fff" },
  signin: { icon: IconSignIn, color: "#fff" },
  signout: { icon: IconSignOut, color: "#fff" },
  angup: { icon: IconAngUp, color: "#2c3e50" },
  angright: { icon: IconAngRight, color: "#2c3e50" },
  angdown: { icon: IconAngDown, color: "#2c3e50" },
  angleft: { icon: IconAngLeft, color: "#2c3e50" },
  timescir: { icon: IconTimesCir, color: "#c0392b" },
  plussqu: { icon: IconPlusSqu, color: "#2ecc71" },
  pluscir: { icon: IconPlusCir, color: "#2ecc71" },
  plus: { icon: IconPlus, color: "#2ecc71" },
  search: { icon: IconSearch, color: "#ccc" },
  checkcir: { icon: IconCheckCir, color: "#27ae60" },
  extri: { icon: IconExTri, color: "#f39c12" },
  infocir: { icon: IconInfoCir, color: "#2c3e50" },
  refresh: { icon: IconRefresh, color: "#8e44ad" },
  thlist: { icon: IconThList, color: "#fff" },
  bars: { icon: IconBars, color: "#666" },
  calendar: { icon: IconCalendar, color: "#333" }
};

export const getIcon = extend => {
  if (fileIconMap[extend.toUpperCase()]) {
    return fileIconMap[extend.toUpperCase()].icon;
  } else if (fileIconMap[extend.toLowerCase()]) {
    return fileIconMap[extend.toLowerCase()].icon;
  }
  return IconFile;
};

export const getIconColor = extend => {
  if (fileIconMap[extend.toUpperCase()]) {
    return fileIconMap[extend.toUpperCase()].color;
  } else if (fileIconMap[extend.toLowerCase()]) {
    return fileIconMap[extend.toLowerCase()].color;
  }
  return "#333";
};
