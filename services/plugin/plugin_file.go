package plugin

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

type PositionError struct {
	off int64
	msg string
	val interface{}
}

func (e *PositionError) Error() string {
	msg := e.msg
	if e.val != nil {
		msg += fmt.Sprintf(" '%v' ", e.val)
	}
	msg += fmt.Sprintf("in record at byte %#x", e.off)
	return msg
}

// Initial magic number for ELF files.
const UPPMAG = ".UPP"

type FileHead struct {
	Mag       [4]byte
	Machine   uint16 /* Machine architecture. */
	Version   uint32 /* ELF format version. */
	Flags     uint32 /* Architecture-specific flags. */
	Hsize     uint16 /* Size of header in bytes. */
	Shentsize uint16 /* Size of section header entry. */
	Shoff     uint32 /* Section header file offset. */
	Shnum     uint16 /* Number of section header entries. */
}

type FileSection struct {
	Type  uint32 /* Section type. */
	Flags uint32 /* Section flags. */
	Off   uint32 /* Offset in file. */
	Size  uint32 /* Size in bytes. */
}

type File struct {
}

func (*File) Open(name string) {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(name)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.CopyN(os.Stdout, rc, 68)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		fmt.Println()
	}
}

func NewFile(r io.ReaderAt) (*File, error) {
	sr := io.NewSectionReader(r, 0, 1<<63-1)
	// Read and decode ELF identifier
	var mag [4]uint8
	if _, err := sr.ReadAt(mag[0:], 0); err != nil {
		return nil, err
	}
	if mag[0] != '.' || mag[1] != 'E' || mag[2] != 'L' || mag[3] != 'F' {
		return nil, &PositionError{0, "bad magic number", mag[0:4]}
	}

	f := new(File)

	f.Version = Version(ident[EI_VERSION])
	if f.Version != EV_CURRENT {
		return nil, &PositionError{0, "unknown ELF version", f.Version}
	}

	// Read ELF file header
	var phoff int64
	var phentsize, phnum int
	var shoff int64
	var shentsize, shnum, shstrndx int
	shstrndx = -1
	switch f.Class {
	case ELFCLASS32:
		hdr := new(Header32)
		sr.Seek(0, io.SeekStart)
		if err := binary.Read(sr, f.ByteOrder, hdr); err != nil {
			return nil, err
		}
		f.Type = Type(hdr.Type)
		f.Machine = Machine(hdr.Machine)
		f.Entry = uint64(hdr.Entry)
		if v := Version(hdr.Version); v != f.Version {
			return nil, &PositionError{0, "mismatched ELF version", v}
		}
		phoff = int64(hdr.Phoff)
		phentsize = int(hdr.Phentsize)
		phnum = int(hdr.Phnum)
		shoff = int64(hdr.Shoff)
		shentsize = int(hdr.Shentsize)
		shnum = int(hdr.Shnum)
		shstrndx = int(hdr.Shstrndx)
	case ELFCLASS64:
		hdr := new(Header64)
		sr.Seek(0, io.SeekStart)
		if err := binary.Read(sr, f.ByteOrder, hdr); err != nil {
			return nil, err
		}
		f.Type = Type(hdr.Type)
		f.Machine = Machine(hdr.Machine)
		f.Entry = hdr.Entry
		if v := Version(hdr.Version); v != f.Version {
			return nil, &PositionError{0, "mismatched ELF version", v}
		}
		phoff = int64(hdr.Phoff)
		phentsize = int(hdr.Phentsize)
		phnum = int(hdr.Phnum)
		shoff = int64(hdr.Shoff)
		shentsize = int(hdr.Shentsize)
		shnum = int(hdr.Shnum)
		shstrndx = int(hdr.Shstrndx)
	}

	if shnum > 0 && shoff > 0 && (shstrndx < 0 || shstrndx >= shnum) {
		return nil, &PositionError{0, "invalid ELF shstrndx", shstrndx}
	}

	// Read section headers
	f.Sections = make([]*Section, shnum)
	names := make([]uint32, shnum)
	for i := 0; i < shnum; i++ {
		off := shoff + int64(i)*int64(shentsize)
		sr.Seek(off, io.SeekStart)
		s := new(Section)
		switch f.Class {
		case ELFCLASS32:
			sh := new(Section32)
			if err := binary.Read(sr, f.ByteOrder, sh); err != nil {
				return nil, err
			}
			names[i] = sh.Name
			s.SectionHeader = SectionHeader{
				Type:      SectionType(sh.Type),
				Flags:     SectionFlag(sh.Flags),
				Addr:      uint64(sh.Addr),
				Offset:    uint64(sh.Off),
				FileSize:  uint64(sh.Size),
				Link:      sh.Link,
				Info:      sh.Info,
				Addralign: uint64(sh.Addralign),
				Entsize:   uint64(sh.Entsize),
			}
		case ELFCLASS64:
			sh := new(Section64)
			if err := binary.Read(sr, f.ByteOrder, sh); err != nil {
				return nil, err
			}
			names[i] = sh.Name
			s.SectionHeader = SectionHeader{
				Type:      SectionType(sh.Type),
				Flags:     SectionFlag(sh.Flags),
				Offset:    sh.Off,
				FileSize:  sh.Size,
				Addr:      sh.Addr,
				Link:      sh.Link,
				Info:      sh.Info,
				Addralign: sh.Addralign,
				Entsize:   sh.Entsize,
			}
		}
		s.sr = io.NewSectionReader(r, int64(s.Offset), int64(s.FileSize))

		if s.Flags&SHF_COMPRESSED == 0 {
			s.ReaderAt = s.sr
			s.Size = s.FileSize
		} else {
			// Read the compression header.
			switch f.Class {
			case ELFCLASS32:
				ch := new(Chdr32)
				if err := binary.Read(s.sr, f.ByteOrder, ch); err != nil {
					return nil, err
				}
				s.compressionType = CompressionType(ch.Type)
				s.Size = uint64(ch.Size)
				s.Addralign = uint64(ch.Addralign)
				s.compressionOffset = int64(binary.Size(ch))
			case ELFCLASS64:
				ch := new(Chdr64)
				if err := binary.Read(s.sr, f.ByteOrder, ch); err != nil {
					return nil, err
				}
				s.compressionType = CompressionType(ch.Type)
				s.Size = ch.Size
				s.Addralign = ch.Addralign
				s.compressionOffset = int64(binary.Size(ch))
			}
		}

		f.Sections[i] = s
	}

	if len(f.Sections) == 0 {
		return f, nil
	}

	// Load section header string table.
	shstrtab, err := f.Sections[shstrndx].Data()
	if err != nil {
		return nil, err
	}
	for i, s := range f.Sections {
		var ok bool
		s.Name, ok = getString(shstrtab, int(names[i]))
		if !ok {
			return nil, &PositionError{shoff + int64(i*shentsize), "bad section name index", names[i]}
		}
	}

	return f, nil
}
