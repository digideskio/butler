package main

import "github.com/itchio/butler/comm"
import "github.com/itchio/wharf/archiver"

func untar(file string, dir string) {
	settings := archiver.ExtractSettings{
		Consumer: comm.NewStateConsumer(),
	}

	comm.StartProgress()
	res, err := archiver.ExtractTar(file, dir, settings)
	comm.EndProgress()

	must(err)
	comm.Logf("Extracted %d dirs, %d files, %d symlinks", res.Dirs, res.Files, res.Symlinks)
}
