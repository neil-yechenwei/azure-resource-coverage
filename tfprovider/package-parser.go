package tfprovider

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

var importRe = regexp.MustCompile(`(?sU)import\s+\((?P<imports>.+)\)`)
var pkgDefRe = regexp.MustCompile(`(?m)^\s*(?P<alias>[a-zA-Z0-9]+)?\s*"(?P<packageprefix>.+/)(?P<apiversion>[0-9-.vV]+[a-zA-Z0-9]+)(?P<packagesuffix>/.+)"\s*$`)

func parsePackages(path string) ([]*GoPackage, error) {
	pkgs := make([]*GoPackage, 0)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse Go packages in %q: %v", path, err)
	}

	subm := importRe.FindStringSubmatch(string(buf))
	imports, err := parseImports(subm)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse Go packages in %q: %v", path, err)
	}

	for _, def := range pkgDefRe.FindAllStringSubmatch(imports, -1) {
		pkg, err := parsePackage(def)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse Go packages in %q: %v", path, err)
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func parseImports(content []string) (string, error) {
	_, captures, err := toNamedCaptures(content, importRe)
	if err != nil {
		return "", err
	}

	imports, ok := captures["imports"]
	if !ok || imports == "" {
		return "", fmt.Errorf("Cannot parse Go imports content")
	}

	return imports, nil
}

func parsePackage(def []string) (*GoPackage, error) {
	m, captures, err := toNamedCaptures(def, pkgDefRe)
	if err != nil {
		return nil, err
	}

	alias := captures["alias"]

	pkgPrefix, ok := captures["packageprefix"]
	if !ok || pkgPrefix == "" {
		return nil, fmt.Errorf("Cannot parse Go package prefix name in %q", m)
	}

	apiVersion, ok := captures["apiversion"]
	if !ok || apiVersion == "" {
		return nil, fmt.Errorf("Cannot parse Go api version in %q", m)
	}

	packageSuffix, ok := captures["packagesuffix"]
	if !ok || packageSuffix == "" {
		return nil, fmt.Errorf("Cannot parse Go package suffix name in %q", m)
	}

	pkg := pkgPrefix + apiVersion + packageSuffix

	return &GoPackage{
		alias,
		pkg,
		apiVersion,
	}, nil
}
