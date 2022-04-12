
# WinSSH-Pageant ![Go](https://github.com/ndbeals/winssh-pageant/workflows/Go/badge.svg)

Proxy Pageant requests to the Windows OpenSSH agent (from Microsoft), enabling applications that only support Pageant to use openssh.

## Contents

- [Installation](#installation)
  - [Winget](#winget)
  - [MSI Installer](#msi-installer)
  - [Standalone Binary](#standalone-binary)
- [Usage](#usage)
  - [Autostart](#autostart)
- [Frequently Asked Questions](#frequently-asked-questions)
- [Building](#building)
  - [Easy Build Instructions](#easy-build-instructions)
  - [Advanced Build Instructions](#advanced-build-instructions)
  - [Antivirus Flagging](#antivirus-flagging)
- [Bug Reporting, Help & Feature Requests](#bug-reporting-help--feature-requests)
- [Credits](#credits)

## Background

I use the Windows OpenSSH agent as my single ssh key backing. Many solutions exist that do the opposite of this, but I prefer the convenience of Windows OpenSSH agent.

This has been tested on Windows 10 2004 using WSL2. Earlier versions of windows up to 1803 should work too.

# Installation

WinSSH-Pageant now features an MSI installer for easy upgrading and install/uninstall actions. The installer will:

 1. Create an appropriate, user-specific autostart entry which you can manager from Task Manager -> Startup tab
 2. Autostart the application after the installer is finished

### Prerequisites

Install the [Microsoft OpenSSH package, found on their Github](https://github.com/PowerShell/Win32-OpenSSH/releases). Do not install this using Windows update, that one is quite outdated, and will not work with this software. [Follow the instructions for adding OpenSSH to your System PATH.](https://github.com/PowerShell/Win32-OpenSSH/wiki/Install-Win32-OpenSSH-Using-MSI)

## Winget

This application is now available in the offcial [Microsoft Package Manager: `winget`](https://github.com/microsoft/winget-cli) and is the *preferred* way to install and upgrade this software.

| Action | Command |
| -----: |-------- |
| Install | `winget install winssh-pageant` |
| Upgrade | `winget upgrade winssh-pageant` |

## MSI Installer

[Download the latest version from the releases page](https://github.com/ndbeals/winssh-pageant/releases/latest) and install it.

### Standalone Binary

For those who do not want an installer, there is also an option to download the compiled, standalone executable. [Download the latest .zip](https://github.com/ndbeals/winssh-pageant/releases/latest) and follow the [instructions for configuring autostart](#autostart)

# Usage

Run the executable `winssh-pageant.exe`. There are two (optional) flags:

- `--sshpipe` - name of the windows openssh agent pipe, default is: `\\.\pipe\openssh-ssh-agent`
- `--no-pageant-pipe` - disable pageant named pipe proxying

## Autostart

Start Menu Autostart:

 1. Open Windows Explorer and navigate to:

    ```
    %appdata%\Microsoft\Windows\Start Menu\Programs\Startup
    ```

 2. Inside this folder, Create a shortcut pointing at wherever you put `winssh-pageant.exe`
 3. If the shortcut is valid, there should be a new `WinSSH-Pageant Bridge` entry found in Task Manager -> Startup

Note: Task Scheduler autostart method is now deprecated and unsupported. It causes possible issues with executable ownership.

---

# Frequently Asked Questions

## How do I add my private keys to pageant?
You don't. Add your private keys to the standard ssh-agent with the following command:
```
ssh-add <your key>
```
[Detailed explanation.](https://github.com/ndbeals/winssh-pageant/issues/14)

---

## Building

clone the repo:

```
git clone https://github.com/ndbeals/winssh-pageant.git
cd winssh-pageant
```

### Easy Build Instructions

```
go build -ldflags="-w -s -H=windowsgui" -trimpath
```

### Advanced Build Instructions

The build script `build.ps1` accepts numerous, optional flags and two dev dependencies. Install the dev dependencies:

```
go install github.com/josephspurrier/goversioninfo@latest
go install github.com/mh-cbon/go-msi@latest
```

Run the build script:

```
.\build.ps1 -Release
```

The release flag is *optional*, though highly recommended, if it is omitted you do not need the two aforementioned dev dependencies.

### Antivirus Flagging

Your antivirus software may flag this as malware, It's a false positive and a known quirk with go binaries (<https://golang.org/doc/faq#virus>). The official releases use reproducible builds via `-trimpath`. The expected checksums are posted with the release they're meant for, some users may choose to build this project themself and confirm the checksums, `sha256sum`.

More information can be found here: <https://github.com/ndbeals/winssh-pageant/issues/7#issuecomment-787520972>

## Bug Reporting, Help & Feature Requests

Please put report all

- Feature Requests
- Bugs
- Help Requests
- General Questsions

[as an issue.](https://github.com/ndbeals/winssh-pageant/issues)

## Credits

Big thanks to <https://github.com/benpye/wsl-ssh-pageant>, Ben Pye and the other contributors for the examples of interacting with the win32 api, the build script, and help they have given me directly.

- <https://github.com/buptczq/WinCryptSSHAgent> for a working example of how to open a file mapping another process created.
