package main

import (
    "io/ioutil"
    "os/user"
    "runtime"
    "strconv"
    "strings"
    "regexp"
    "fmt"
    "log"
    "os"

    cf "updateMinecraftMods/curseForge"
)

func matchModFilename(text string) (string) {
    m := []string{
        `^(\w+)[\+\-\_](\w+)[\+\-\_]\d`,
        `^(\w+)[\+\-\_]\d`,
        `^(\w+)[\+\-\_]\w\d`,
        `^(\w+)[\+\-\_]mc\d`,
    }
    for i := 0; i < len(m); i++ {
        re := regexp.MustCompile(m[i])
        match := re.FindStringSubmatch(text)
        if len(match) == 2 {
            return match[1]
        } else if len(match) >= 3 {
            return match[1] + "+" + match[2]
        }
    }
    return ""
}
func searching(text,pattern string) (bool) {
    getWord := func(text string) []string {
        res := []string{""}
        index := 0
        s := strings.Split(text,"")
        for i := 0; i < len(s); i++ {
            re := regexp.MustCompile(`[\!\"\#\$\%\&\'\(\)\*\+\-\.\/\:\;\<\=\>\?\@\[\]\^\_\{\}\|\~]`)
            match := re.FindStringSubmatch(s[i])
            if len(match) != 0 {
                index = index + 1
                res = append(res,"")
            } else {
                res[index] = res[index] + s[i]
            }
        }
        return res
    }
    finding := 0
    l := getWord(pattern)
    for i := 0; i < len(l); i++ { 
        f,err := regexp.Match(strings.ToLower(l[i]),[]byte(strings.ToLower(text)))
        if f && err == nil {
            finding += 1
        }
    }
    return (finding==len(l))//,finding,len(l)
}

var PATH_MODS string
var UPDATING_LIST []string

func main() {
//  "/home/vespan/.minecraft/mods/"

    checkDefaultPath := func() {
        for {
            GOOS := runtime.GOOS
            username,err := user.Current()
            if err != nil {
                log.Fatal(err)
            }
            paths := map[string]string{
                "linux": username.HomeDir + "/.minecraft/mods/",
                "windows": username.HomeDir + "/AppData/Roaming/.minecraft",
            }
            _,exits := paths[GOOS]
            if _, err := os.Open(paths[GOOS]); exits && err == nil {
                fmt.Println("i found the default path mods!("+paths[GOOS]+").Enter \"Y\" for select or enter your path.")
                fmt.Scan(&PATH_MODS)
                if strings.ToLower(PATH_MODS) == "y" {
                    PATH_MODS = paths[GOOS]
                }
            } else {
                fmt.Println("enter your path.")
                fmt.Scan(&PATH_MODS)
            }
            if _, err := os.Open(PATH_MODS); err != nil {
                fmt.Println("This path was not found!")
            } else {
                break
            }
        }
    }
    checkDefaultPath()


    _, err := os.Open("mods/")
    if err != nil {
        if err := os.Mkdir("mods/", os.ModePerm); err != nil {
            log.Fatal(err)
        }
    }

    files, err := ioutil.ReadDir(PATH_MODS)
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
        if file.IsDir() == false {
            filename := file.Name()
            if matchModFilename(filename) == "" {
                fmt.Println(filename,"NIL MATCHED!")
                moveMod(filename)
            } else {

                j,err := cf.SearchMod(matchModFilename(filename))
                if err != nil {
                    log.Fatal(err)
                }
                res := parsing(*j,filename) 
                if !res {
                    moveMod(filename)
                }
            }
        }
    }

    fmt.Println("\n\nupdating mods\n")
    for i := 0; i < len(UPDATING_LIST); i++ {
        fmt.Println(UPDATING_LIST[i])
    }

}

func moveMod (filename string) {
    content,err := ioutil.ReadFile(PATH_MODS+filename)
    if err != nil {
        log.Fatal(err)
    }
    ioutil.WriteFile("mods/"+filename, content, 0644)
}

func parsing(j cf.SearchMod_root__struct, filename string) bool {

    fmt.Println("parsing..",filename)

    getNewMod := func(modId int, filename string) (int,error) {


        files,err := cf.GetFiles(strconv.Itoa(modId))
        // files := *filesC
        if err != nil {
            return 0,err
        }
        for i := 0; i < len(files.Data); i++ {
            if strings.ToLower(files.Data[i].FileName) == strings.ToLower(filename) {
                gameVersion := files.Data[i].GameVersions[len(files.Data[i].GameVersions)-1]

                for l := 0; l < len(files.Data); l++ {
                    if l >= i {
                        moveMod(filename)
                        return 1,nil
                    } else {
                        for g := 0; g < len(files.Data[l].GameVersions); g++ {
                            if files.Data[l].GameVersions[g] == gameVersion && strings.ToLower(files.Data[l].FileName) != strings.ToLower(filename) {
                                
                                UPDATING_LIST = append(UPDATING_LIST,filename+"-->"+files.Data[l].FileName)
                                fmt.Println(filename,"new version of mod",files.Data[l].FileName)
                                _, err := cf.Download(strconv.Itoa(modId),strconv.Itoa(files.Data[l].Id),files.Data[l].FileName, "mods/")
                                if err != nil {
                                    fmt.Println(filename,"error downloading:(")
                                    return 0,err
                                }
                                return 2,nil
                            }
                        }
                    }
                }
                moveMod(filename)
                return 1,nil
            }  
        }
        moveMod(filename)
        return 0,nil
    }

    getNewMod_map := map[int]string{
        0: "not found",
        1: "you have the current version",
        2: "downloading the current version...",
    }

    for i := 0; i < len(j.Data); i++ {
        data := j.Data[i]
        if strings.ToLower(data.Slug) == strings.ToLower(matchModFilename(filename)) || strings.ToLower(data.Name) == strings.ToLower(matchModFilename(filename)) {
            ind,err := getNewMod(data.Id,filename)
            if err != nil {
                log.Fatal(err,j.Data[i].Id)
                return false
            }
            fmt.Println(filename,getNewMod_map[ind])
            return true
        } else {
            if j.Data[i].Class.Name == "Mods" {
                files,err := cf.GetFiles(string(j.Data[i].Id))
                if err != nil {
                    log.Fatal(err,j.Data[i].Id)
                    return false
                }
                for l := 0; l < len(files.Data); l++ {
                    if searching(files.Data[i].FileName,filename) || searching(files.Data[i].FileName,matchModFilename(filename)) || strings.ToLower(files.Data[i].FileName) == strings.ToLower(filename) {
                        ind,err := getNewMod(data.Id,filename)
                        if err != nil {
                            log.Fatal(err,j.Data[i].Id)
                            return false
                        }
                        fmt.Println(filename,getNewMod_map[ind])
                        return true
                    }
                }
            }

        }
    }
    return false
}
