![Go](https://github.com/ndbeals/winssh-pageant/workflows/Go/badge.svg)
# winssh-pageant
Proxy Pageant requests to the Windows OpenSSH agent (from Microsoft), enabling applications that only support Pageant to use openssh.


## Background
I use the Windows OpenSSH agent as my single ssh key backing. Many solutions exist that do the opposite of this, but I prefer the convenience of Windows OpenSSH agent.

This has been tested on Windows 10 2004 using WSL2. Earlier versions of windows up to 1803 should work too.


## Installation
Install the [Microsoft OpenSSH package, found on their Github](https://github.com/PowerShell/Win32-OpenSSH/releases). Do not install this using Windows update, that one is quite outdated, and will not work with this software.

Download a compiled binary from the [releases page](https://github.com/ndbeals/winssh-pageant/releases) or otherwise build it yourself.

### Building
clone the repo and build it:
```
git clone https://github.com/ndbeals/winssh-pageant.git
cd winssh-pageant
go build -ldflags -H=windowsgui
```


## Usage
Run the executable `winssh-pageant.exe`. There are two (optional) flags:

 - `--sshpipe` - name of the windows openssh agent pipe, default is `"\\.\pipe\ssh-pageant"`
 - `--no-pageant-pipe` - disable pageant named pipe proxying


### Task Scheduler
Until I decide on a better way to do this, you can auto-start this program by creating a task using the Task Scheduler, here are the basic steps:

1. Start the Task Scheduler, on the left pane, select "Task Scheduler Library"
2. Create a Basic Task named `winssh-pageant`
3. Set the trigger to "When I log on"
4. Set the action to "Start a program"
5. Locate and select the `winssh-pageant.exe` executable
6. Finish and run the task (or otherwise log out and back in)


## Bug Reporting, Help & Feature Requests
Please put report all
 - Feature Requests
 - Bugs
 - Help Requests
 - General Questsions
 
[as an issue.](https://github.com/ndbeals/winssh-pageant/issues)


## Credits
Big thanks to https://github.com/benpye/wsl-ssh-pageant, Ben Pye and the other contributors for the examples of interacting with the win32 api, the build script, and help they have given me directly.
- https://github.com/buptczq/WinCryptSSHAgent for a working example of how to open a file mapping another process created.
