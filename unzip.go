package main

import (
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/butler/comm"
	"github.com/itchio/wharf/archiver"
	"github.com/itchio/wharf/eos"
)

func unzip(file string, dir string, resumeFile string, dryRun bool, concurrency int) {
	comm.Opf("Extracting zip %s to %s", eos.Redact(file), dir)

	var zipUncompressedSize int64

	settings := archiver.ExtractSettings{
		Consumer:   comm.NewStateConsumer(),
		ResumeFrom: resumeFile,
		OnUncompressedSizeKnown: func(uncompressedSize int64) {
			zipUncompressedSize = uncompressedSize
			comm.StartProgressWithTotalBytes(uncompressedSize)
		},
		DryRun:      dryRun,
		Concurrency: concurrency,
	}

	startTime := time.Now()

	res, err := archiver.ExtractPath(file, dir, settings)
	comm.EndProgress()

	duration := time.Since(startTime)
	bytesPerSec := float64(zipUncompressedSize) / duration.Seconds()

	must(err)
	comm.Logf("Extracted %d dirs, %d files, %d symlinks, %s at %s/s", res.Dirs, res.Files, res.Symlinks,
		humanize.IBytes(uint64(zipUncompressedSize)), humanize.IBytes(uint64(bytesPerSec)))
}
