module github.com/ihexxa/quickshare/v2

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/ihexxa/gocfg v0.0.0-00010101000000-000000000000
	github.com/ihexxa/quickshare v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robbert229/jwt v2.0.0+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/skratchdot/open-golang v0.0.0-20160302144031-75fb7ed4208c
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

# clone the repo manually, replace this to your local repo dir
replace github.com/ihexxa/gocfg => D:\golang\src\github.com\ihexxa\gocfg

# clone the repo manually, reeplace this to your local repo dir
replace github.com/ihexxa/quickshare => D:\golang\src\github.com\ihexxa\quickshare
