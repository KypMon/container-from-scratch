
package main

import (
	`fmt`
	`io/ioutil`
	`os`
	`os/exec`
	`path/filepath`
	`strconv`
	`syscall`
)

const (
 	HOSTNAME = "HOSTNAME"
 	HOSTROOTPATH = "/home/temp"
)

// go run main.go run <cmd> <args>
func main() {

	initDefault()

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func initDefault() {
	defaultMap := make(map[string]string)

	// default value
	defaultMap["HOSTNAME"] = "container-by-scratch"

	for k, v := range defaultMap {
		checkEnvExist(k, v)
	}
}

func checkEnvExist (envName string, defaultVal string) {
	_, present := os.LookupEnv(envName)
	if !present {
		os.Setenv(envName, defaultVal)
	}
}

func run ()  {

	fmt.Printf("RUnning %v\n", os.Args[2:])

	//cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// create new namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// UTS namespace => hostname
		// PID namespace => process
		// NEWUSER => rootless container
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		UidMappings: []syscall.SysProcIDMap{
			{containerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), size: 1},
		},
	}

	// trigger cli to running
	// we don't create the namespace before we finally run this!
	// so we need to create the namespace first
	must(cmd.Run())
}

// bash picked up the hostname before we changed it, so that's why we need to have a child process before running bash
func child () {

	cg()

	fmt.Printf("RUnning %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// set hostname
	setHostname(os.Getenv(HOSTNAME))

	// limit the access using the chroot
	setRootFS(os.Getenv(HOSTROOTPATH))

	// processs
	setProcessMount()

	// FS
	setFSMount()

	// trigger cli to running
	// we don't create the namespace before we finally run this!
	// so we need to create the namespace first
	must(cmd.Run())

	// after execution
	setProcessUnmount()
	setFSUnmount()
}

func setHostname (hostname string) {
	must(syscall.Sethostname([]byte(hostname)))
}

func setRootFS (rootpath string) {
	must(syscall.Chroot(rootpath))
	must(os.Chdir(rootpath))
}

func setProcessMount () {
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
}

func setProcessUnmount () {
	must(syscall.Unmount("proc", 0))
}

func setFSMount () {
	// source string, target string, fstype string, flags uintptr, data string
	must(syscall.Mount("something", "mytemp", "tmpfs", 0, ""))
}

func setFSUnmount () {
	must(syscall.Unmount("mytemp", 0))
}

// Cgroup
func cg () {
	cgroups := "/sys/fs/cgroup/"

	mem := filepath.Join(cgroups, "memory")
	os.Mkdir(filepath.Join(mem, "ubuntu"), 0755)
	// memory limit
	must(ioutil.WriteFile(filepath.Join(mem, "ubuntu/memory.limit_in_bytes"), []byte("999424"), 0700))
	// swap memory
	must(ioutil.WriteFile(filepath.Join(mem, "ubuntu/memory.memsw.limit_in_bytes"), []byte("999424"), 0700))
	// tidy up no longer needed
	must(ioutil.WriteFile(filepath.Join(mem, "ubuntu/memory.notify_on_release"), []byte("999424"), 0700))

	// get the current process id and write into the cgroup
	pid := strconv.Itoa(os.Getppid())
	must(ioutil.WriteFile(filepath.Join(mem, "ubutu/cgroup.procs"), []byte(pid), 0700))
}


// ERR HANDLING
func must(err error)  {
	if err != nil {
		panic(err)
	}
}