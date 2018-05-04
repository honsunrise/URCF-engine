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
	Name         string       `yaml:"name"`
	Desc         string       `yaml:"desc"`
	Version      string       `yaml:"version"`
	Architecture Architecture `yaml:"architecture"`
	OS           OS           `yaml:"os"`
	Homepage     string       `yaml:"homepage"`
	Maintainer   string       `yaml:"maintainer"`
	Checksum     string       `yaml:"checksum"`
	EnterPoint   string       `yaml:"enter-point"`
	Conffiles    []string     `yaml:"conffiles"`
	Deps         []Pkg        `yaml:"deps"`
	SysDeps      []Pkg        `yaml:"sys-deps"`
	Licenses     []License    `yaml:"licenses"`
	PreInstall   []string     `yaml:"pre-install"`
	PostInstall  []string     `yaml:"post-install"`
	CoverFile    string       `yaml:"cover-file"`
	WebsDir      string       `yaml:"webs-dir"`
}
