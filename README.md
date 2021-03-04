<h1 align="center">
  Quickshare
</h1>
<p align="center">
  Quick and simple file sharing between different devices.
</p>
<p align="center">
  <a href="https://github.com/ihexxa/quickshare/actions">
    <img src="https://github.com/ihexxa/quickshare/workflows/quickshare-ci/badge.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/ihexxa/quickshare">
    <img src="https://goreportcard.com/badge/github.com/ihexxa/quickshare" />
  </a>
  <a href="https://gitter.im/quickshare/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link">
    <img src="https://badges.gitter.im/Join%20Chat.svg" />
  </a>
<p>

![Quickshare on desktop](./docs/imgs/desktop.jpeg)

![Quickshare on mobile](./docs/imgs/mobile.jpeg)

Choose Language: English | [简体中文](./docs/README_zh-cn.md)

## Main Features

- Sharing and accessing from different devices (Adaptive UI)
- Be compatible with Linux, Mac and Windows
- Stopping and resuming uploading/downloading
- Do uploading and downloading in web browser
- Manage files through browser or OS

## Quick Start

### Run in Docker

Following will start a `quickshare` docker and listen to `8686` port.

Then you can open `http://127.0.0.1:8686` and log in with user name `qs` and password `1234`:

```
docker run \
    --name quickshare \
    -d -p 8686:8686 \
    -v `pwd`/quickshare/root:/quickshare/root \
    -e DEFAULTADMIN=qs \
    -e DEFAULTADMINPWD=1234 \
    hexxa/quickshare
```

- `DEFAULTADMIN` is the default user name
- `DEFAULTADMINPWD` is the default user password
- `/quickshare/root` is where Quickshare stores files and directories.

### Run from source code

Before start, please confirm Go/Golang (>1.15), Node.js and Yarn are installed on your machine.

```
# clone this repo
git clone git@github.com:ihexxa/quickshare.git

# go to repo's folder
cd quickshare

DEFAULTADMIN=qs DEFAULTADMINPWD=1234 yarn start
```

OK! Open `http://127.0.0.1:8686` in browser, and log in with user name `qs` and password `1234`.

### Run executable file

- **Downloading**: Download last distribution(s) in [Release Page](https://github.com/ihexxa/quickshare/releases).
- **Unzipping**: Unzip it and run following command `DEFAULTADMIN=qs DEFAULTADMINPWD=1234 ./quickshare`. (You may update its execution permission: e.g. run `chmod u+x quickshare`)
- **Accessing**: At last, open `http://127.0.0.1:8686` in browser, and log in with user name `qs` and password `1234`.

### FAQ

Coming soon.
