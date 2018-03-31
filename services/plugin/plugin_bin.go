package plugin

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
