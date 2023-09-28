package main

import (
	"io/fs"
	"path/filepath"
	"bytes"
	"strings"
	"fmt"
	"os"
	"io/ioutil"
	rs "github.com/cshu/golangrs"
)


func main() {
	dMap := make(map[string]bool)
	for idx, dStr := range os.Args {//fixme imitate gcc, use -D to define macro?
		if 0 == idx {
			continue
		}
		dMap[dStr] = true
	}
	//var filelist []fs.DirEntry
	var filelist []string
	err := filepath.WalkDir(`.`, func(path string, d fs.DirEntry, errRelatedToPath error)error{
		rs.CheckErr(errRelatedToPath)
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, `.go`) {
			return nil
		}
		//filelist = append(filelist, d)
		filelist = append(filelist, path)
		return nil
	})
	rs.CheckErr(err)
	//fmt.Println(`Files:`)
	//for _, filenm := range filelist{
	//	fmt.Println(filenm)
	//}
	for _, filenm := range filelist{
		var nestedDirective int
		var nestedLvlOmit int = -1
		//var ifdefOrIfndef []bool//note true means lines are preserved; false means lines are omitted
		var outbuf bytes.Buffer
		var anyOmit bool
		rs.ReadHugeFilelines(filenm, func(line string) bool {
			const ifdef = `//pmpre#ifdef `
			const ifndef = `//pmpre#ifndef `
			const endif = `//pmpre#endif`
			var omitLine bool = true
			if nestedLvlOmit < 0 {
				outbuf.WriteString(line)
				outbuf.WriteString("\n")
				omitLine = false
			}
			trimmedLine := strings.TrimSpace(line)//.TrimLeft(line, "\t ")
			switch {
			case strings.HasPrefix(trimmedLine, ifdef):
				nestedDirective++
				identifier := trimmedLine[len(ifdef):]
				isDefined := dMap[identifier]
				preserveElseOmit := isDefined
				if !preserveElseOmit && nestedLvlOmit < 0 {
					nestedLvlOmit = nestedDirective
				}
			case strings.HasPrefix(trimmedLine, ifndef):
				nestedDirective++
				identifier := trimmedLine[len(ifndef):]
				isDefined := dMap[identifier]
				preserveElseOmit := !isDefined
				if !preserveElseOmit && nestedLvlOmit < 0 {
					nestedLvlOmit = nestedDirective
				}
			case strings.HasPrefix(trimmedLine, endif):
				if len(trimmedLine) != len(endif) {
					panic(`Bad endif`)//fixme you should not use panic here, just print msg and quit
				}
				if nestedDirective == nestedLvlOmit {
					nestedLvlOmit = -1
					outbuf.WriteString(line)
					outbuf.WriteString("\n")
					omitLine = false
				}
				nestedDirective--
				if nestedDirective < 0 {
					panic(`Excessive endif`)//fixme you should not use panic here, just print msg and quit
				}
			}
			anyOmit = anyOmit || omitLine
			//if strings.HasPrefix(line, ifdef) {
			//	identifier := line[len(ifdef):]
			//	isDefined := dMap[identifier]
			//	preserveElseOmit := isDefined
			//	ifdefOrIfndef = append(ifdefOrIfndef, preserveElseOmit)
			//}
			//if strings.HasPrefix(line, ifndef) {
			//	identifier := line[len(ifndef):]
			//	isDefined := dMap[identifier]
			//	preserveElseOmit := !isDefined
			//	ifdefOrIfndef = append(ifdefOrIfndef, preserveElseOmit)
			//}
			//if strings.HasPrefix(line, endif) {
			//	if line != endif {
			//		panic(`Bad endif`)//fixme you should not use panic here, just print msg and quit
			//	}
			//	ifdefOrIfndef = ifdefOrIfndef[:len(ifdefOrIfndef)-1]
			//	//undone
			//}
			//if ifdefOrIfndef[len(ifdefOrIfndef)-1] {
			//	outbuf.WriteString(line)
			//	outbuf.WriteString("\n")
			//}
			return true
		})
		//if 0 != len(ifdefOrIfndef) {
		//	panic(`Block not closed`)//fixme you should not use panic here, just print msg and quit
		//}
		if 0 != nestedDirective {
			panic(`Missing endif`)//fixme you should not use panic here, just print msg and quit
		}
		if anyOmit {
			fmt.Println(`Replacing: `+filenm)
			err = ioutil.WriteFile(filenm, outbuf.Bytes(), 0644)
			rs.CheckErr(err)
		} else {
			fmt.Println(`Skipping: `+filenm)
		}
	}
	fmt.Println(`All src processed`)
	var outbuf bytes.Buffer
	outbuf.WriteString(strings.Join(os.Args, ` `))
	outbuf.WriteString("\n")
	err = ioutil.WriteFile(`PMPRE_RUN_REPORT`, outbuf.Bytes(), 0644)
	rs.CheckErr(err)
	fmt.Println(`All done`)
}
