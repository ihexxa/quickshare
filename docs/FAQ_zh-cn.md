## 常见问题

- 怎么改用户名和密码?
  - 进入 quickshare 文件夹
  - 用文本编辑器打开 config.json. (比如 notpad++, sublime, vscode, 等等...)
  - 查找行 `"AdminPwd": "quicksh@re",`
  - 用你的密码替换 `quicksh@re`, 比如 `"AdminPwd": "你的密码",`
- 怎么改监听地址(端口)?
  - 进入 quickshare 文件夹
  - 用文本编辑器打开 config.json. (比如 notpad++, sublime, vscode, 等等...)
  - 查找行 `"HostName": "",`
  - 更改`HostName`的值, 比如 `"HostName": "192.168.0.6",`
  - 你可以用上面的方法更改端口值(`"Port": 8888,`)
- 怎么更改背景?
  - 进入 quickshare 文件夹下的`public`文件夹
  - 使用文本编辑器打开 style.css. (比如 notpad++, sublime, vscode, 等等...)
  - 更新 `body`'的 css
