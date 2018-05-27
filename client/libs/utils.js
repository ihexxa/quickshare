export function makePostBody(paramMap) {
  return Object.keys(paramMap)
    .map(key => `${key}=${paramMap[key]}`)
    .join("&");
}
