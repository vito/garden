package main

const (
	CLONE_VFORK   = 0x00004000
	CLONE_NEWNS   = 0x00020000
	CLONE_NEWUTS  = 0x04000000
	CLONE_NEWIPC  = 0x08000000
	CLONE_NEWUSER = 0x10000000
	CLONE_NEWPID  = 0x20000000
	CLONE_NEWNET  = 0x40000000
	SIGCHLD       = 0x14       /* Should set SIGCHLD for fork()-like behavior on Linux */
)
