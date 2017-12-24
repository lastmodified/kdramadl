# kdramadl

 Alternative downloader for [https://goplay.anontpp.com](https://goplay.anontpp.com) (via [/r/koreanvariety](https://www.reddit.com/r/koreanvariety/comments/723mtd/i_created_this_website_that_streams_korean_shows/?sort=new) or [/r/KDRAMA/](https://www.reddit.com/r/KDRAMA/comments/723n1y/i_created_this_website_that_streams_korean_shows/)).

## Install

Download and extract the [latest release](https://github.com/lastmodified/kdramadl/releases/latest) for your OS.

If you do not have ``ffmpeg`` already installed, choose the ``*_ffmpeg.zip`` version (e.g.  ``kdramadl_windows_32bit_ffmpeg.zip``).

## Usage

You can launch the downloader by double-clicking on ``kdramadl.exe`` / ``kdramadl`` in Windows Explorer / Finder.

Alternatively, you may launch it from the Command Prompt / Terminal.

```
NAME:
   kdramadl - Alternative downloader for https://goplay.anontpp.com

USAGE:
   kdramadl [global options] command [command options] [arguments...]

VERSION:
   0.1.8

DESCRIPTION:
   Make sure you have ffmpeg installed in PATH or in the current folder.

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -c value, --code value        Download Code
   -r value, --resolution value  Resolution of video, for example: 720p.
   -f value, --format value      Video format. Choose from: "mkv" "mp4". Default is "mkv".
   --filename value              Filename to save as (without extension).
   --sub                         Download only subtitles.
   --hardsubs                    Enable hard subs (for mp4 only).
   --hardsubsstyle value         Custom hard subs font style, e.g. To make subs blue and font size 22 'FontSize=22,PrimaryColour=&H00FF0000' (default: "PrimaryColour=&H0000FFFF")
   --ffmpeg value                Path to ffmpeg executable. (default: "ffmpeg")
   --folder value                Path to download folder.
   --alt                         Use kdrama.armsasuncion.com instead of goplay.anontpp.com
   --proxy value                 Proxy address (only HTTP proxies supported), example "http://127.0.0.1:80".
   --timeout value               Connection timeout interval in seconds. Default 10. (default: 10)
   --autoquit                    Automatically quit when done (skip the "Press ENTER to continue" prompt)
   --nocolor                     Disable color output
   --verbose                     Generate more verbose messages
   --logfile value               Path to logfile (for debugging/reporting)
   --config value                Path to custom yaml config file (default: "kdramadl.yml")
   --help, -h                    show help
   --version, -v                 print the version

COPYRIGHT:
   2017 https://github.com/lastmodified/kdramadl
```

### Advance Examples

#### Using the Command Prompt / Terminal

- For Windows users, you may need to change ``kdramadl`` to ``kdramadl.exe``.
- For MacOS/Linux users, you may need to change ``kdramadl`` to ``./kdramadl``

```bash

# Download resolution 720p, mp4 format and filename "example_video"
kdramadl --code "yourcode..." --resolution "720p" --format "mp4" --filename "example_video" --folder "C:\Downloads"

# Download subtitles only
kdramadl -c "yourcode..." --filename "example_video" --sub

# Download using the alternative host
kdramadl -c "yourcode..." --resolution "1" --format "mp4" --filename "example_video" --alt

# Download via proxy (that you must provide)
kdramadl -c "yourcode..." --resolution "1" --format "mkv" --filename "example_video" --proxy "http://192.168.0.1:80"

```

#### Using a Config file

You can create a configuration file ``kdramadl.yml`` and populate it with your desired default options. These options will then be used when you execute the app.

Example ``kdramadl.yml``:

```
format: mp4
folder: C:\Downloads
ffmpeg: C:\ffmpeg\ffmpeg.exe
autoquit: true
```