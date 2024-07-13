package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkdai/youtube/v2"
	"github.com/schollz/progressbar/v3"
)

type choice int

const (
	chooseOneResolution choice = iota
	chooseSeparately
)

type item struct {
	title       string
	description string
	format      youtube.Format
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type model struct {
	list         list.Model
	chosenFormat *youtube.Format
	choice       choice
	quit         bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.quit = true
			return m, tea.Quit
		case "enter":
			if m.choice == -1 {
				if i, ok := m.list.SelectedItem().(item); ok {
					if i.title == "Choose one resolution for all videos" {
						m.choice = chooseOneResolution
					} else if i.title == "Choose resolution for each video separately" {
						m.choice = chooseSeparately
					}
					return m, tea.Quit
				}
			} else {
				if i, ok := m.list.SelectedItem().(item); ok {
					m.chosenFormat = &i.format
					return m, tea.Quit
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quit {
		return ""
	}
	return m.list.View()
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: youtube-downloader <youtube_url_or_playlist_url> [destination_path]")
		waitForEnter()
		return
	}

	url := os.Args[1]
	var destPath string

	if len(os.Args) == 3 {
		destPath = os.Args[2]
		err := os.MkdirAll(destPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create destination directory: %v", err)
		}
	} else {
		destPath, _ = os.Getwd()
	}

	client := youtube.Client{}
	if strings.Contains(url, "playlist?list=") {
		choice := getUserChoice()
		if choice == chooseOneResolution {
			fmt.Println("Fetching playlist details...")
			downloadPlaylistWithSingleResolution(client, url, destPath)
		} else {
			downloadPlaylist(client, url, destPath)
		}
	} else {
		downloadVideo(client, url, destPath)
	}
}

func getUserChoice() choice {
	items := []list.Item{
		item{title: "Choose one resolution for all videos"},
		item{title: "Choose resolution for each video separately"},
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 10)
	l.Title = "How would you like to download the playlist?"

	p := tea.NewProgram(model{list: l, choice: -1})

	chosenModel, err := p.StartReturningModel()
	if err != nil {
		log.Fatalf("Failed to start tea program: %v", err)
	}

	chosenItem := chosenModel.(model).list.SelectedItem()
	switch chosenItem.(item).title {
	case "Choose one resolution for all videos":
		return chooseOneResolution
	case "Choose resolution for each video separately":
		return chooseSeparately
	default:
		return chooseSeparately
	}
}

func downloadPlaylist(client youtube.Client, playlistURL, destPath string) {
	playlist, err := client.GetPlaylist(playlistURL)
	if err != nil {
		log.Fatalf("Failed to get playlist: %v", err)
	}

	for _, video := range playlist.Videos {
		fmt.Printf("Downloading video: %s\n", video.Title)
		downloadVideo(client, video.ID, destPath)
	}
}

func downloadPlaylistWithSingleResolution(client youtube.Client, playlistURL, destPath string) {
	playlist, err := client.GetPlaylist(playlistURL)
	if err != nil {
		log.Fatalf("Failed to get playlist: %v", err)
	}

	if len(playlist.Videos) == 0 {
		fmt.Println("No videos found in the playlist")
		return
	}

	firstVideo := playlist.Videos[0]
	video, err := client.GetVideo(firstVideo.ID)
	if err != nil {
		log.Fatalf("Failed to get video: %v", err)
	}

	items := []list.Item{}
	for _, format := range video.Formats.Type("video/mp4") {
		description := "Resolution: " + format.QualityLabel
		items = append(items, item{
			title:       format.QualityLabel,
			description: description,
			format:      format,
		})
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 30)
	l.Title = "Choose a resolution for all videos"

	p := tea.NewProgram(model{list: l, choice: chooseOneResolution})

	chosenFormat, err := p.StartReturningModel()
	if err != nil {
		log.Fatalf("Failed to start tea program: %v", err)
	}

	format := chosenFormat.(model).chosenFormat

	for _, video := range playlist.Videos {
		fmt.Printf("Downloading video: %s\n", video.Title)
		downloadVideoWithFormat(client, video.ID, destPath, format)
	}
}

func downloadVideo(client youtube.Client, videoURL, destPath string) {
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Fatalf("Failed to get video: %v", err)
	}

	items := []list.Item{}
	for _, format := range video.Formats.Type("video/mp4") {
		description := "Resolution: " + format.QualityLabel
		items = append(items, item{
			title:       format.QualityLabel,
			description: description,
			format:      format,
		})
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 30)
	l.Title = "Choose a resolution"

	p := tea.NewProgram(model{list: l, choice: chooseSeparately})

	chosenFormat, err := p.StartReturningModel()
	if err != nil {
		log.Fatalf("Failed to start tea program: %v", err)
	}

	format := chosenFormat.(model).chosenFormat

	downloadVideoWithFormat(client, videoURL, destPath, format)
}

func downloadVideoWithFormat(client youtube.Client, videoURL, destPath string, format *youtube.Format) {
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Fatalf("Failed to get video: %v", err)
	}

	videoStream, _, err := client.GetStream(video, format)
	if err != nil {
		log.Fatalf("Failed to get video stream: %v", err)
	}

	audioFormat := video.Formats.Type("audio/mp4")[0]
	audioStream, _, err := client.GetStream(video, &audioFormat)
	if err != nil {
		log.Fatalf("Failed to get audio stream: %v", err)
	}

	videoTitle := sanitizeFileName(video.Title)
	videoPath := filepath.Join(destPath, videoTitle+"_video.mp4")
	audioPath := filepath.Join(destPath, videoTitle+"_audio.m4a")

	err = downloadStream(videoPath, videoStream, format.ContentLength)
	if err != nil {
		log.Fatalf("Failed to download video: %v", err)
	}

	err = downloadStream(audioPath, audioStream, audioFormat.ContentLength)
	if err != nil {
		log.Fatalf("Failed to download audio: %v", err)
	}

	outputPath := filepath.Join(destPath, videoTitle+".mp4")
	mergeStreams(videoPath, audioPath, outputPath)

	fmt.Println("Download complete:", outputPath)
}

func downloadStream(path string, stream io.Reader, length int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	bar := progressbar.DefaultBytes(
		length,
		fmt.Sprintf("downloading %s", filepath.Base(path)),
	)

	_, err = io.Copy(io.MultiWriter(file, bar), stream)
	return err
}

func mergeStreams(videoPath, audioPath, outputPath string) {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-c", "copy", outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to merge video and audio: %v", err)
	}

	os.Remove(videoPath)
	os.Remove(audioPath)
}

func sanitizeFileName(name string) string {
	var re = regexp.MustCompile(`[<>:"/\\|?*]`)
	return re.ReplaceAllString(name, "_")
}

func waitForEnter() {
	fmt.Println("Press 'Enter' to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
