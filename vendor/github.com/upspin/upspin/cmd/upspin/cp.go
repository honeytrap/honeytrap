// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

func (s *State) cp(args ...string) {
	const help = `
Cp copies files into, out of, and within Upspin. If the final
argument is a directory, the files are placed inside it.  The other
arguments must not be directories unless the -R flag is set.

If the final argument is not a directory, cp requires exactly two
path names and copies the contents of the first to the second.
The -R flag requires that the final argument be a directory.

When copying from one Upspin path to another Upspin path, cp can be
very efficient, copying only the references to the data rather than
the data itself.
`
	fs := flag.NewFlagSet("cp", flag.ExitOnError)
	fs.Bool("v", false, "log each file as it is copied")
	fs.Bool("R", false, "recursively copy directories")
	s.parseFlags(fs, args, help, "cp [opts] file... file or cp [opts] file... directory")

	cs := &copyState{
		state:   s,
		flagSet: fs,
		recur:   boolFlag(fs, "R"),
		verbose: boolFlag(fs, "v"),
	}

	// Do all the glob processing here.
	// Special one-at-time glob processing because each item may be local or Upspin.
	var files []cpFile
	for _, file := range fs.Args() {
		files = append(files, cs.glob(file)...)
	}

	if len(files) < 2 {
		fs.Usage()
	}

	nSrc := len(files) - 1
	src, dest := files[:nSrc], files[nSrc]
	s.copyCommand(cs, src, dest)
}

type copyState struct {
	state   *State
	flagSet *flag.FlagSet // Used only to call Usage.
	verbose bool
	recur   bool
}

func (c *copyState) logf(format string, args ...interface{}) {
	if c.verbose {
		log.Printf(format, args...)
	}
}

// A cpFile is a glob-expanded file name and an indication of whether
// it resides on Upspin.
type cpFile struct {
	path     string
	isUpspin bool
}

var (
	errExist    = errors.E(errors.Exist)
	errNotExist = errors.E(errors.NotExist)
	errIsDir    = errors.E(errors.IsDir)
)

func (s *State) copyCommand(cs *copyState, srcFiles []cpFile, dstFile cpFile) {
	// TODO: Check for nugatory copies.
	if s.isDir(dstFile) {
		s.copyToDir(cs, srcFiles, dstFile)
		return
	}
	if len(srcFiles) != 1 {
		s.failf("copying multiple files but %s is not a directory", dstFile.path)
		cs.flagSet.Usage()
	}
	if cs.recur {
		s.failf("recursive copy requires that final argument (%s) be an existing directory", dstFile.path)
		cs.flagSet.Usage()
	}
	reader, err := s.open(srcFiles[0])
	if err != nil {
		s.exit(err)
	}
	s.copyToFile(cs, reader, srcFiles[0], dstFile)
}

// isDir reports whether the file is a directory either in Upspin
// or in the local file system.
func (s *State) isDir(cf cpFile) bool {
	if cf.isUpspin {
		entry, err := s.client.Lookup(upspin.PathName(cf.path), true)
		// Report the error here if it's anything odd, because otherwise
		// we'll report "not a directory" misleadingly.
		if err != nil && !errors.Match(errNotExist, err) {
			log.Printf("%q: %v", cf.path, err)
		}
		return err == nil && entry.IsDir()
	}
	// Not an Upspin name. Is it a local directory?
	info, err := os.Stat(cf.path)
	return err == nil && info.IsDir()
}

// open opens the file regardless of its location.
func (s *State) open(file cpFile) (io.ReadCloser, error) {
	if s.isDir(file) {
		return nil, errors.E(upspin.PathName(file.path), errors.IsDir)
	}
	if file.isUpspin {
		return s.client.Open(upspin.PathName(file.path))
	}
	return os.Open(file.path)
}

// create creates the file regardless of its location.
func (s *State) create(file cpFile) (io.WriteCloser, error) {
	if file.isUpspin {
		fd, err := s.client.Create(upspin.PathName(file.path))
		return fd, err
	}
	fd, err := os.Create(file.path)
	return fd, err
}

