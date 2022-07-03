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

//type Lpdword uintptr
//type Lpcstr uintptr
//type Ulong uint32
//type Ulong64 uint64

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
)

func main() {
	globalFlags := Flags{}
	getFlags(&globalFlags)
	lpProcessAttributes := SecurityAttributes{}
	lpThreadAttributes := SecurityAttributes{}
	bInheritHandles := 0
	dwCreationFlags := 0
	lpStartupInfo := StartupInfo{}
	lpProcessInformation := ProcessInformation{}

	//sysCall()

	res, _, err := syscall.Syscall12(
		createProcess.Addr(),
		uintptr(0),
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
		uintptr(0),
		uintptr(0),
	)
	if err != 0 {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(res)
}

func sysCall() {
	// make the windowsapi call, i have no clue how to implement this correctly rn (without syscall)
	// probably do some C shenanigans, shouldnt be to hard. maybe something along the lines of:

	/*
		//#include <windows.h>
		import "C"
		...
		C.CreateProcess(
			C.CString(...),
	*/
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
