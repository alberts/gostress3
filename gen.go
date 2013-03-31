package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type pkgDirsVisitor struct {
	pkgDirs []string
}

type Runnable struct {
	Package string
	Name    string
}

type TestMain struct {
	pkgName    string
	tests      []Runnable
	benchmarks []Runnable
}

func packageName(pkgDir string) string {
	parts := strings.Split(pkgDir, "/")
	return strings.Join(parts[1:len(parts)-1], "/")
}

func underscorePkgName(name string) string {
	return strings.Replace(name, "/", "_", -1)
}

var includePackages = map[string]struct{}{
	"cmd/api":             {},
	"cmd/fix":             {},
	"cmd/go":              {},
	"cmd/gofmt":           {},
	"archive/tar":         {},
	"archive/zip":         {},
	"bufio":               {},
	"bytes":               {},
	"compress/bzip2":      {},
	"compress/flate":      {},
	"compress/gzip":       {},
	"compress/lzw":        {},
	"compress/zlib":       {},
	"container/heap":      {},
	"container/list":      {},
	"container/ring":      {},
	"crypto/aes":          {},
	"crypto/cipher":       {},
	"crypto/des":          {},
	"crypto/dsa":          {},
	"crypto/ecdsa":        {},
	"crypto/elliptic":     {},
	"crypto/hmac":         {},
	"crypto/md5":          {},
	"crypto/rand":         {},
	"crypto/rc4":          {},
	"crypto/rsa":          {},
	"crypto/sha1":         {},
	"crypto/sha256":       {},
	"crypto/sha512":       {},
	"crypto/subtle":       {},
	"crypto/tls":          {},
	"crypto/x509":         {},
	"database/sql":        {},
	"database/sql/driver": {},
	"debug/dwarf":         {},
	"debug/elf":           {},
	"debug/gosym":         {},
	"debug/macho":         {},
	"debug/pe":            {},
	"encoding/ascii85":    {},
	"encoding/asn1":       {},
	"encoding/base32":     {},
	"encoding/base64":     {},
	"encoding/binary":     {},
	"encoding/csv":        {},
	"encoding/gob":        {},
	"encoding/hex":        {},
	"encoding/json":       {},
	"encoding/pem":        {},
	"encoding/xml":        {},
	"errors":              {},
	"expvar":              {},
	"flag":                {},
	"fmt":                 {},
	"go/ast":              {},
	"go/build":            {},
	"go/doc":              {},
	"go/format":           {},
	"go/parser":           {},
	"go/printer":          {},
	"go/scanner":          {},
	"go/token":            {},
	"hash/adler32":        {},
	"hash/crc32":          {},
	"hash/crc64":          {},
	"hash/fnv":            {},
	"html":                {},
	"html/template":       {},
	"image":               {},
	"image/color":         {},
	"image/draw":          {},
	"image/gif":           {},
	"image/jpeg":          {},
	"image/png":           {},
	"index/suffixarray":   {},
	"io":                  {},
	"io/ioutil":           {},
	"log":                 {},
	"log/syslog":          {},
	"math":                {},
	"math/big":            {},
	"math/cmplx":          {},
	"math/rand":           {},
	"mime":                {},
	"mime/multipart":      {},
	"net":                 {},
	"net/http":            {},
	"net/http/cgi":        {},
	"net/http/cookiejar":  {},
	"net/http/fcgi":       {},
	"net/http/httptest":   {},
	"net/http/httputil":   {},
	"net/mail":            {},
	"net/rpc":             {},
	"net/rpc/jsonrpc":     {},
	"net/smtp":            {},
	"net/textproto":       {},
	"net/url":             {},
	"os":                  {},
	"os/exec":             {},
	"os/signal":           {},
	"os/user":             {},
	"path":                {},
	"path/filepath":       {},
	"reflect":             {},
	"regexp":              {},
	"regexp/syntax":       {},
	"runtime":             {},
	"runtime/debug":       {},
	"runtime/pprof":       {},
	"sort":                {},
	"strconv":             {},
	"strings":             {},
	"sync":                {},
	"sync/atomic":         {},
	"syscall":             {},
	"testing/quick":       {},
	"text/scanner":        {},
	"text/tabwriter":      {},
	"text/template":       {},
	"text/template/parse": {},
	"time":                {},
	"unicode":             {},
	"unicode/utf16":       {},
	"unicode/utf8":        {},
}

