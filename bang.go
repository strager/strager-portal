package main

import (
	"net/http"
	"net/url"
	"strings"
)

type bangHandler func(response http.ResponseWriter, request *http.Request, query string)

var bangHandlers map[string]bangHandler = map[string]bangHandler{
	"": func(response http.ResponseWriter, request *http.Request, query string) {
		http.Redirect(response, request, "https://kagi.com/search?q="+url.QueryEscape(query), http.StatusFound)
	},
	"!go": func(response http.ResponseWriter, request *http.Request, query string) {
		var isGoStandardPackage bool

		// Example: !go net/http -> pkg.go.dev/net/http
		_, isGoStandardPackage = goStandardPackageLongNames[query]
		if isGoStandardPackage {
			// NOTE(security): query might contain "/", but that's okay.
			http.Redirect(response, request, "https://pkg.go.dev/"+query, http.StatusFound)
			return
		}

		// Example: !go http -> pkg.go.dev/net/http
		var packageName string
		packageName, isGoStandardPackage = goStandardPackageShortNames[query]
		if isGoStandardPackage {
			// NOTE(security): packageName might contain "/", but that's okay.
			http.Redirect(response, request, "https://pkg.go.dev/"+packageName, http.StatusFound)
			return
		}

		// Example: !go http.Client -> pkg.go.dev/net/http#Client
		var importName string
		var symbolName string
		var isPackageDotSymbol bool
		importName, symbolName, isPackageDotSymbol = strings.Cut(query, ".")
		if isPackageDotSymbol {
			var isGoStandardPackage bool
			packageName, isGoStandardPackage = goStandardPackageShortNames[importName]
			if isGoStandardPackage {
				// NOTE(security): packageName might contain "/", but that's okay.
				http.Redirect(response, request, "https://pkg.go.dev/"+packageName+"#"+url.QueryEscape(symbolName), http.StatusFound)
				return
			}
		}

		http.Redirect(response, request, "https://pkg.go.dev/search?utm_source=godoc&q="+url.QueryEscape(query), http.StatusFound)
	},
}

var goStandardPackageNames []string = []string{
	"archive/tar",
	"archive/zip",
	"bufio",
	"builtin",
	"bytes",
	"cmp",
	"compress/bzip2",
	"compress/flate",
	"compress/gzip",
	"compress/lzw",
	"compress/zlib",
	"container/heap",
	"container/list",
	"container/ring",
	"context",
	"crypto",
	"crypto/aes",
	"crypto/cipher",
	"crypto/des",
	"crypto/dsa",
	"crypto/ecdh",
	"crypto/ecdsa",
	"crypto/ed25519",
	"crypto/elliptic",
	"crypto/fips140",
	"crypto/hkdf",
	"crypto/hmac",
	"crypto/md5",
	"crypto/mlkem",
	"crypto/pbkdf2",
	"crypto/rand",
	"crypto/rc4",
	"crypto/rsa",
	"crypto/sha1",
	"crypto/sha256",
	"crypto/sha3",
	"crypto/sha512",
	"crypto/subtle",
	"crypto/tls",
	"crypto/x509",
	"crypto/x509/pkix",
	"database/sql",
	"database/sql/driver",
	"debug/buildinfo",
	"debug/dwarf",
	"debug/elf",
	"debug/gosym",
	"debug/macho",
	"debug/pe",
	"debug/plan9obj",
	"embed",
	"encoding",
	"encoding/ascii85",
	"encoding/asn1",
	"encoding/base32",
	"encoding/base64",
	"encoding/binary",
	"encoding/csv",
	"encoding/gob",
	"encoding/hex",
	"encoding/json",
	"encoding/pem",
	"encoding/xml",
	"errors",
	"expvar",
	"flag",
	"fmt",
	"go/ast",
	"go/build",
	"go/build/constraint",
	"go/constant",
	"go/doc",
	"go/doc/comment",
	"go/format",
	"go/importer",
	"go/parser",
	"go/printer",
	"go/scanner",
	"go/token",
	"go/types",
	"go/version",
	"hash",
	"hash/adler32",
	"hash/crc32",
	"hash/crc64",
	"hash/fnv",
	"hash/maphash",
	"html",
	"html/template",
	"image",
	"image/color",
	"image/color/palette",
	"image/draw",
	"image/gif",
	"image/jpeg",
	"image/png",
	"index/suffixarray",
	"io",
	"io/fs",
	"io/ioutil",
	"iter",
	"log",
	"log/slog",
	"log/syslog",
	"maps",
	"math",
	"math/big",
	"math/bits",
	"math/cmplx",
	"math/rand",
	"math/rand/v2",
	"mime",
	"mime/multipart",
	"mime/quotedprintable",
	"net",
	"net/http",
	"net/http/cgi",
	"net/http/cookiejar",
	"net/http/fcgi",
	"net/http/httptest",
	"net/http/httptrace",
	"net/http/httputil",
	"net/http/pprof",
	"net/mail",
	"net/netip",
	"net/rpc",
	"net/rpc/jsonrpc",
	"net/smtp",
	"net/textproto",
	"net/url",
	"os",
	"os/exec",
	"os/signal",
	"os/user",
	"path",
	"path/filepath",
	"plugin",
	"reflect",
	"regexp",
	"regexp/syntax",
	"runtime",
	"runtime/cgo",
	"runtime/coverage",
	"runtime/debug",
	"runtime/metrics",
	"runtime/pprof",
	"runtime/race",
	"runtime/trace",
	"slices",
	"sort",
	"strconv",
	"strings",
	"structs",
	"sync",
	"sync/atomic",
	"syscall",
	"syscall/js",
	"testing",
	"testing/fstest",
	"testing/iotest",
	"testing/quick",
	"testing/slogtest",
	"text/scanner",
	"text/tabwriter",
	"text/template",
	"text/template/parse",
	"time",
	"time/tzdata",
	"unicode",
	"unicode/utf16",
	"unicode/utf8",
	"unique",
	"unsafe",
	"weak",
}

// Maps a short name to its long name.
var goStandardPackageShortNames map[string]string

var goStandardPackageLongNames map[string]struct{}

func init() {
	goStandardPackageShortNames = map[string]string{}
	goStandardPackageLongNames = map[string]struct{}{}
	var packageName string
	for _, packageName = range goStandardPackageNames {
		goStandardPackageLongNames[packageName] = struct{}{}
		var parts []string = strings.Split(packageName, "/")
		goStandardPackageShortNames[parts[len(parts)-1]] = packageName
	}
}
