/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"strconv"
)

type service struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Config      string /* `toml:"config"` */
	Filename    string
}

var services []service

func parseFile(path string) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	svc := service{Filename: path}
	var tomlData string
parsing:
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text
			if !strings.HasPrefix(text, "/* Metadata:") {
				continue
			}
			text = strings.Replace(text, "/* Metadata:", "", 1)
			i := strings.LastIndex(text, "*/")
			if i == -1 {
				fmt.Fprintf(os.Stderr, "%s: Found metadata beginning but no end\n", path)
				continue
			}
			tomlData = text[:i]
			break parsing
		}
	}
	if tomlData == "" {
		return
	}
	md, err := toml.Decode(tomlData, &svc)
	if len(md.Undecoded()) != 0 {
		fmt.Fprintf(os.Stderr, "Unrecognized keys: %v\n", md.Undecoded())
	}

	configMap := make(map[string]interface{})
	var configStruct *ast.StructType
	for _, _decl := range file.Decls {
		decl, ok := _decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if decl.Tok != token.TYPE {
			continue
		}
		typeSpec := decl.Specs[0].(*ast.TypeSpec)
		if !strings.HasSuffix(typeSpec.Name.Name, "Config") {
			continue
		}
		fmt.Printf("Inspecting struct %s.\n", typeSpec.Name.Name)
		configStruct, ok = typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}
	}
	if configStruct == nil {
		panic("No config struct found!")
	}
	for _, field := range configStruct.Fields.List {
		if field.Tag == nil {
			continue
		}
		if len(field.Names) != 1 {
			panic("Not supported")
		}
		if field.Type.(*ast.Ident).Name != "string" {
			panic("Not supported")
		}
		name := field.Names[0].Name
		tag := reflect.StructTag(strings.Replace(field.Tag.Value, "`", "", -1))
		tomlTag, tomlOk := tag.Lookup("toml")
		defaultTag, defaultOk := tag.Lookup("default")
		if !tomlOk {
			if !defaultOk {
				continue
			} else {
				panic("Field " + name + "has a default tag, but no toml tag")
			}
		}
		if !defaultOk {
			panic("No default tag for field " + name)
			continue
		}
		switch field.Type.(*ast.Ident).Name {
		case "string":
			configMap[tomlTag] = defaultTag
		case "int":
			val, err := strconv.Atoi(defaultTag)
			if err != nil {
				panic(err)
			}
			configMap[tomlTag] = val
		default:
			panic("Type " + field.Type.(*ast.Ident).Name + "is not supported")
		}
	}

	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(configMap)
	if err != nil {
		panic(err)
	}
	svc.Config = buf.String()

	tmp := struct{}{}
	_, err = toml.Decode(svc.Config, &tmp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "`config` is not valid TOML: %s\n", err.Error())
	}
	services = append(services, svc)
	fmt.Fprintf(os.Stderr, "Added service %s\n", svc.Name)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gen-service-docs directory (where `directory` contains the services to be documented)")
		return
	}
	// We do not use parser.ParseDir, because there are subfolders that we need to explore
	err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			parseFile(path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	j, err := json.Marshal(services)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	json.Indent(&out, j, "", "    ")
	fmt.Println(out.String())
}
