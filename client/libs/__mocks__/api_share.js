let _infos = [];
const shadowedId = "shadowedId";
const publicId = "publicId";

export function __addInfos(infos) {
  _infos = [..._infos, ...infos];
}

export function __truncInfos(info) {
  _infos = [];
}

export const del = shareId => {
  _infos = _infos.filter(info => {
    return !info.shareId == shareId;
  });
  return Promise.resolve(true);
};

export const list = () => {
  return Promise.resolve(_infos);
};

export const shadowId = shareId => {
  return Promise.resolve(shadowedId);
};

export const publishId = shareId => {
  return Promise.resolve(publicId);
};

export const setDownLimit = (shareId, downLimit) => {
  _infos = _infos.map(info => {
    return info.shareId == shareId ? { ...info, downLimit } : info;
  });
  return Promise.resolve(true);
};

export const addLocalFiles = () => {
  return Promise.resolve(true);
};
