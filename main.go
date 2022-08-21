package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd  := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// clone new uts container
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}


	cmd.Run()
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cgroups()

	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/ubuntu-fs")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd  := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr


	cmd.Run()

	syscall.Unmount("/proc", 0)
}

func cgroups() {
	cgroups := "/sys/fs/cgroup"
	pids := filepath.Join(cgroups, "pids")
	err := os.Mkdir(filepath.Join(pids, "my-container"), 0755)
	check(err)

	check(ioutil.WriteFile(filepath.Join(pids, "my-container/pids.max"), []byte("20"), 0644))
	check(ioutil.WriteFile(filepath.Join(pids, "my-container/notify_on_release"), []byte("1"), 0644))
	check(ioutil.WriteFile(filepath.Join(pids, "my-container/cgroup.procs"), []byte(fmt.Sprintf("%d", os.Getpid())), 0644))

}
 
func check(err error) {
	if err != nil {
		panic(err)
	}
}