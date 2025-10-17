@echo off

start "Go Massmail"      cmd.exe /K "cd /d C:\goprojects\go-massmail      && color 2F"
start "FTP Uploader"     cmd.exe /K "cd /d C:\goprojects\ftp-uploader     && color 4F"
start "Go Questionnaire" cmd.exe /K "cd /d C:\goprojects\go-questionnaire && color 6F"



@REM start wt.exe -w 0 nt -p "Command Prompt" -d "C:\goprojects\go-questionnaire" cmd /K "color 6F"
@REM start wt.exe -w 0 nt -p "Command Prompt" -d "C:\goprojects\go-massmail"      cmd /K "color 2F"


exit
