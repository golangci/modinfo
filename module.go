package modinfo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
)

type ModInfo struct {
	Path      string `json:"Path"`
	Dir       string `json:"Dir"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
	Main      bool   `json:"Main"`
}

var Analyzer = &analysis.Analyzer{
	Name:       "modinfo",
	Doc:        "Module information",
	Run:        run,
	ResultType: reflect.TypeOf(([]ModInfo)(nil)),
}

func run(pass *analysis.Pass) (any, error) {
	// https://github.com/golang/go/issues/44753#issuecomment-790089020
	cmd := exec.Command("go", "list", "-m", "-json")
	// FIXME useless?
	for _, file := range pass.Files {
		f := pass.Fset.File(file.Pos()).Name()
		if filepath.Ext(f) != ".go" {
			continue
		}

		cmd.Dir = filepath.Dir(f)
		break
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command go list: %w: %s", err, string(out))
	}

	var infos []ModInfo

	for dec := json.NewDecoder(bytes.NewBuffer(out)); dec.More(); {
		var v ModInfo
		if err := dec.Decode(&v); err != nil {
			return nil, fmt.Errorf("unmarshaling error: %w: %s", err, string(out))
		}

		if v.GoMod == "" {
			return nil, errors.New("working directory is not part of a module")
		}

		if !v.Main || v.Dir == "" {
			continue
		}

		infos = append(infos, v)
	}

	if len(infos) == 0 {
		return nil, errors.New("go.mod file not found")
	}

	sort.Slice(infos, func(i, j int) bool {
		return len(infos[i].Path) > len(infos[j].Path)
	})

	return infos, nil
}

func FindModule(infos []ModInfo, pass *analysis.Pass) (ModInfo, error) {
	var name string
	for _, file := range pass.Files {
		f := pass.Fset.File(file.Pos()).Name()
		if filepath.Ext(f) != ".go" {
			continue
		}

		name = f
		break
	}

	if name == "" {
		return ModInfo{}, errors.New("OOPS")
	}

	for _, info := range infos {
		if !strings.HasPrefix(name, info.Dir) {
			continue
		}
		return info, nil
	}

	return ModInfo{}, errors.New("module information not found")
}

func ReadModuleFile(info ModInfo) (*modfile.File, error) {
	raw, err := os.ReadFile(info.GoMod)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod file: %w", err)
	}

	return modfile.Parse("go.mod", raw, nil)
}
