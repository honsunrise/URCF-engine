package plugin

type Architecture string

// Machine is found in Header.Machine.
type Machine uint16

const (
	ARCH_ALL     Architecture = "ALL"     /* Unknown machine. */
	ARCH_X86     Architecture = "X86"     /* x86. */
	ARCH_X86_64  Architecture = "X86_64"  /* x86-64. */
	ARCH_ARM     Architecture = "ARM"     /* ARM. */
	ARCH_AARCH64 Architecture = "AArch64" /* ARM 64-bit Architecture (AArch64) */
	ARCH_MIPS    Architecture = "MIPS"    /* MIPS. */
	ARCH_IA_64   Architecture = "IA-64"   /* Intel IA-64 Processor. */
)

type OS string

type Pkg struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type License struct {
	Name string `yaml:"name"`
	Path string `yaml:"desc"`
}

type PluginManifest struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Architecture Architecture
	OS           OS
	Homepage     string
	Maintainer   string
	Checksum     string
	Conffiles    []string
	Deps         []Pkg
	SysDeps      []Pkg `yaml:"sys-deps"`
	Licenses     []License
	PreInstall   []string `yaml:"pre-install"`
	PostInstall  []string `yaml:"post-install"`
}
