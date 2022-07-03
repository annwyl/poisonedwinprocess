package main

import (
	//"C"
	"flag"
	"fmt"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

type Handle uintptr
type Dword uint32
type Word uint16
type Lpstr uintptr
type Lpvoid uintptr
type Lpbyte uintptr
type Process_Information_Class uintptr

type ProcessInformation struct {
	ProcessHandle            Handle
	ProcessInformationClass  Process_Information_Class
	ProcessInformation       Lpvoid
	ProcessInformationLength Dword
}

type StartupInfo struct {
	Cb              Dword
	lpReserved      Lpstr
	lpDesktop       Lpstr
	lpTitle         Lpstr
	dwX             Dword
	dwY             Dword
	dwXSize         Dword
	dwYSize         Dword
	dwXCountChars   Dword
	dwYCountChars   Dword
	dwFillAttribute Dword
	dwFlags         Dword
	wShowWindow     Word
	cbReserved2     Word
	lpReserved2     Lpbyte
	hStdInput       Handle
	hStdOutput      Handle
	hStdError       Handle
}

type SecurityAttributes struct {
	nLength              Dword
	lpSecurityDescriptor Lpvoid
	bInheritHandle       bool
}

type Flags struct {
	Process string
}

var (
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	createProcess = kernel32.NewProc("CreateProcessW")
	allocConsole  = kernel32.NewProc("AllocConsole")
	freeConsole   = kernel32.NewProc("FreeConsole")
)

func main() {
	// variables should be able to be set by the user later on
	globalFlags := Flags{}
	getFlags(&globalFlags)
	lpProcessAttributes := SecurityAttributes{}
	lpThreadAttributes := SecurityAttributes{}
	bInheritHandles := 0
	dwCreationFlags := 0
	lpStartupInfo := StartupInfo{}
	lpProcessInformation := ProcessInformation{}

	freeConsole.Call()
	allocConsole.Call()
	fmt.Println("Console allocated")

	res, _, _ := createProcess.Call(
		uintptr(unsafe.Pointer(nil)),
		stringToUTF16Ptr(globalFlags.Process),
		uintptr(unsafe.Pointer(&lpProcessAttributes)),
		uintptr(unsafe.Pointer(&lpThreadAttributes)),
		uintptr(bInheritHandles),
		uintptr(dwCreationFlags),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(&lpStartupInfo)),
		uintptr(unsafe.Pointer(&lpProcessInformation)),
	)
	fmt.Println(res)
}

func stringToUTF16Ptr(str string) uintptr {
	c := utf16.Encode([]rune(str + "\x00"))
	return uintptr(unsafe.Pointer(&c[0]))
}

func getFlags(flags *Flags) {
	flag.StringVar(&flags.Process, "path", "", "Process path")
	flag.Parse()
	if flags.Process == "" {
		fmt.Println("No path specified")
		os.Exit(1)
	}
}
