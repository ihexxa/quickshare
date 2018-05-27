const fileTypeMap = {
  jpg: "image",
  jpeg: "image",
  png: "image",
  bmp: "image",
  gz: "archive",
  mov: "video",
  mp4: "video",
  mov: "video",
  avi: "video"
};

export const getFileExt = fileName => fileName.split(".").pop();

export const getFileType = fileName => {
  const ext = getFileExt(fileName);
  return fileTypeMap[ext] != null ? fileTypeMap[ext] : "file";
};
