package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type (
	Plugin struct {
		Path []string
		Retry uint
	}
)

func (p Plugin) Exec() error {
	startAt := time.Now()

	if 0 == len(p.Path) {
		// setting default path
		p.Path = []string{"package.json"}
	}

	packageLists := getPackages(p.Path)

	for packageName, packageList := range packageLists {
		go syncPackage(packageName)
		for _, version := range packageList {
			wg.Add(1)
			go func(name, version string) {
				defer wg.Done()
				checkTime := 0
				ok := checkVersion(name, version)
				for !ok {
					if uint(checkTime) > p.Retry {
						fmt.Println("package: ", name, " ,version: ", version, " ,sync timeout")
						break
					}

					syncPackage(name)
					checkTime++
					ok = checkVersion(name, version)
				}
			}(packageName, version)
			//go checkVersion(packageName, version)
		}
	}

	wg.Wait()

	fmt.Println("finish, time consuming: ", time.Now().Sub(startAt))
	return nil
}

func getPackages(paths []string) map[string][]string {
	re, _ := regexp.Compile(`^[0-9].*?$`)

	total := map[string][]string{}

	for _, packagePath := range paths {
		s, err := os.Stat(packagePath)
		if nil != err {
			fmt.Println(packagePath, " no such file or directory")
			continue
		}

		if s.IsDir() {
			packagePath = path.Join(packagePath, "package.json")
		}

		b, err := ioutil.ReadFile(packagePath)
		if nil != err {
			fmt.Println(err)
			continue
		}

		var v map[string]interface{}

		err = json.Unmarshal(b, &v)
		if nil != err {
			fmt.Println(err)
			continue
		}

		for name, lists := range v {
			if !strings.Contains(name, "Dependencies") {
				continue
			}
			pks, ok := lists.(map[string]interface{})
			if !ok {
				continue
			}

			for name, version := range pks {
				if re.MatchString(version.(string)) {
					inArr := false
					for _, val := range total[name] {
						if val == version {
							inArr = true
						}
					}

					if !inArr {
						total[name] = append(total[name], version.(string))
					}
				}
			}
		}
	}

	return total
}

func syncPackage(p string) {
	baseURL := "https://cnpmjs.org/sync"

	url := fmt.Sprintf("%s/%s", baseURL, p)

	_, _ = http.Get(url)
}

func checkVersion(packageName, packageVersion string) bool {

	baseUrl := "https://r.cnpmjs.org"

	url := fmt.Sprintf("%s/%s", baseUrl, packageName)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if nil != err {
		fmt.Println(err)
		return false
	}

	var v map[string]interface{}

	err = json.Unmarshal(body, &v)
	if nil != err {
		return false
	}

	if _, ok := v["versions"].(map[string]interface{})[packageVersion]; ok {
		fmt.Println("package: ", packageName, ", version: ", packageVersion, ", sync success")
		return true
	}

	fmt.Println("package: ", packageName, " ,version: ", packageVersion, " not found!")
	return false
}
