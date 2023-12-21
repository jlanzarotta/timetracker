@echo off
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
set BUILD_VERSION=1.0.4
set BUILD_DATETIME="%date:~10,4%-%date:~4,2%-%date:~7,2%T%time: =0%"
go build -ldflags="-s -w -X 'timetracker/cmd.BuildDateTime=%BUILD_DATETIME%' -X 'timetracker/cmd.BuildVersion=%BUILD_VERSION%'" -work -x -v -o tt.exe main.go
