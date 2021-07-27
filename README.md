NOCLOSE
======
~~~
~~~

Utility for disabling buttons and items in the system menu of applications
---

```
    >>noclose -help
    noclose 1.0
    Author: Garry G.

    Disable the Close button (X) of selected window.

    Usage: noclose [-class class] [-title title] -disable command | -enable
    Command: SC_CLOSE, SC_SIZE, SC_MOVE, SC_MAXIMIZE, SC_MINIMIZE, SC_RESTORE

    If you start without specifying the class or title of the window
    the X button in the current console window will be blocked

    Example:
    noclose -class Notepad -disable SC_CLOSE
    noclose -class Notepad -disable SC_MAXIMIZE
    noclose -class Notepad -disable SC_MOVE
    noclose -class Notepad -enable
    noclose -disable SC_CLOSE
    noclose -enable

      -class string
            Window class name
      -disable string
            System command
      -enable
            Restore menu
      -title string
            Window title
      -v    Verbose output
```

Command line help
-----------------
***
**optional arguments:**


  * **-help**                  *Show this help message and exit*
  * **-v**                     *Enabling logging of all program actions.*
  * **-class window class**    *The class of the window in which you want to disable the buttons\system menu items.* 
  * **-title window title**    *The title of the window in which you want to disable the buttons\system menu items.* 
  
**required arguments:** 
  * **-disable system command**    *The command that will be disabled: SC_CLOSE, SC_SIZE, SC_MOVE, SC_MAXIMIZE, SC_MINIMIZE, SC_RESTORE.*  
  * **-enable**                    *Restoring the original menu*
  

~~~
~~~
EXAMPLES:  
=========

**for a window defined by a class**
```
noclose -class Notepad -disable SC_CLOSE
noclose -class Notepad -disable SC_MAXIMIZE
noclose -class Notepad -disable SC_MOVE
noclose -class Notepad -enable
```

**for the current console window**
```
noclose -disable SC_CLOSE
noclose -enable
```
  
  
  