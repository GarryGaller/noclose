package main

import (
	//"errors"
	"flag"
	"fmt"
	"golang.org/x/sys/windows"
	"log"
	"os"
	"path/filepath"
    "strconv"
	"strings"
	"syscall"
	"unsafe"
)

type Opts struct {
	class, title string
	disable      string
	enable       bool
	verbose      bool
}

const MF_BYCOMMAND = 0
const MF_BYPOSITION = 1
const MF_DISABLED = 2 // Указывает, что пункт меню заблокирован, а не недоступный, так что он не может быть выбран.
const MF_ENABLED = 0  // Указывает, что пункт меню включен и восстановлен из недоступного состояния так, чтобы он может быть выбран.
const MF_GRAYED = 1   // Указывает, что пункт меню заблокирован и недоступен так, чтобы его невозможно выбрать.

var SYSCOMMAND = map[string]int{
	"SC_CLOSE": 0xF060, // Closes the window.
	"SC_SIZE":  0xF000, // Sizes the window.
	"SC_MOVE":  0xF010, // Moves the window.
	//"SC_CONTEXTHELP":0xF180,   // Changes the cursor to a question mark with a pointer. If the user then clicks a control in the dialog box, the control receives a WM_HELP message.
	//"SC_DEFAULT":0xF160,       // Selects the default item; the user double-clicked the window menu.
	//"SC_HOTKEY":0xF150,        // Activates the window associated with the application-specified hot key. The lParam parameter identifies the window to activate.
	//"SC_HSCROLL":0xF080,       // Scrolls horizontally.
	//"SCF_ISSECURE":0x00000001, // Indicates whether the screen saver is secure.
	//"SC_KEYMENU": 0xF100,      // Retrieves the window menu as a result of a keystroke. For more information, see the Remarks section.
	"SC_MAXIMIZE": 0xF030, // Maximizes the window.
	"SC_MINIMIZE": 0xF020, // Minimizes the window.
	//"SC_MONITORPOWER": 0xF170, // Sets the state of the display. This command supports devices that have power-saving features, such as a battery-powered personal computer.
	//"SC_MOUSEMENU" :0xF090,    // Retrieves the window menu as a result of a mouse click.
	//"SC_NEXTWINDOW":0xF040,    // Moves to the next window.
	//"SC_PREVWINDOW":0xF050,    // Moves to the previous window.
	"SC_RESTORE": 0xF120, //Restores the window to its normal position and size.
	//"SC_SCREENSAVE":0xF140,    // Executes the screen saver application specified in the [boot] section of the System.ini file.
	//"SC_TASKLIST": 0xF130,     // Activates the Start menu.
	//"SC_VSCROLL":0xF070,       // Scrolls vertically.
}

var REVERSE_SYSCOMMAND = RevSysCommand()

func RevSysCommand() map[int]string {
	var reverse_syscommand = make(map[int]string)
	for k, v := range SYSCOMMAND {
		reverse_syscommand[v] = k
	}

	return reverse_syscommand
}
func HexToInt(s string) (uint, bool) {
	s = strings.ToUpper(s)
	s = strings.Replace(s, "0X", "", -1)
	n, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, false
	}
	return uint(n), true
}

