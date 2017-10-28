package main

import "encoding/json"
import "encoding/xml"
import "fmt"
import "io/ioutil"
import "log"

import "net/http"
import "os"
import "path/filepath"

import "strconv"
import "strings"

import "github.com/kardianos/osext"

func main() {
	cfgpath, _ := osext.ExecutableFolder()
	file, _ := os.Open(cfgpath + "/vlg.json")
	decoder := json.NewDecoder(file)
	cfg = configuration{}
	err := decoder.Decode(&cfg)
	if err != nil {
		fmt.Println("error:", err)
	}

	if cfg.HostName == "" {
		cfg.HostName = "http://localhost"
	}
	if cfg.HostPort == 0 {
		cfg.HostPort = 32400
	}

	if cfg.PlexToken == "" {
		writeToLog("ERROR: Invalid PlexToken")
		return
	}

	urlRoot := cfg.HostName + ":" + strconv.Itoa(cfg.HostPort)
	plexToken := cfg.PlexToken

	plexSections := getPlexSections(urlRoot, plexToken)
	for _, section := range plexSections {
		plexCollection := getPlexCollection(urlRoot, section.ID, plexToken)
		for _, c := range plexCollection {
			mediaPaths := getPlexCollectionContents(urlRoot, section.ID, c.Key, plexToken)
			for _, p := range mediaPaths {
				vlTitle := strings.Replace(c.Title, "VL-", "", 1)

				// Create Symlink on non-pooled libraries
				if cfg.getVirtualLibPoolRoot(section.Title) == "" {
					createFolderandLinks(vlTitle, section.VirtualLibPath, p)
				} else {
					root := cfg.getVirtualLibPoolRoot(section.Title)
					pool := cfg.getVirtualLibPool(section.Title)
					for _, vlPool := range pool {
						vlPath := strings.Replace(section.VirtualLibPath, root, vlPool, 1)
						fPath := strings.Replace(p, root, vlPool, 1)
						fmt.Println(vlPath)
						fmt.Println(fPath)
						createFolderandLinks(vlTitle, vlPath, fPath)
					}
				}
			}
		}
	}
}

func getPlexSections(urlRoot, plexToken string) (sections []section) {
	urlPath := fmt.Sprintf("%s/library/sections/?X-Plex-Token=%s", urlRoot, plexToken)

	res, err := http.Get(urlPath)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	var data sectionMediaContainer
	err = xml.Unmarshal([]byte(body), &data)
	if err != nil {
		panic(err.Error())
	}

	for _, dir := range data.Directory {
		for _, s := range cfg.Sections {
			if strings.ToUpper(s.Name) == strings.ToUpper(dir.Title) {
				sections = append(sections, section{Title: dir.Title, Path: dir.Location.Path, ID: dir.Location.ID, VirtualLibPath: s.VirtualLibPath})
			}
		}
	}

	return
}

func getPlexCollection(urlRoot, section, plexToken string) (collections []collection) {
	urlPath := fmt.Sprintf("%s/library/sections/%s/collection?X-Plex-Token=%s", urlRoot, section, plexToken)

	res, err := http.Get(urlPath)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	var data collectionMediaContainer
	err = xml.Unmarshal([]byte(body), &data)
	if err != nil {
		panic(err.Error())
	}

	for _, dir := range data.Directory {
		if strings.Contains(dir.Title, "VL-") {
			collections = append(collections, collection{Title: dir.Title, Key: dir.Key})
		}
	}
	return
}

func getPlexCollectionContents(urlRoot, section, collection, plexToken string) (paths []string) {
	urlPath := fmt.Sprintf("%s/library/sections/%s/all?collection=%s&X-Plex-Token=%s", urlRoot, section, collection, plexToken)

	res, err := http.Get(urlPath)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	var data mediaMediaContainer
	err = xml.Unmarshal([]byte(body), &data)
	if err != nil {
		panic(err.Error())
	}

	for _, media := range data.Video {
		//fmt.Printf("%#v\n", media.Media.Part.File)
		paths = append(paths, media.Media.Part.File)
	}

	return
}

