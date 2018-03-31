package plugin

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/utils"
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

// Version is found in Header.Ident[EI_VERSION] and Header.Version.
type Version uint32

// Machine is found in Header.Machine.
type Machine uint16

const (
	M_NONE    Machine = 0 /* Unknown machine. */
	M_X86     Machine = 1 /* x86. */
	M_X86_64  Machine = 2 /* x86-64. */
	M_ARM     Machine = 3 /* ARM. */
	M_AARCH64 Machine = 4 /* ARM 64-bit Architecture (AArch64) */
	M_MIPS    Machine = 5 /* MIPS. */
	M_IA_64   Machine = 6 /* Intel IA-64 Processor. */
)

var machineStrings = []utils.IntName{
	{0, "M_NONE"},
	{1, "M_X86"},
	{2, "M_X86_64"},
	{3, "M_ARM"},
	{4, "M_AARCH64"},
	{5, "M_MIPS"},
	{6, "M_IA_64"},
}

func (i Machine) String() string   { return utils.StringName(uint32(i), machineStrings, "upp.", false) }
func (i Machine) GoString() string { return utils.StringName(uint32(i), machineStrings, "upp.", true) }

type Flags uint32

var flagsStrings = []utils.IntName{
	{0, "F_NONE"},
}

func (i Flags) String() string   { return utils.StringName(uint32(i), flagsStrings, "upp.", false) }
func (i Flags) GoString() string { return utils.StringName(uint32(i), flagsStrings, "upp.", true) }

type File struct {
	io.Closer
	Version Version
	Machine Machine
	Flags   Flags
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
	if mag[0] != '.' || mag[1] != 'U' || mag[2] != 'P' || mag[3] != 'P' {
		return nil, &PositionError{0, "bad magic number", mag[0:4]}
	}

	f := new(File)

	// Read ELF file header
	var shoff int64
	var shentsize, shnum int

	hdr := new(FileHead)
	sr.Seek(0, io.SeekStart)
	if err := binary.Read(sr, binary.BigEndian, hdr); err != nil {
		return nil, err
	}
	f.Version = Version(hdr.Version)
	f.Machine = Machine(hdr.Machine)
	f.Flags = Flags(hdr.Flags)
	shoff = int64(hdr.Shoff)
	shentsize = int(hdr.Shentsize)
	shnum = int(hdr.Shnum)

	if shnum < 0 {
		return nil, &PositionError{0, "invalid UPP shnum", shnum}
	}

	if shoff < 0 {
		return nil, &PositionError{0, "invalid UPP shoff", shoff}
	}

	if shentsize < 0 {
		return nil, &PositionError{0, "invalid UPP shoff", shentsize}
	}

	return f, nil
}
