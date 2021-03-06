package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type config struct {
	Token     string              `json:"token"`
	FolderMap map[string][]string `json:"folderMapping"`
}

const maxImages = 5

var (
	imageRegex = regexp.MustCompile(`(?i)(.jpg|.png|.gif|.jpeg)$`)
	letters    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	var configFile string
	if len(os.Args) == 2 {
		configFile = os.Args[1]
	} else {
		log.Fatalln("Please provide a configuration file as a second parameter")
	}

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("There was an error opening the file %s", configFile)
		log.Fatalln(err)
	}

	var config config
	json.Unmarshal(file, &config)

	bot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Println("Error creating discord client")
		log.Fatalln(err)
	}

	bot.Open()
	defer bot.Close()

	var wg sync.WaitGroup
	wg.Add(len(config.FolderMap))

	for d, channels := range config.FolderMap {
		go func(d string, channels []string) {
			defer wg.Done()

			images := setupDirectory(d)
			for i, image := range images {
				if (i + 1) <= maxImages {
					for _, c := range channels {

						f, err := os.Open(path.Join(d, image))
						if err != nil {
							log.Fatalln(err)
						}
						defer f.Close()

						_, err = bot.ChannelFileSend(c, image, f)
						if err != nil {
							log.Printf("Error posting image %s to channel %s\n", image, c)
							log.Println(err)
						}
					}

					originalName := image
					if _, err := os.Stat(path.Join(d, "uploaded", image)); err == nil {
						split := strings.Split(image, ".")
						image = randSeq(16) + "." + split[len(split)-1]
					}

					err := os.Rename(path.Join(d, originalName), path.Join(d, "uploaded", image))
					if err != nil {
						println(err.Error())
					}
				}
			}
		}(d, channels)
	}

	wg.Wait()
}

func setupDirectory(directory string) []string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Printf("Error walking directory %s\n", directory)
		log.Fatalln(err)
	}

	_ = os.Mkdir(path.Join(directory, "uploaded"), 0755)
	var images []string
	for _, file := range files {
		if !file.IsDir() && imageRegex.MatchString(file.Name()) {
			images = append(images, file.Name())
		}
	}
	return images
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
