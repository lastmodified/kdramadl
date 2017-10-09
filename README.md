# kdramadl

 Alternative downloader for [https://kdrama.anontpp.com](https://kdrama.anontpp.com) (via [/r/koreanvariety](https://www.reddit.com/r/koreanvariety/comments/723mtd/i_created_this_website_that_streams_korean_shows/?sort=new) or [/r/KDRAMA/](https://www.reddit.com/r/KDRAMA/comments/723n1y/i_created_this_website_that_streams_korean_shows/)).

## Install

Download and extract the [latest release](https://github.com/lastmodified/kdramadl/releases/latest) for your OS.

If you do not have ``ffmpeg`` already installed, choose the ``*_ffmpeg.zip`` version (e.g.  ``kdramadl_windows_32bit_ffmpeg.zip``).

## Usage

You can launch the downloader by double-clicking on ``kdramadl.exe`` / ``kdramadl`` in Windows Explorer / Finder.

Alternatively, you may launch it from the Command Prompt / Terminal.

```
NAME:
   kdramadl - Make sure you have ffmpeg installed in PATH or in the current folder.

USAGE:
   kdramadl [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   Alternative downloader for https://kdrama.anontpp.com

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dc value        Download Code
   --res value       Resolution of video. Choose from: "1080p" "720p" "480p" "360p". Default is "1080p".
   --format value    Video format. Choose from: "mkv" "mp4". Default is "mkv".
   --filename value  Filename to save as (without extension).
   --sub             Download only subtitles.
   --ffmpeg value    Path to ffmpeg executable. (default: "ffmpeg")
   --folder value    Path to download folder.
   --alt             Use kdrama.armsasuncion.com instead of kdrama.anontpp.com
   --timeout value   Connection timeout interval in seconds. Default 10. (default: 10)
   --help, -h        show help
   --version, -v     print the version
```