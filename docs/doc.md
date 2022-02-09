### File Management
#### Resume Uploading
By clicking the upload button and re-upload the stopped file, the client will resume the uploading.

#### Move Files or Folders
You can move files or folders by following these steps:
Choose files or folders by ticking them (in the right)
Go to the target folder
Click the “Paste” button at the top of the pane.

#### Share Directories
You can share a folder and its files by following these steps:
Go to “Files” tab, 
Go to the folder you want to share
Click the “Share Folder” button

#### Cancel Sharings
There are 2 ways to cancel one sharing:
In the “Files” tab, go to the folder and click the “Stop Sharing” button
In the “Sharings” tab, find the target directory and click the “Cancel” button

#### Manage Files and Folders outside the Docker Container
If the Quickshare is started inside a docker, all files and folders are also persisted inside the docker. Then it is difficult to manage files and folders through the OS.
 
Here is a solution:
##### About Permissions
In the Quickshare docker image, a user `quickshare` (uid=8686) and group `quickshare` (gid=8686) are predefined. Normally in Linux, you can not manage files outside the docker, because your uid is not 8686 and you are not a member of `quickshare` group. By creating a `quickshare` group and adding yourself into it, you are able to manage files:
```
groupadd -g 8686 quickshare
usermod -aG quickshare $USER
```


##### Use [Bind Mounts](https://docs.docker.com/storage/bind-mounts/)
You can mount a non-empty directory with uid=8686 and gid=8686 in running the docker:
```
docker run \
--name quickshare \
-d -p 8686:8686 \
-u 8686:8686 \
-v `pwd`/non-empty-directory:/quickshare/root \
-e DEFAULTADMIN=qs \
-e DEFAULTADMINPWD=1234 \
hexxa/quickshare
```
Then you can find files and folders created by the Quickshare under `non-empty-directory`.
 
You can also start a container with a [volume](https://docs.docker.com/storage/volumes/), however it is not easy to manage from the OS in this way.
 
### User Management
#### Add Predefined Users
Predefined users can be added by the config file in the `users.predefinedUsers` array, for example, prepare a partial configuration file `predefined_users.yaml`:
```
users:
  predefinedUsers:◊
    - name: "user1"
      pwd: "Quicksh@re"
      role: "user"
    - name: "user2"
      pwd: "Quicksh@re"
      role: "user"
```
In the yaml, 2 users are predefined: `user1` and `user2` who are identified by password `Quicksh@re`.
Start the Quickshare by adding this configuration:
```
./quickshare -c predefined_users.yaml
```
Then you can see these users in the Settings > Management > Users.
 
### System Management
#### Customized Config
You are able to overwrite default configuration by providing your own configuration. 
For example, if you want to turn off the captcha, you can set `captchaEnabled` as `false` in your configuration or create a new configuration `disable_captcha.yaml`:
```
users:
  captchaEnabled: false
```
Then start the Quickshare by appending this configuration:
```
./quickshare -c disable_captcha.yaml
```
 
#### Background Customization
You can customize the background by following these steps:
Upload the wallpaper to some directory
Share this directory
Copy the link of the wallpaper
Go to `Settings > Preference` and set the Background URL in the Background Pane.
 
### MISC
