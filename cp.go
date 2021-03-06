package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/butler/comm"
	"github.com/itchio/wharf/counter"
	"github.com/itchio/wharf/eos"
)

func cp(src string, dest string, resume bool) {
	must(doCp(src, dest, resume))
}

func doCp(srcPath string, destPath string, resume bool) error {
	src, err := eos.Open(srcPath)
	if err != nil {
		return err
	}

	defer src.Close()

	dir := filepath.Dir(destPath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	flags := os.O_CREATE | os.O_WRONLY

	dest, err := os.OpenFile(destPath, flags, 0644)
	if err != nil {
		return err
	}

	defer dest.Close()

	stats, err := src.Stat()
	if err != nil {
		return err
	}

	totalBytes := int64(stats.Size())
	startOffset := int64(0)

	if resume {
		startOffset, err = dest.Seek(0, os.SEEK_END)
		if err != nil {
			return err
		}

		if startOffset == 0 {
			comm.Logf("Downloading %s", humanize.IBytes(uint64(totalBytes)))
		} else if startOffset > totalBytes {
			comm.Logf("Existing data too big (%s > %s), starting over", humanize.IBytes(uint64(startOffset)), humanize.IBytes(uint64(totalBytes)))
		} else if startOffset == totalBytes {
			comm.Logf("All %s already there", humanize.IBytes(uint64(totalBytes)))
			return nil
		}

		comm.Logf("Resuming at %s / %s", humanize.IBytes(uint64(startOffset)), humanize.IBytes(uint64(totalBytes)))

		_, err = src.Seek(startOffset, os.SEEK_SET)
		if err != nil {
			return err
		}
	} else {
		comm.Logf("Downloading %s", humanize.IBytes(uint64(totalBytes)))
	}

	start := time.Now()

	comm.Progress(float64(startOffset) / float64(totalBytes))
	comm.StartProgressWithTotalBytes(totalBytes)

	cw := counter.NewWriterCallback(func(count int64) {
		alpha := float64(startOffset+count) / float64(totalBytes)
		comm.Progress(alpha)
	}, dest)

	copiedBytes, err := io.Copy(cw, src)
	if err != nil {
		return err
	}
	comm.EndProgress()

	totalDuration := time.Since(start)
	prettyStartOffset := humanize.IBytes(uint64(startOffset))
	prettySize := humanize.IBytes(uint64(copiedBytes))
	perSecond := humanize.IBytes(uint64(float64(totalBytes-startOffset) / totalDuration.Seconds()))
	comm.Statf("%s + %s copied @ %s/s\n", prettyStartOffset, prettySize, perSecond)

	return nil
}
