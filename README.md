# sethosts
[![Build Status](https://travis-ci.com/cvila84/sethosts.svg?branch=master)](https://travis-ci.com/cvila84/sethosts)

## What is sethosts ?

sethosts allows to set the Windows hosts file directly from CLI as JSON payload

For example,

```shell
C:\>sethosts [{\"IP\":\"127.0.0.1\",\"HostName\":\"localhost\"}]  
```

will result in the following hosts file

```shell
127.0.0.1   localhost  
```
