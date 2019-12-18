package main

import (
	"fmt"
	"io"
	"os"
)

const IsoAlign = 0x800

type PackFileSystemEntry struct {
	offset int16
	length int16
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fIndex, err := os.Open("binindex.bin")
	check(err)
	defer fIndex.Close()

	packCount := getPackCount(fIndex)

	err = os.MkdirAll("out", 0700)
	check(err)

	for packIndex := 0; packIndex < packCount; packIndex++ {
		packFileName := getPackName(packIndex)
		fPack, err := os.Open(packFileName)
		check(err)

		fs, err := readFileSystem(fIndex, packIndex)
		check(err)

		err = unPack(fPack, packIndex, fs)
		check(err)

		fPack.Close()
	}
}

func unPack(fPack io.ReadSeeker, packIndex int, entries []PackFileSystemEntry) error {
	for index, entry := range entries {
		fPack.Seek(int64(entry.offset)*IsoAlign, os.SEEK_SET)

		fileLength := int(entry.length) * IsoAlign
		data := make([]byte, fileLength)
		_, err := fPack.Read(data)
		if err != nil {
			return err
		}

		fOut, err := os.Create(getOutputFileName(packIndex, index))
		if err != nil {
			return err
		}

		defer fOut.Close()

		_, err = fOut.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func readFileSystem(reader io.ReadSeeker, index int) ([]PackFileSystemEntry, error) {
	reader.Seek(int64(index*IsoAlign), os.SEEK_SET)

	entries := make([]PackFileSystemEntry, 0)

	for {
		entry, err := readFileSystemEntry(reader)
		if err != nil {
			return entries, nil
		}

		if isEndOfFileSystem(entry) {
			break
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func readFileSystemEntry(reader io.Reader) (PackFileSystemEntry, error) {
	data := make([]byte, 4)
	_, err := reader.Read(data)
	if err != nil {
		return PackFileSystemEntry{}, err
	}

	var entry PackFileSystemEntry
	entry.offset = int16(data[0]) | (int16(data[1]) << 8)
	entry.length = int16(data[2]) | (int16(data[3]) << 8)

	return entry, nil
}

func isEndOfFileSystem(entry PackFileSystemEntry) bool {
	return entry.length == 0
}

func getOutputFileName(packIndex int, entryIndex int) string {
	return fmt.Sprintf("out/pack%d_file%d.bin", packIndex, entryIndex)
}

func getPackName(index int) string {
	return fmt.Sprintf("binpack%d.bin", index)
}

func getPackCount(f *os.File) int {
	stat, err := f.Stat()
	check(err)

	return int(stat.Size()) / IsoAlign
}
