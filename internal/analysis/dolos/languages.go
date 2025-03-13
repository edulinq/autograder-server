package dolos

import (
	"path/filepath"
	"strings"
)

const (
	DEFAULT_LANGUAGE = "char"

	LANG_BASH       = "bash"
	LANG_C          = "c"
	LANG_CPP        = "cpp"
	LANG_CSHARP     = "c-sharp"
	LANG_PYTHON3    = "python"
	LANG_PHP        = "php"
	LANG_MODELICA   = "modelica"
	LANG_OCAML      = "ocaml"
	LANG_JAVA       = "java"
	LANG_JAVASCRIPT = "javascript"
	LANG_ELM        = "elm"
	LANG_GO         = "go"
	LANG_GROOVY     = "groovy"
	LANG_R          = "r"
	LANG_RUST       = "rust"
	LANG_SCALA      = "scala"
	LANG_SQL        = "sql"
	LANG_TYPESCRIPT = "typescript"
	LANG_TSX        = "tsx"
	LANG_VERILOG    = "verilog"
)

// {extension (with period): language name, ...}
var extensionToLanguage map[string]string = map[string]string{
	".sh":   LANG_BASH,
	".bash": LANG_BASH,

	".c": LANG_C,
	".h": LANG_C,

	".cpp": LANG_CPP,
	".c++": LANG_CPP,
	".cxx": LANG_CPP,
	".cc":  LANG_CPP,
	".cp":  LANG_CPP,
	".h++": LANG_CPP,
	".hxx": LANG_CPP,
	".hh":  LANG_CPP,
	".hpp": LANG_CPP,

	".cs":  LANG_CSHARP,
	".csx": LANG_CSHARP,

	".py":  LANG_PYTHON3,
	".py3": LANG_PYTHON3,

	".php":   LANG_PHP,
	".php3":  LANG_PHP,
	".php4":  LANG_PHP,
	".php5":  LANG_PHP,
	".php7":  LANG_PHP,
	".phps":  LANG_PHP,
	".phpt":  LANG_PHP,
	".phtml": LANG_PHP,

	".mo":  LANG_MODELICA,
	".mos": LANG_MODELICA,

	".ml": LANG_OCAML,

	".java": LANG_JAVA,

	".js": LANG_JAVASCRIPT,

	".elm": LANG_ELM,

	".go": LANG_GO,

	".groovy": LANG_GROOVY,
	".gvy":    LANG_GROOVY,
	".gy":     LANG_GROOVY,
	".gsh":    LANG_GROOVY,

	".r":     LANG_R,
	".rdata": LANG_R,
	".rds":   LANG_R,
	".rda":   LANG_R,

	".rs":   LANG_RUST,
	".rlib": LANG_RUST,

	".scala": LANG_SCALA,
	".sc":    LANG_SCALA,

	".sql": LANG_SQL,

	".ts": LANG_TYPESCRIPT,

	".tsx": LANG_TSX,

	".v":  LANG_VERILOG,
	".vh": LANG_VERILOG,
}

func getLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	language, ok := extensionToLanguage[ext]
	if ok {
		return language
	}

	return DEFAULT_LANGUAGE
}
