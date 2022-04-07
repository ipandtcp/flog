/*
 * Copyright (C) 2018  kevin.zhu. All rights reserved
 *
 * Direct questions, comments to <ipandtcp@gmail.com>
 *
 * NOTE: This package will redirect stdout to info logfile and redirect stderr to err logfil when mode = release
 *       All Debug log will igore when mode != debug
 */

package flog

import (
	"fmt"
	"io"
	"os"
	"time"
	"errors"
	"runtime"
	"strings"
)

type Log struct {
	LogPath         string
	filesCreateTime time.Time
	infoFile        io.Writer
	warnFile        io.Writer
	errFile         io.Writer
	FilePrefix	string	// log file name prefix, FIXME: check valid
	iFileName	string	// info file name
	wFileName	string	// warning file name
	eFileName	string	// err file name
	AllInOne	bool	// all level logs in one file
	NewFileInterval int	// day
	Umask		uint32
	Mode	        Mode
}

type Mode int

const (
	MODE_RELEASE Mode  = 0
	MODE_DEBUG   Mode  = 1
)

var StdOut *Log

func init() {
	StdOut = NewLog(MODE_DEBUG)
	StdOut.Init()
}

func NewLog(mode Mode) *Log {
	return &Log{Mode : mode}
}

func (l *Log) Init() error {
	if l.Mode != MODE_RELEASE {
		l.errFile = os.Stderr
		l.infoFile = os.Stderr
		l.warnFile = os.Stderr
		return nil
	}
	if l.LogPath == "" || l.Umask == 0 || l.NewFileInterval == 0 {
		return errors.New("logPath or Umask or NewFileInterval value error")
	}

	if l.LogPath[len(l.LogPath)-1] == '/' {
		l.LogPath = l.LogPath[:len(l.LogPath)-1]
	}

	_, err := os.Stat(l.LogPath)
	if err != nil {
		err := os.MkdirAll(l.LogPath, os.FileMode(l.Umask))
		if err != nil {
			return err
		}
	}

	return nil
}

func fileExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		return false
	}
	return true
}

// Check if need create new log file
// is depend on l.NewFileInterval, unit is day
// Every time restart , will create a new log file or open if today file exist
func (l *Log) writerCheck() {
	if l.Mode != MODE_RELEASE {
		return 
	}
	now := time.Now()

	if l.AllInOne {
		// if file not need create new
		if now.Before(l.filesCreateTime.AddDate(0, 0, l.NewFileInterval)) && fileExist(l.iFileName) {
			return
		}

		name := now.Format("2006-01-02")
		if l.FilePrefix != "" {
			name = l.FilePrefix + "-" + name
		}

		ifileName := fmt.Sprintf("%s/%s.log", l.LogPath, name)

		ifile, err := os.OpenFile(ifileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.FileMode(l.Umask))
		if err != nil {
			panic(err)
		} else {
			if l.infoFile != nil {
				err := l.infoFile.(*os.File).Close()
				if err != nil {
					io.WriteString(ifile, "Close expired log file err:"+err.Error()+"\n")
				}
			}
			l.infoFile  = ifile
			l.warnFile  = ifile
			l.errFile   = ifile
			l.iFileName = ifileName
			l.wFileName = ifileName
			l.eFileName = ifileName
			os.Stdout   = ifile
			os.Stderr   = ifile
		}
	} else {
		// if file not need create new
		if now.Before(l.filesCreateTime.AddDate(0, 0, l.NewFileInterval)) && 
			fileExist(l.iFileName) && fileExist(l.wFileName) && fileExist(l.eFileName) {
			return
		}

		ifileName := ""
		wfileName := ""
		efileName := ""
		name      := now.Format("2006-01-02")

		if l.FilePrefix != "" {
			ifileName = fmt.Sprintf("%s/%s-info-%s.log", l.LogPath, l.FilePrefix, name)
			wfileName = fmt.Sprintf("%s/%s-warn-%s.log", l.LogPath, l.FilePrefix, name)
			efileName = fmt.Sprintf("%s/%s-err-%s.log", l.LogPath, l.FilePrefix, name)
		} else {
			ifileName = fmt.Sprintf("%s/info-%s.log", l.LogPath, name)
			wfileName = fmt.Sprintf("%s/warn-%s.log", l.LogPath, name)
			efileName = fmt.Sprintf("%s/err-%s.log", l.LogPath, name)
		}

		ifile, err := os.OpenFile(ifileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.FileMode(l.Umask))
		if ifile != nil {
			if l.infoFile != nil {
				err := l.infoFile.(*os.File).Close()
				if err != nil {
					io.WriteString(ifile, "Close expired log file err:"+err.Error()+"\n")
				}
			}
			l.infoFile = ifile
			l.iFileName = ifileName
			os.Stdout = ifile
		} else {
			panic(err)
		}

		wfile, err := os.OpenFile(wfileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.FileMode(l.Umask))
		if wfile != nil {
			if l.warnFile != nil {
				err := l.warnFile.(*os.File).Close()
				if err != nil {
					io.WriteString(ifile, "Close expired log file err:"+err.Error()+"\n")
				}
			}
			l.warnFile = wfile
			l.wFileName = wfileName
		} else {
			panic(err)
		}

		efile, err := os.OpenFile(efileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.FileMode(l.Umask))
		if efile != nil {
			if l.errFile != nil {
				err := l.errFile.(*os.File).Close()
				if err != nil {
					io.WriteString(ifile, "Close expired log file err:"+err.Error()+"\n")
				}
			}
			l.errFile = efile
			l.eFileName = efileName
			os.Stderr  = efile
		} else {
			panic(err)
		}
	}

	// truncate to 0 clock
	l.filesCreateTime = now.Truncate(time.Duration(now.Hour()) * time.Hour + 
				time.Duration(now.Minute()) * time.Minute + 
				time.Duration(now.Second()) * time.Second)
}

