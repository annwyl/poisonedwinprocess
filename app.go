package main

import (
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
	ProcessHandle Handle
	ThreadHandle  Handle
	ProcessId     Dword
	ThreadId      Dword
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
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	createProcess           = kernel32.NewProc("CreateProcessW")
	allocConsole            = kernel32.NewProc("AllocConsole")
	freeConsole             = kernel32.NewProc("FreeConsole")
	suspendThread           = kernel32.NewProc("SuspendThread")
	resumeThread            = kernel32.NewProc("ResumeThread")
	terminateProcess        = kernel32.NewProc("TerminateProcess")
	setProcessPriorityClass = kernel32.NewProc("SetPriorityClass")
	writeProcessMemory      = kernel32.NewProc("WriteProcessMemory")
	readProcessMemory       = kernel32.NewProc("ReadProcessMemory")
	virtualAllocEx          = kernel32.NewProc("VirtualAllocEx")
	createRemoteThread      = kernel32.NewProc("CreateRemoteThread")
)

func main() {
	globalFlags := Flags{}
	getFlags(&globalFlags)

	lpProcessAttributes := SecurityAttributes{}
	lpThreadAttributes := SecurityAttributes{}
	bInheritHandles := 0
	dwCreationFlags := 0
	lpStartupInfo := StartupInfo{}
	var lpProcessInformation ProcessInformation

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

	if res == 0 {
		fmt.Println("Failed to create process")
		return
	}

	fmt.Printf("Process created PID: %d\n", lpProcessInformation.ProcessId)

	for {
		fmt.Println("\nChoose an action:")
		fmt.Println("1. Suspend process")
		fmt.Println("2. Resume process")
		fmt.Println("3. Terminate process")
		fmt.Println("4. Set process priority")
		fmt.Println("5. Inject DLL")
		fmt.Println("6. Read process memory")
		fmt.Println("7. Write process memory")
		fmt.Println("8. Exit")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			suspendProcess(lpProcessInformation.ThreadHandle)
		case 2:
			resumeProcess(lpProcessInformation.ThreadHandle)
		case 3:
			terminateProcessFunc(lpProcessInformation.ProcessHandle)
			return
		case 4:
			setProcessPriority(lpProcessInformation.ProcessHandle)
		case 5:
			injectDLL(lpProcessInformation.ProcessHandle)
		case 6:
			readProcessMemoryFunc(lpProcessInformation.ProcessHandle)
		case 7:
			writeProcessMemoryFunc(lpProcessInformation.ProcessHandle)
		case 8:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func suspendProcess(threadHandle Handle) {
	ret, _, _ := suspendThread.Call(uintptr(threadHandle))
	if ret == 0xFFFFFFFF {
		fmt.Println("Failed to suspend process")
	} else {
		fmt.Println("Process suspended")
	}
}

func resumeProcess(threadHandle Handle) {
	ret, _, _ := resumeThread.Call(uintptr(threadHandle))
	if ret == 0xFFFFFFFF {
		fmt.Println("Failed to resume process")
	} else {
		fmt.Println("Process resumed")
	}
}

func terminateProcessFunc(processHandle Handle) {
	ret, _, _ := terminateProcess.Call(uintptr(processHandle), 0)
	if ret == 0 {
		fmt.Println("Failed to terminate process")
	} else {
		fmt.Println("Process terminated")
	}
}

func setProcessPriority(processHandle Handle) {
	fmt.Println("Choose priority:")
	fmt.Println("1. Idle")
	fmt.Println("2. Below Normal")
	fmt.Println("3. Normal")
	fmt.Println("4. Above Normal")
	fmt.Println("5. High")
	fmt.Println("6. Realtime")

	var choice int
	fmt.Scanln(&choice)

	var priorityClass uintptr
	switch choice {
	case 1:
		priorityClass = 0x00000040 // IDLE_PRIORITY_CLASS
	case 2:
		priorityClass = 0x00004000 // BELOW_NORMAL_PRIORITY_CLASS
	case 3:
		priorityClass = 0x00000020 // NORMAL_PRIORITY_CLASS
	case 4:
		priorityClass = 0x00008000 // ABOVE_NORMAL_PRIORITY_CLASS
	case 5:
		priorityClass = 0x00000080 // HIGH_PRIORITY_CLASS
	case 6:
		priorityClass = 0x00000100 // REALTIME_PRIORITY_CLASS
	default:
		fmt.Println("Invalid choice")
		return
	}

	ret, _, _ := setProcessPriorityClass.Call(uintptr(processHandle), priorityClass)
	if ret == 0 {
		fmt.Println("Failed to set priority")
	} else {
		fmt.Println("priority set")
	}
}

func injectDLL(processHandle Handle) {
	fmt.Println("Enter path to DLL:")
	var dllPath string
	fmt.Scanln(&dllPath)

	dllPathUTF16 := stringToUTF16Ptr(dllPath)
	dllPathSize := uintptr(len(dllPath) * 2)

	// no idea what im doing really
	addr, _, _ := virtualAllocEx.Call(
		uintptr(processHandle),
		0,
		dllPathSize,
		0x1000|0x2000, // MEM_COMMIT | MEM_RESERVE
		0x40,          // PAGE_EXECUTE_READWRITE
	)

	if addr == 0 {
		fmt.Println("Failed to allocate memory")
		return
	}

	_, _, _ = writeProcessMemory.Call(
		uintptr(processHandle),
		addr,
		dllPathUTF16,
		dllPathSize,
		0,
	)

	kernel32Handle, _ := syscall.LoadLibrary("kernel32.dll")
	loadLibraryAddr, _ := syscall.GetProcAddress(kernel32Handle, "LoadLibraryW")

	_, _, _ = createRemoteThread.Call(
		uintptr(processHandle),
		0,
		0,
		uintptr(loadLibraryAddr),
		addr,
		0,
		0,
	)

	fmt.Println("DLL injected")
}

func readProcessMemoryFunc(processHandle Handle) {
	var address uintptr
	var size uintptr

	fmt.Print("Enter memory address (in hexadecimal): 0x")
	fmt.Scanf("%x", &address)

	fmt.Print("Enter bytes to read: ")
	fmt.Scanf("%d", &size)

	buffer := make([]byte, size)
	var bytesRead uintptr

	ret, _, _ := readProcessMemory.Call(
		uintptr(processHandle),
		address,
		uintptr(unsafe.Pointer(&buffer[0])),
		size,
		uintptr(unsafe.Pointer(&bytesRead)),
	)

	if ret == 0 {
		fmt.Println("Failed to read process memory")
		return
	}

	fmt.Printf("Read %d bytes:\n", bytesRead)
	fmt.Printf("%x\n", buffer)
}

func writeProcessMemoryFunc(processHandle Handle) {
	var address uintptr
	var value uint64

	fmt.Print("Enter memory address (in hexadecimal): 0x")
	fmt.Scanf("%x", &address)

	fmt.Print("Enter value (in hexadecimal): 0x")
	fmt.Scanf("%x", &value)

	size := unsafe.Sizeof(value)
	var bytesWritten uintptr

	ret, _, _ := writeProcessMemory.Call(
		uintptr(processHandle),
		address,
		uintptr(unsafe.Pointer(&value)),
		size,
		uintptr(unsafe.Pointer(&bytesWritten)),
	)

	if ret == 0 {
		fmt.Println("Failed to write process memory")
		return
	}

	fmt.Printf("Wrote %d bytes\n", bytesWritten)
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
