package jplag

import (
	"path/filepath"
	"strings"
)

const (
	DEFAULT_LANGUAGE = "text"

	LANG_C          = "c"
	LANG_CPP        = "cpp"
	LANG_CSHARP     = "csharp"
	LANG_EMF        = "emf"
	LANG_EMF_MODEL  = "emf-model"
	LANG_GO         = "go"
	LANG_JAVA       = "java"
	LANG_JAVASCRIPT = "javascript"
	LANG_KOTLIN     = "kotlin"
	LANG_LLVMIR     = "llvmir"
	LANG_PYTHON3    = "python3"
	LANG_RLANG      = "rlang"
	LANG_RUST       = "rust"
	LANG_SCALA      = "scala"
	LANG_SCHEME     = "scheme"
	LANG_SCXML      = "scxml"
	LANG_SWIFT      = "swift"
	LANG_TYPESCRIPT = "typescript"
)

// {extension (with period): language name, ...}
var extensionToLanguage map[string]string = map[string]string{
	".c": LANG_C,
	".h": LANG_C,

	".cpp": LANG_CPP,
	".c++": LANG_CPP,
	".cc":  LANG_CPP,
	".cp":  LANG_CPP,
	".h++": LANG_CPP,
	".hh":  LANG_CPP,
	".hpp": LANG_CPP,

	".cs":  LANG_CSHARP,
	".csx": LANG_CSHARP,

	".go": LANG_GO,

	".java": LANG_JAVA,

	".js": LANG_JAVASCRIPT,

	".kt": LANG_KOTLIN,
	".km": LANG_KOTLIN,
	".ks": LANG_KOTLIN,

	".ll": LANG_LLVMIR,

	".py": LANG_PYTHON3,

	".r": LANG_RLANG,

	".rs": LANG_RUST,

	".scala": LANG_SCALA,

	".scm": LANG_SCHEME,
	".sps": LANG_SCHEME,
	".sls": LANG_SCHEME,
	".sld": LANG_SCHEME,

	".scxml": LANG_SCXML,

	".swift": LANG_SWIFT,

	".ts": LANG_TYPESCRIPT,
}

func getLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	language, ok := extensionToLanguage[ext]
	if ok {
		return language
	}

	return DEFAULT_LANGUAGE
}
