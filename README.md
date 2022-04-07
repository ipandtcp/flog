Golang simple file log
================

### NOTE
This package will redirect stdout to info logfile and redirect stderr to err logfile when mode = release, All Debug log will ignore when mode != debug

### Example

```go
package main

import "github.com/ipandtcp/flog"

const (
	mode = "release"
)

func main() {
	var lg *flog.Log

	if mode != "release" {
		lg = flog.NewLog(flog.MODE_DEBUG)
	} else {
		lg		= flog.NewLog(flog.MODE_RELEASE)
		//lg.LogPath	= "/tmp/flog/"
		//lg.Umask	= 0660
		lg.FilePrefix	= "flog-example"

		// True: all level log to single file
		// False: every level log to each files

		//c.log.AllInOne = true
	}

	err := lg.Init()
	if err != nil {
	    panic(err)
	}

	lg.Debug("Hello world 1.")
	lg.Info("Hello world 2.")
}
```
