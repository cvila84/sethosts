@echo off
rsrc -manifest sethosts.manifest -o ..\cmd\sethosts\sethosts.syso
pushd ..\cmd\sethosts
go install
popd