func (l *Log) Debug(s string, args ...interface{}) {
	if l.Mode != MODE_DEBUG {
		return
	}

	head := "[DEBUG]   | " + time.Now().Format("06/01/02 - 15:04:05.000 | ")
	l.outPut(l.infoFile, head,  fmt.Sprintf(s, args...))
}

func (l *Log) Info(s string, args ...interface{}) {
	l.writerCheck()
	head := "[INFO]    | " + time.Now().Format("06/01/02 - 15:04:05.000 | ") 
	l.outPut(l.infoFile, head, fmt.Sprintf(s, args...))
}

func (l *Log) Warning(s string, args ...interface{}) {
	l.writerCheck()
	head := "[WARNING] | " + time.Now().Format("06/01/02 - 15:04:05.000 | ") 
	l.outPut(l.warnFile, head, fmt.Sprintf(s, args...))
}

func (l *Log) Error(s string, args ...interface{}) {
	l.writerCheck()
	head := "[ERROR]   | " + time.Now().Format("06/01/02 - 15:04:05.000 | ")
	l.outPut(l.errFile, head, fmt.Sprintf(s, args...))
}

// Just like println to print string to info file
// If AllInOne is set, all log in one file
func (l *Log) PrintlnInfo(s string, args ...interface{}) {
	l.writerCheck()
	io.WriteString(l.infoFile, fmt.Sprintf(s+"\n", args...))
}

// Just like println to print string to warning file 
// If AllInOne is set, all log in one file
func (l *Log) PrintlnWarning(s string, args ...interface{}) {
	l.writerCheck()
	io.WriteString(l.warnFile, fmt.Sprintf(s+"\n", args...))
}

// Just like println to print string to err file 
// If AllInOne is set, all log in one file
func (l *Log) PrintlnErr(s string, args ...interface{}) {
	l.writerCheck()
	io.WriteString(l.errFile, fmt.Sprintf(s+"\n", args...))
}

func (l *Log) outPut(d io.Writer, head, s string) {
	var funcName string
	pc := make([]uintptr, 1)
	pcNum := runtime.Callers(3, pc)
	if pcNum == 0 {
		funcName = "UnknownFuncName"
	} else {
		caller := runtime.FuncForPC(pc[0])
		if caller ==  nil {
			funcName = "UnknownFuncName"
		} else {
			str	:= caller.Name()
			strs	:= strings.Split(str, ".")
			funcName = strs[len(strs)-1]
		}
	}

	io.WriteString(d, head + "[" + funcName + "] " + s + string('\n'))
}
