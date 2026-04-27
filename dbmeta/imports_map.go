package dbmeta

import (
	"strconv"
	"strings"
)

type ImportPackageName string

// ImportItem stores Go import package and its short name
type ImportItem struct {
	Package   ImportPackageName
	ShortName string
}

type ImportsMap struct {
	byShortName   map[string]*ImportItem
	byPackageName map[ImportPackageName]*ImportItem

	pkSpecificImports map[ImportPackageName]*ImportItem
}

// NewImportsMap returns an ImportsMap with initialized internal maps.
func NewImportsMap() *ImportsMap {
	return &ImportsMap{
		byShortName:        make(map[string]*ImportItem),
		byPackageName:      make(map[ImportPackageName]*ImportItem),
		pkSpecificImports:  make(map[ImportPackageName]*ImportItem),
	}
}

// ExtractImport finds the import package from goType, add it to import maps, and replace it with valid field type format.
func (m *ImportsMap) ExtractImport(goType string, isPrimaryKey bool) string {
	parts := strings.Split(goType, ":")
	if len(parts) != 2 {
		return goType
	}
	packageName := ImportPackageName(parts[0])
	typeName := parts[1]
	var shortName string
	importPackage, ok := m.byPackageName[packageName]
	if ok {
		shortName = importPackage.ShortName
		if isPrimaryKey {
			m.pkSpecificImports[packageName] = importPackage
		}
		return shortName + "." + typeName
	}

	packageParts := strings.Split(string(packageName), "/")
	shortName = packageParts[len(packageParts)-1]
	shortName = string(regexNotAlphanum.ReplaceAll([]byte(shortName), []byte("")))
	suffix := 0
	initialShortName := shortName
	for {
		if _, ok := m.byShortName[shortName]; !ok {
			break
		}
		suffix++
		shortName = initialShortName + strconv.Itoa(suffix)
	}
	importItem := &ImportItem{
		Package:   packageName,
		ShortName: shortName,
	}
	m.byPackageName[packageName] = importItem
	m.byShortName[shortName] = importItem
	if isPrimaryKey {
		m.pkSpecificImports[packageName] = importItem
	}
	return shortName + "." + typeName
}

func (m *ImportsMap) GetImports() []*ImportItem {
	imports := make([]*ImportItem, 0, len(m.byPackageName))
	for _, item := range m.byPackageName {
		imports = append(imports, item)
	}
	return imports
}

func (m *ImportsMap) GetPKImports() []*ImportItem {
	imports := make([]*ImportItem, 0, len(m.pkSpecificImports))
	for _, item := range m.pkSpecificImports {
		imports = append(imports, item)
	}
	return imports
}
