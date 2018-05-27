## Configuration

```Javascript
{
  "AppName": "qs",
  "AdminId": "admin", // login user name
  "AdminPwd": "quicksh@re", // login password
  "SecretKey": "qs", // key for hashing cookie (jwt)
  "Production": true,
  "HostName": "", // listening address
  "Port": 8888, // listening port
  "MaxUpBytesPerSec": 2000000, // upload speed limit
  "MaxDownBytesPerSec": 1000000, // download speed limit
  "MaxRangeLength": 10485760, // max length of chunk to upload at once
  "Timeout": 7200000, // connection timeout
  "ReadTimeout": 5000, // connection read request timeout
  "WriteTimeout": 7200000, // connection write response timeout
  "IdleTimeout": 10000, // connection idle timeout
  "WorkerPoolSize": 16, // number of workers, it decides how many download connections are provided at same time
  "TaskQueueSize": 16, // how many requests can be queued
  "QueueSize": 16,
  "ParseFormBufSize": 5000000, // buffer for parsing request
  "MaxHeaderBytes": 1024, // max header size in byte
  "DownLimit": -1, // default download limit
  "MaxShares": 16384, // max number of sharing
  "LocalFileLimit": -1, // max number of listing file at once
  "CookieDomain": "",
  "CookieHttpOnly": false,
  "CookieMaxAge": 604800,
  "CookiePath": "/",
  "CookieSecure": false,
  "KeyAdminId": "adminid",
  "KeyAdminPwd": "adminpwd",
  "KeyToken": "token",
  "KeyFileName": "fname",
  "KeyFileSize": "size",
  "KeyShareId": "shareid",
  "KeyStart": "start",
  "KeyLen": "len",
  "KeyChunk": "chunk",
  "KeyAct": "act",
  "KeyExpires": "expires",
  "KeyDownLimit": "downlimit",
  "ActStartUpload": "startupload",
  "ActUpload": "upload",
  "ActFinishUpload": "finishupload",
  "ActLogin": "login",
  "ActLogout": "logout",
  "ActShadowId": "shadowid",
  "ActPublishId": "publishid",
  "ActSetDownLimit": "setdownlimit",
  "ActAddLocalFiles": "addlocalfiles",
  "AllUsers": "addlocalfiles",
  "OpIdIpVisit": 0,
  "OpIdUpload": 1,
  "OpIdDownload": 2,
  "OpIdLogin": 3,
  "OpIdGetFInfo": 4,
  "OpIdDelFInfo": 5,
  "OpIdOpFInfo": 6,
  "PathLocal": "files",
  "PathLogin": "/login",
  "PathDownloadLogin": "/download-login",
  "PathDownload": "/download",
  "PathUpload": "/upload",
  "PathStartUpload": "/startupload",
  "PathFinishUpload": "/finishupload",
  "PathFileInfo": "/fileinfo",
  "PathClient": "/",
  "LimiterCap": 256,
  "LimiterTtl": 3600,
  "LimiterCyc": 1,
  "BucketCap": 10, // operation is allowed at most 10 times per second, but SpecialCapsStr will override this value
  "SpecialCapsStr": {
    "0": 30, // IpVisit is allowed at most 30 times per second
    "1": 10, // Uploading is allowed at most 10 times per second
    "2": 10, // Downloading is allowed at most 10 times per second
    "3": 1 // Login/Logout is allowed at most 1 time per second
    // You can also add rate limits according to OpIdxxx above.
  }
}
```