func parseTestMains(pkgDirs []string) ([]*TestMain, error) {
	testMains := make([]*TestMain, 0)

	for _, pkgDir := range pkgDirs {
		testmain := path.Join(pkgDir, "_testmain.go")
		fileNode, err := parser.ParseFile(token.NewFileSet(), testmain, nil, 0)
		if err != nil {
			return nil, err
		}

		pkgName := packageName(pkgDir)
		if _, ok := includePackages[pkgName]; !ok {
			continue
		}

		tests := make([]Runnable, 0)
		benchmarks := make([]Runnable, 0)
		for _, decl := range fileNode.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR || len(genDecl.Specs) != 1 {
				continue
			}
			spec := genDecl.Specs[0].(*ast.ValueSpec)
			name := spec.Names[0].Name
			if name != "tests" && name != "benchmarks" {
				continue
			}
			elts := spec.Values[0].(*ast.CompositeLit).Elts
			for _, elt := range elts {
				selExpr := elt.(*ast.CompositeLit).Elts[1].(*ast.SelectorExpr)
				expr := selExpr.X.(*ast.Ident).Name
				fieldSel := selExpr.Sel.Name
				if name == "tests" {
					tests = append(tests, Runnable{expr, fieldSel})
				} else {
					benchmarks = append(benchmarks, Runnable{expr, fieldSel})
				}
			}
		}
		if len(tests) == 0 && len(benchmarks) == 0 {
			continue
		}
		testMains = append(testMains, &TestMain{pkgName, tests, benchmarks})
	}
	return testMains, nil
}

const code = `
package main

import "flag"
import "sync"
import "testing"

import (
{{range .Packages}}{{.Name}} "{{.Path}}"
{{end}}
)

{{range .Tests}}var {{.UnderscoreName}}_tests = []testing.InternalTest{
{{range .Functions}}{"{{.Name}}", {{.Package}}.{{.Name}}},
{{end}}
}

{{end}}
{{range .Benchmarks}}var {{.UnderscoreName}}_benchmarks = []testing.InternalBenchmark{
{{range .Functions}}{"{{.Name}}", {{.Package}}.{{.Name}}},
{{end}}
}

{{end}}
func matchString(pat, str string) (bool, error) {
	return true, nil
}

func main() {
	flag.Parse()
	testing.ParseCpuList()

	var wg sync.WaitGroup
{{range .Packages}}
wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if len({{.UnderscoreName}}_tests) == 0 {
				return
			}
			testing.RunTests(matchString, {{.UnderscoreName}}_tests)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		return
		for {
			if len({{.UnderscoreName}}_benchmarks) == 0 {
				return
			}
			testing.RunBenchmarks(matchString, {{.UnderscoreName}}_benchmarks)
		}
	}()
{{end}}

	wg.Wait()
}
`

type TestMains []*TestMain

type Package struct {
	Name           string
	Path           string
	UnderscoreName string
}

type TestBench struct {
	Name    string
	Package string
}

type PackageTests struct {
	UnderscoreName string
	Functions      []TestBench
}

type Data struct {
	Packages   []Package
	Tests      []PackageTests
	Benchmarks []PackageTests
}

func (this TestMains) Data() Data {
	var data Data
	packages := make(map[Package]struct{})
	for _, tm := range this {
		underscoreName := underscorePkgName(tm.pkgName)
		var both []Runnable
		both = append(both, tm.tests...)
		both = append(both, tm.benchmarks...)
		for _, t := range both {
			path := tm.pkgName + "/_test/" + tm.pkgName
			if t.Package == "_xtest" {
				path = path + "_test"
			}
			pkg := Package{
				Name:           underscoreName + t.Package,
				Path:           path,
				UnderscoreName: underscoreName,
			}
			packages[pkg] = struct{}{}
		}

		tests := PackageTests{UnderscoreName: underscoreName}
		for _, t := range tm.tests {
			tb := TestBench{
				Name:    t.Name,
				Package: underscoreName + t.Package,
			}
			tests.Functions = append(tests.Functions, tb)
		}

		benchs := PackageTests{UnderscoreName: underscoreName}
		for _, t := range tm.benchmarks {
			tb := TestBench{
				Name:    t.Name,
				Package: underscoreName + t.Package,
			}
			benchs.Functions = append(benchs.Functions, tb)
		}

		data.Tests = append(data.Tests, tests)
		data.Benchmarks = append(data.Benchmarks, benchs)
	}

	for pkg := range packages {
		data.Packages = append(data.Packages, pkg)
	}

	return data
}

func main() {
	tmpl := template.Must(template.New("code").Parse(code))

	var testDirs []string
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if path.Base(p) == "_test" && info.IsDir() {
			testDirs = append(testDirs, p)
		}
		return nil
	}
	filepath.Walk("pkg", walkFunc)
	testMains, err := parseTestMains(testDirs)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("main.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, TestMains(testMains).Data()); err != nil {
		panic(err)
	}
}