func (opts *Opts) Usage() {
	fmt.Printf(
		("noclose 1.0\n" +
			"Author: Garry G.\n\n" +
			"Disable the Close button (X) of selected window.\n\n" +
			"Usage: %s [-class class] [-title title] -disable command | -enable \n" +
			"Command: SC_CLOSE, SC_SIZE, SC_MOVE, SC_MAXIMIZE, SC_MINIMIZE, SC_RESTORE\n\n" +
			"If you start without specifying the class or title of the window\n" +
			"the X button in the current console window will be blocked\n\n" +
			"Example: \n" +
			"noclose -class Notepad -disable SC_CLOSE\n" +
			"noclose -class Notepad -disable SC_MAXIMIZE\n" +
			"noclose -class Notepad -disable SC_MOVE\n" +
			"noclose -class Notepad -enable\n" +
			"noclose -disable SC_CLOSE\n" +
			"noclose -enable\n" +
			"\n"),
		filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func (opts *Opts) Parse() *Opts {
	flag.Usage = opts.Usage
	flag.StringVar(&opts.class, "class", "", "Window class name")
	flag.StringVar(&opts.title, "title", "", "Window title")
	flag.StringVar(&opts.disable, "disable", "", "System command")
	flag.BoolVar(&opts.enable, "enable", false, "Restore menu")
	flag.BoolVar(&opts.verbose, "v", false, "Verbose output")
	flag.Parse()
	return opts
}

func getCommand(cmd string) (int, error) {
	var lastErr error

	if cmd == "" {
		return SYSCOMMAND["SC_CLOSE"], nil
	}

	cmdStr := strings.ToUpper(cmd)
	cmdHex, isHex := HexToInt(cmdStr)
	if !isHex {
		if !strings.HasPrefix(cmdStr, "SC_") {
			cmdStr = "SC_" + cmdStr
		}
	} else {
		cmdStr, _ = REVERSE_SYSCOMMAND[int(cmdHex)]
	}

	command, ok := SYSCOMMAND[cmdStr]
	if !ok {
		return 0, fmt.Errorf("Sys command <%s> not found", cmd)
	}

	return command, lastErr
}

func main() {
	opts := &Opts{}
	opts.Parse()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	user32 := windows.MustLoadDLL("user32.dll")
	defer user32.Release()
	kernel32 := windows.MustLoadDLL("kernel32.dll")
	defer kernel32.Release()

	DeleteMenu := user32.MustFindProc("DeleteMenu")
	DrawMenuBar := user32.MustFindProc("DrawMenuBar")
	GetSystemMenu := user32.MustFindProc("GetSystemMenu")
	GetConsoleWindow := kernel32.MustFindProc("GetConsoleWindow")
	FindWindow := user32.MustFindProc("FindWindowW")
	CloseHandle := kernel32.MustFindProc("CloseHandle")

	var hWnd uintptr
	var lastErr error
	var pWindowName unsafe.Pointer
	var pClassName unsafe.Pointer
	var class *uint16 = nil
	var window *uint16 = nil
	var command int

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if opts.class == "" && opts.title == "" {
		hWnd, _, lastErr = GetConsoleWindow.Call()
	} else {

		if opts.class != "" {
			class = syscall.StringToUTF16Ptr(opts.class)
		}
		if opts.title != "" {
			window = syscall.StringToUTF16Ptr(opts.title)
		}

		pClassName = unsafe.Pointer(class)
		pWindowName = unsafe.Pointer(window)

		hWnd, _, lastErr = FindWindow.Call(
			uintptr(pClassName),
			uintptr(pWindowName),
		)

		if opts.verbose {
			log.Printf("FindWindow       :[HWND:%v] Class:%s Title:%s [Err:%s]\n",
				hWnd, opts.class, opts.title, lastErr)
		}

	}

	if hWnd != 0 {
		defer CloseHandle.Call(hWnd)
		hSysMenu, _, lastErr := GetSystemMenu.Call(hWnd, uintptr(0))
		if opts.verbose {
			log.Printf("GetSystemMenu    :[HWND:%v] [Err:%s]\n", hSysMenu, lastErr)
		}

		if hSysMenu != 0 {
			defer CloseHandle.Call(hSysMenu)

			if opts.disable != "" {
				command, lastErr = getCommand(opts.disable)
				if opts.verbose {
					log.Printf("GET SYSCOMMAND   :%v %#x [Err:%v]\n",
						opts.disable, command, lastErr)
				}

				if lastErr == nil {
					result, _, lastErr := DeleteMenu.Call(
						hSysMenu,
						uintptr(command),
						uintptr(MF_BYCOMMAND))

					if opts.verbose {
						log.Printf("DeleteMenu       :[RESULT:%v] [Err:%s]\n",
							result, lastErr)
					}
					result, _, _ = DrawMenuBar.Call(hWnd)
					if opts.verbose {
						log.Printf("DrawMenuBar      :[RESULT:%v] [Err:%s]\n",
							result, lastErr)
					}
				}
			}

			if opts.enable {
				result, _, lastErr := GetSystemMenu.Call(hWnd, uintptr(1))
				if opts.verbose {
					log.Printf("Revert SystemMenu:[RESULT:%v] [Err:%s]\n",
						result, lastErr)
				}

				result, _, lastErr = DrawMenuBar.Call(hWnd)
				if opts.verbose {
					log.Printf("DrawMenuBar      :[RESULT:%v] [Err:%s]\n",
						result, lastErr)
				}
			}
		}
	}
}
