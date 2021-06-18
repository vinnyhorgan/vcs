package main

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("No command was specified. Use -h to see a list of all the possible commands you can use.")

        return
    }

    command := os.Args[1]

    if command == "-h" || command == "-help" {
        fmt.Println("COMMANDS\n- init [folder] => initializes the specified folder as a govcs repository.\n- checkout [branch-name] => if the specified branch doesn't exist, it will create it and point at it, else it will just point at it.\n- commit [message] => saves the current state of the folder as a commit in the branch it's pointing at with the specified message.\n- revert [commit-id] => reverts the project to the state it was at the specified commit of the branch it's pointing at.\n- log [branch-name] => logs all of the commits of the specified branch.")
    } else if command == "init" {
    	initialze()
    } else if command == "checkout" {
        checkout()
    } else if command == "commit" {
        commit()
    } else if command == "revert" {
        revert()
    } else if command == "log" {
        log()
    } else {
        fmt.Println("Invalid command. Use -h to see a list of all the possible commands you can use.")
    }
}

func initialze() {
    name := getargument()

    check := getdir() + "/" + name

    info, err := os.Stat(check)

    if err != nil {
        fmt.Println("Please choose a valid folder.")

        return
    }

    if info.IsDir() {
        mkdir(getdir() + "/.vcs")

        NAME := create(getdir() + "/.vcs/NAME")
        write(NAME, name)
        NAME.Close()

        create_branch("master")

        POINTER := create(getdir() + "/.vcs/POINTER")
        write(POINTER, "master")
        POINTER.Close()

        fmt.Println("The folder was successfully initialized as a govcs repository.")
    } else {
        fmt.Println("Please choose a valid folder.")

        return
    }
}

func checkout() {
    branch := getargument()

    check := getdir() + "/.vcs/" + branch

    info, err := os.Stat(check)

    if err != nil {
        create_branch(branch)

        fmt.Println("Created " + branch + " branch.")

        POINTER := create(getdir() + "/.vcs/POINTER")
        write(POINTER, branch)
        POINTER.Close()

        fmt.Println("Now pointing at " + branch + " branch.")

        return
    }

    if info.IsDir() {
        POINTER := create(getdir() + "/.vcs/POINTER")
        write(POINTER, branch)
        POINTER.Close()

        fmt.Println("Now pointing at " + branch + " branch.")
    }
}

func commit() {
    message := getargument()

    branch := read(getdir() + "/.vcs/POINTER")

    commit := read(getdir() + "/.vcs/" + branch + "/COMMITS")

    name := read(getdir() + "/.vcs/NAME")

    path := getdir() + "/.vcs/" + branch + "/" + commit + "/" + name

    err := CopyDir(getdir() + "/" + name, path)

    if err != nil {
        fmt.Println(err)
    }

    MESSAGE := create(getdir() + "/.vcs/" + branch + "/" + commit + "/MESSAGE")
    write(MESSAGE, message)
    MESSAGE.Close()

    i, _ := strconv.Atoi(commit)
    i++

    COMMITS := create(getdir() + "/.vcs/" + branch + "/COMMITS")
    write(COMMITS, strconv.Itoa(i))
    COMMITS.Close()

    fmt.Println("Successfully committed the " + name + " folder to " + branch + " branch, as commit number " + strconv.Itoa(i) + ".")
}

func revert() {
    commit := getargument()

    branch := read(getdir() + "/.vcs/POINTER")

    name := read(getdir() + "/.vcs/NAME")

    path := getdir() + "/.vcs/" + branch + "/" + commit + "/" + name

    err_1 := os.RemoveAll(getdir() + "/" + name)

    if err_1 != nil {
        fmt.Println(err_1)
    }

    err_2 := CopyDir(path, getdir() + "/" + name)

    if err_2 != nil {
        fmt.Println(err_2)
    }

    fmt.Println("Successfully reverted the project to commit " + commit + " from branch " + branch + ".")
}

func log() {
    branch := getargument()

    path := getdir() + "/.vcs/" + branch

    commits := read(path + "/COMMITS")

    integer, _ := strconv.Atoi(commits)

    if integer == 0 {
        fmt.Println("No commits in this branch...yet.")
    }

    for i := 0; i < integer; i++ {
        message := read(path + "/" + strconv.Itoa(i) + "/MESSAGE")

        fmt.Println("COMMIT: " + strconv.Itoa(i) + " ==> " + message)
    }
}

func getdir() string {
    dir, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
	}

    return dir
}

func mkdir(path string) {
    err := os.Mkdir(path, 0755)

    if err != nil {
        fmt.Println(err)
    }
}

func create(path string) *os.File {
    file, err := os.Create(path)

    if err != nil {
        fmt.Println(err)
    }

    return file
}

func write(file *os.File, text string) {
    _, err := file.WriteString(text)

    if err != nil {
        fmt.Println(err)
    }
}

func read(path string) string {
    data, err := ioutil.ReadFile(path)

    if err != nil {
        fmt.Println(err)
    }

    return string(data)
}

func create_branch(name string) {
    path := getdir() + "/.vcs/" + name

    mkdir(path)

    COMMITS := create(path + "/COMMITS")
    write(COMMITS, "0")
    COMMITS.Close()
}

func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func getargument() string {
    if len(os.Args) != 3 {
        fmt.Println("No argument was specified. Use -h to see a list of all the possible commands you can use.")

        os.Exit(1)
    }

    return os.Args[2]
}
