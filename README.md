# youtube-downloader

A robust and fast command-line program to download YouTube videos or playlists by simply providing their URLs. The program is built using Go, and features an interactive menu to select video resolutions. It ensures videos are downloaded with audio and provides progress feedback using a progress bar.

## Features

- Download individual YouTube videos or entire playlists.
- Select video resolutions through an interactive menu.
- Choose a single resolution for all videos in a playlist or specify resolutions for each video individually.
- Download progress displayed using a progress bar.
- Merges video and audio streams using `ffmpeg`.

## Installation

### Prerequisites

- [ffmpeg](https://ffmpeg.org/download.html) (Ensure `ffmpeg` is installed and added to your system's PATH)

### Download the Executable
 Download the latest version of `youtube-downloader` from the [releases page](https://github.com/Spandan7724/youtube-downloader/releases).

### Clone the Repository

```bash
git clone https://github.com/yourusername/youtube-downloader-cli.git
cd youtube-downloader-cli 
```

### Build the Project

```bash
go build -o youtube-downloader.exe
```
 ## Usage

### Basic Usage

```bash
./youtube-downloader <youtube_url_or_playlist_url> [destination_path]
```
-   `<youtube_url_or_playlist_url>`: URL of the YouTube video or playlist you want to download.
-   `[destination_path]` (optional): The directory where the downloaded videos will be saved. Defaults to the current directory if not specified.

### Examples

#### Download a Single Video

```bash
./youtube-downloader https://www.youtube.com/watch?v=your_video_id
```
#### Download a Playlist

```bash
./youtube-downloader https://www.youtube.com/playlist?list=your_playlist_id
```
#### Specify a Destination Path

```bash
./youtube-downloader https://www.youtube.com/watch?v=your_video_id C:\path\to\destination
```
#### Download Process
The program downloads the video and audio streams separately.

#### Merging Streams
The program uses ffmpeg to merge the video and audio streams into a single file.

