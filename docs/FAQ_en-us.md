## FAQ

* How to change accound name and password?
  * Go to quickshare folder
  * Open config.json using text editor. (e.g. notpad++, sublime, vscode, etc...)
  * Search for line `"AdminPwd": "quicksh@re",`
  * Replace `quicksh@re` with your password, e.g. `"AdminPwd": "myPassword",`
  * Then you can also update user name `"AdminId": "myUserName",` in above way.
* How to change listening address(or port)?
  * Go to quickshare folder
  * Open config.json using text editor. (e.g. notpad++, sublime, vscode, etc...)
  * Search for line `"HostName": "",`
  * Change the value of `HostName`, e.g. `"HostName": "192.168.0.6",`
  * You can also change port value (`"Port": 8888,`) in above way
* How to change background?
  * Go to `public` folder under quickshare folder
  * Open style.css using text editor. (e.g. notpad++, sublime, vscode, etc...)
  * Update `body`'s css.
