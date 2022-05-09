
package main

import (
	`fmt`
	`os`
	`os/exec`
	`syscall`
)

const HOSTNAME string = "HOSTNAME"

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
		Cloneflags: syscall.CLONE_NEWUTS, // UTS namespace => hostname
	}

	// trigger cli to running
	// we don't create the namespace before we finally run this!
	// so we need to create the namespace first
	must(cmd.Run())
}

// bash picked up the hostname before we changed it, so that's why we need to have a child process before running bash
func child () {

	fmt.Printf("RUnning %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	setHostname(os.Getenv(HOSTNAME))

	// trigger cli to running
	// we don't create the namespace before we finally run this!
	// so we need to create the namespace first
	must(cmd.Run())
}

func setHostname (hostname string) {
	must(syscall.Sethostname([]byte(hostname)))
}

func must(err error)  {
	if err != nil {
		panic(err)
	}
}