// copyToDir copies the source files to the destination directory.
// It recurs if -R is set and a source is a subdirectory.
func (s *State) copyToDir(cs *copyState, src []cpFile, dir cpFile) {
	for _, from := range src {
		dstPath := path.Join(upspin.PathName(dir.path), filepath.Base(from.path))
		if dir.isUpspin && from.isUpspin {
			// Try a fast copy. It can fail but that's OK.
			cs.logf("try fast copy to %s", dstPath)
			if s.fastCopy(upspin.PathName(from.path), dstPath) == nil {
				continue
			}
		}
		reader, err := s.open(from)
		if cs.recur && errors.Match(errIsDir, err) {
			// If the problem is that from is a directory but we have -R,
			// recur on the contents.
			cs.logf("recursive descent into %s", from.path)
			newFiles, err := s.contents(cs, from)
			if len(newFiles) == 0 && err != nil {
				continue
			}
			// May need to make subdirectory (even if it will have no files).
			subDir := dir
			if dir.isUpspin {
				// Rather than use the libraries and a lot of casting, it's easiest just to cat the strings here.
				subDir.path = subDir.path + "/" + filepath.Base(from.path) // TODO: is filepath.Base OK?
				_, err := s.client.MakeDirectory(upspin.PathName(subDir.path))
				if err != nil && !errors.Match(errExist, err) {
					s.fail(err)
					continue
				}
			} else {
				subDir.path = filepath.Join(subDir.path, filepath.Base(from.path))
				err := os.Mkdir(subDir.path, 0755) // TODO: Mode.
				if err != nil && !os.IsExist(err) {
					s.fail(err)
					continue
				}
			}
			s.copyToDir(cs, newFiles, subDir)
			continue
		}
		if err != nil {
			s.fail(err)
			continue
		}
		dst := cpFile{
			path:     string(dstPath),
			isUpspin: dir.isUpspin,
		}
		s.copyToFile(cs, reader, from, dst)
	}
}

// copyToFile copies the source to the destination. The source file has already been opened.
func (s *State) copyToFile(cs *copyState, reader io.ReadCloser, src, dst cpFile) {
	cs.logf("start cp %s %s", src.path, dst.path)
	defer cs.logf("end cp %s %s", src.path, dst.path)
	// If both are in Upspin, we can avoid touching the data by copying
	// just the references.
	if src.isUpspin && dst.isUpspin {
		cs.logf("try fast copy to %v", dst)
		err := s.fastCopy(upspin.PathName(src.path), upspin.PathName(dst.path))
		if err == nil {
			return
		}
	}
	writer, err := s.create(dst)
	if err != nil {
		s.fail(err)
		reader.Close()
		return
	}
	cs.doCopy(reader, writer)
}

// fastCopy copies the source to the destination using the references rather than the data.
// If it fails, PutDuplicate failed because the file exists or the source is a directory.
// (Any other error is unexpected and exits the copy command.)
// The caller may be able to retry with a regular copy.
func (s *State) fastCopy(src, dst upspin.PathName) error {
	_, err := s.client.PutDuplicate(src, dst)
	if err == nil {
		return nil
	}
	if errors.Match(errExist, err) {
		// File already exists, which PutDuplicate doesn't handle.
		// Use regular copy. We could remove it and retry
		// but that's a little scary.
		return err
	}
	if errors.Match(errIsDir, err) {
		// Oops, we have a directory. Retry.
		return err
	}
	// Unexpected error. Die.
	s.fail(err)
	return nil
}

func (cs *copyState) doCopy(reader io.ReadCloser, writer io.WriteCloser) {
	defer func() {
		reader.Close()
		err := writer.Close()
		if err != nil {
			cs.state.fail(err)
		}
	}()
	_, err := io.Copy(writer, reader)
	if err != nil {
		cs.state.fail(err)
	}
}

// glob glob-expands the argument, which could be a local file
// name or an Upspin path name.
func (cs *copyState) glob(pattern string) (files []cpFile) {
	parsed, err := path.Parse(upspin.PathName(pattern))
	if err == nil {
		// It's an Upspin path.
		for _, path := range cs.state.globUpspinPath(parsed.String()) {
			files = append(files, cpFile{
				path:     string(path),
				isUpspin: true,
			})
		}
		return files
	}
	// It's a local path.
	for _, path := range cs.state.globLocal(pattern) {
		files = append(files, cpFile{
			path:     path,
			isUpspin: false,
		})
	}
	return files
}

// contents return the top-level contents of dir as a slice of cpFiles.
func (s *State) contents(cs *copyState, dir cpFile) ([]cpFile, error) {
	if dir.isUpspin {
		entries, err := s.client.Glob(upspin.AllFilesGlob(upspin.PathName(dir.path)))
		if err != nil {
			s.fail(err)
			// OK to continue; there may still be files.
		}
		files := make([]cpFile, len(entries))
		for i, entry := range entries {
			files[i] = cpFile{
				path:     string(entry.Name),
				isUpspin: true,
			}
		}
		return files, err
	}
	// Local directory.
	fd, err := os.Open(dir.path)
	if err != nil {
		s.fail(err)
		return nil, err
	}
	defer fd.Close()
	names, err := fd.Readdirnames(0)
	if err != nil {
		s.fail(err)
		// OK to continue; there may still be files.
	}
	files := make([]cpFile, len(names))
	for i, name := range names {
		files[i] = cpFile{
			path:     filepath.Join(dir.path, name),
			isUpspin: false,
		}
	}
	return files, err
}