func createFolderandLinks(vlTitle, vlPath, filePath string) {
	// Create VirtualLib root if it doesn't exist
	vlPath = filepath.Join(vlPath, vlTitle)
	err := os.Mkdir(vlPath, os.ModeDir)
	if err != nil {
		fmt.Println(err)
	}

	// Create Base folder within VirtualLib root to place the Symlink
	fp := filepath.Dir(filePath)
	fpb := filepath.Base(fp)
	vlPath = filepath.Join(vlPath, fpb)

	err = os.Mkdir(vlPath, os.ModeDir)
	if err != nil {
		fmt.Println(err)
	}

	// Create Symlink
	fileName := filepath.Base(filePath)
	vlPath = filepath.Join(vlPath, fileName)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		fmt.Printf("Making Link: %s => %s\n", filePath, vlPath)
		err = os.Link(filePath, vlPath)
		if err != nil {
			fmt.Printf("LinkError: %#v\n", err)
		}
	}
}

func writeToLog(str string) {
	filename := cfg.LogLocation
	if cfg.LogLocation == "" {
		filename, _ = osext.ExecutableFolder()
	}

	f, err := os.OpenFile(filename+"/vlg.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(str)
}

// Configuration
type configuration struct {
	LogLocation string `json:"loglocation"`
	HostName    string `json:"hostname"`
	HostPort    int    `json:"port"`
	PlexToken   string `json:"plextoken"`
	Sections    []struct {
		ID                 string   `json:"_id"`
		Name               string   `json:"name"`
		VirtualLibPath     string   `json:"virtuallibpath"`
		VirtualLibPoolRoot string   `json:"virtuallibpoolroot"`
		VirtualLibPool     []string `json:"virtuallibpool"`
	} `json:"sections"`
}

func (c configuration) getVirtualLibPoolRoot(vlLibrary string) string {
	for _, c := range c.Sections {
		if c.Name == vlLibrary {
			fmt.Printf("Root: %s\n", c.VirtualLibPoolRoot)
			return c.VirtualLibPoolRoot
		}
	}
	return ""
}

func (c configuration) getVirtualLibPool(vlLibrary string) (pool []string) {
	for _, c := range c.Sections {
		if c.Name == vlLibrary {
			return c.VirtualLibPool
		}
	}
	return
}

var cfg configuration

// Sections
type sectionMediaContainer struct {
	Directory       []sectionDirectory `xml:"Directory"`
	Size            int                `xml:"size,attr"`
	AllowSync       int                `xml:"allowSync,attr"`
	Identifier      string             `xml:"identifier,attr"`
	MediaTagPrefix  string             `xml:"mediaTagPrefix,attr"`
	MediaTagVersion string             `xml:"mediaTagVersion,attr"`
	Title1          string             `xml:"title1,attr"`
}

type sectionDirectory struct {
	Location            sectionLocation `xml:"Location"`
	Refreshing          int             `xml:"refreshing,attr"`
	UpdatedAt           int64           `xml:"updatedAt,attr"`
	Filters             int             `xml:"filters,attr"`
	Thumb               string          `xml:"thumb,attr"`
	Language            string          `xml:"language,attr"`
	Key                 int             `xml:"key,attr"`
	Scanner             string          `xml:"scanner,attr"`
	CreatedAt           int64           `xml:"createdAt,attr"`
	Composite           string          `xml:"composite,attr"`
	Art                 string          `xml:"art,attr"`
	AllowSync           int             `xml:"allowSync,attr"`
	Type                string          `xml:"type,attr"`
	EnableAutoPhotoTags string          `xml:"enableAutoPhotoTags,attr"`
	Title               string          `xml:"title,attr"`
	Agent               string          `xml:"agent,attr"`
	UUID                string          `xml:"uuid,attr"`
}
type sectionLocation struct {
	Path string `xml:"path,attr"`
	ID   string `xml:"id,attr"`
}

type section struct {
	Title          string
	Path           string
	ID             string
	VirtualLibPath string
}

// Collections
type collectionMediaContainer struct {
	ViewMode        string                `xml:"viewMode,attr"`
	Content         string                `xml:"content,attr"`
	MediaTagVersion string                `xml:"mediaTagVersion,attr"`
	Thumb           string                `xml:"thumb,attr"`
	Size            string                `xml:"size,attr"`
	AllowSync       string                `xml:"allowSync,attr"`
	Identifier      string                `xml:"identifier,attr"`
	Directory       []collectionDirectory `xml:"Directory"`
	Art             string                `xml:"art,attr"`
	MediaTagPrefix  string                `xml:"mediaTagPrefix,attr"`
	Title2          string                `xml:"title2,attr"`
	ViewGroup       string                `xml:"viewGroup,attr"`
	Title1          string                `xml:"title1,attr"`
}

type collectionDirectory struct {
	FastKey string `xml:"fastKey,attr"`
	Key     string `xml:"key,attr"`
	Title   string `xml:"title,attr"`
}

type collection struct {
	Key   string
	Title string
}

// Collection Contents
type mediaMediaContainer struct {
	LibrarySectionID    string       `xml:"librarySectionID,attr"`
	ViewMode            string       `xml:"viewMode,attr"`
	Thumb               string       `xml:"thumb,attr"`
	Video               []mediaVideo `xml:"Video"`
	Size                string       `xml:"size,attr"`
	LibrarySectionTitle string       `xml:"librarySectionTitle,attr"`
	Title2              string       `xml:"title2,attr"`
	ViewGroup           string       `xml:"viewGroup,attr"`
	MediaTagPrefix      string       `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string       `xml:"mediaTagVersion,attr"`
	LibrarySectionUUID  string       `xml:"librarySectionUUID,attr"`
	Identifier          string       `xml:"identifier,attr"`
	AllowSync           string       `xml:"allowSync,attr"`
	Title1              string       `xml:"title1,attr"`
	Art                 string       `xml:"art,attr"`
}

type mediaVideo struct {
	Studio                string     `xml:"studio,attr"`
	RatingImage           string     `xml:"ratingImage,attr"`
	Rating                string     `xml:"rating,attr"`
	Art                   string     `xml:"art,attr"`
	RatingKey             string     `xml:"ratingKey,attr"`
	AddedAt               string     `xml:"addedAt,attr"`
	Thumb                 string     `xml:"thumb,attr"`
	UpdatedAt             string     `xml:"updatedAt,attr"`
	Duration              string     `xml:"duration,attr"`
	Type                  string     `xml:"type,attr"`
	OriginallyAvailableAt string     `xml:"originallyAvailableAt,attr"`
	Title                 string     `xml:"title,attr"`
	Year                  string     `xml:"year,attr"`
	Key                   string     `xml:"key,attr"`
	Summary               string     `xml:"summary,attr"`
	Media                 mediaMedia `xml:"Media"`
}
type mediaMedia struct {
	Duration              string    `xml:"duration,attr"`
	Bitrate               string    `xml:"bitrate,attr"`
	AudioChannels         string    `xml:"audioChannels,attr"`
	VideoResolution       string    `xml:"videoResolution,attr"`
	VideoFrameRate        string    `xml:"videoFrameRate,attr"`
	ID                    string    `xml:"id,attr"`
	Height                string    `xml:"height,attr"`
	Container             string    `xml:"container,attr"`
	OptimizedForStreaming string    `xml:"optimizedForStreaming,attr"`
	AudioProfile          string    `xml:"audioProfile,attr"`
	Width                 string    `xml:"width,attr"`
	VideoCodec            string    `xml:"videoCodec,attr"`
	AspectRatio           string    `xml:"aspectRatio,attr"`
	AudioCodec            string    `xml:"audioCodec,attr"`
	Has64bitOffsets       string    `xml:"has64bitOffsets,attr"`
	VideoProfile          string    `xml:"videoProfile,attr"`
	Part                  mediaPart `xml:"Part"`
}
type mediaPart struct {
	Size                  string `xml:"size,attr"`
	AudioProfile          string `xml:"audioProfile,attr"`
	VideoProfile          string `xml:"videoProfile,attr"`
	Key                   string `xml:"key,attr"`
	Duration              string `xml:"duration,attr"`
	File                  string `xml:"file,attr"`
	OptimizedForStreaming string `xml:"optimizedForStreaming,attr"`
	Id                    string `xml:"id,attr"`
	Has64bitOffsets       string `xml:"has64bitOffsets,attr"`
	Container             string `xml:"container,attr"`
	HasThumbnail          string `xml:"hasThumbnail,attr"`
}
