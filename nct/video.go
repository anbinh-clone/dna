package nct

import (
	"dna"
	"dna/http"
	"dna/item"
	"dna/sqlpg"
	"errors"
	"time"
)

type Video struct {
	Id          dna.Int
	Key         dna.String // from API
	Title       dna.String // from API
	Artists     dna.StringArray
	Artistid    dna.Int // from API
	Topics      dna.StringArray
	Plays       dna.Int    // from API
	Likes       dna.Int    // from API
	Duration    dna.Int    // from API
	Thumbnail   dna.String // from API
	Image       dna.String // from API
	Type        dna.String // from API
	LinkKey     dna.String // from API
	LinkShare   dna.String // from API
	Lyric       dna.String
	StreamUrl   dna.String // from API
	Relatives   dna.StringArray
	DateCreated time.Time
	Checktime   time.Time
}

func NewVideo() *Video {
	video := new(Video)
	video.Id = 0
	video.Key = ""
	video.Title = ""
	video.Artists = dna.StringArray{}
	video.Artistid = 0
	video.Topics = dna.StringArray{}
	video.Plays = 0
	video.Likes = 0
	video.Duration = 0
	video.Thumbnail = ""
	video.Image = ""
	video.Type = ""
	video.LinkKey = ""
	video.LinkShare = ""
	video.Lyric = ""
	video.StreamUrl = ""
	video.Relatives = dna.StringArray{}
	video.DateCreated = time.Time{}
	video.Checktime = time.Now()
	return video
}

// getVideoPlays returns video plays
func getVideoPlays(video *Video, body dna.String) {
	link := "http://www.nhaccuatui.com/interaction/api/hit-counter?jsoncallback=nct"
	http.DefaulHeader.Set("Content-Type", "application/x-www-form-urlencoded ")
	result, err := http.Post(dna.String(link), body)
	// Log(link)
	if err == nil {
		data := &result.Data
		tpl := dna.String(`{"counter":([0-9]+)}`)
		playsArr := data.FindAllStringSubmatch(tpl, -1)
		if len(playsArr) > 0 {
			video.Plays = playsArr[0][1].ToInt()
		}
	}
}

// getVideoFromMainPage returns video from main page
func getVideoFromMainPage(video *Video) <-chan bool {
	channel := make(chan bool, 1)
	go func() {
		link := "http://www.nhaccuatui.com/video/google-bot." + video.Key + ".html"
		result, err := http.Get(link)
		if err == nil {
			data := &result.Data

			topicsArr := data.FindAllStringSubmatch(`<strong>Thể loại</strong></p>[\n\t\r]+(.+)`, 1)
			if len(topicsArr) > 0 {
				video.Topics = topicsArr[0][1].RemoveHtmlTags("").Trim().Split(", ")
			}

			linkkeyArr := data.FindAllStringSubmatch(`"flashPlayer", ".+", "(.+?)"`, 1)
			if len(linkkeyArr) > 0 {
				video.LinkKey = linkkeyArr[0][1].Trim()
			}

			relativesArr := data.FindAllString("<a href=.+[\n\t\r]+.+video_relative", -1)
			video.Relatives = dna.StringArray(relativesArr.Map(func(val dna.String, idx dna.Int) dna.String {
				valArr := val.GetTagAttributes("href").Replace(".html", "").ReplaceWithRegexp(`^.+\.`, "")
				return valArr
			}).([]dna.String))

			// new version in June 2013 does not support lyric of videos

			// idArr := data.FindAllStringSubmatch(`value="(.+)" id="inpHiddenId"`, 1)
			// if len(idArr) > 0 {
			// 	video.Id = idArr[0][1].ToInt()
			// }

			// titleArr := data.FindAllStringSubmatch(`<h1 itemprop="name">(.+?)</h1>`, 1)
			// if len(titleArr) > 0 {
			// 	video.Title = titleArr[0][1].Trim().SplitWithRegexp(" - ", 2)[0].Trim()
			// }

			// artistsArr := data.FindAllStringSubmatch(`<h1 itemprop="name">(.+?)</h1>`, 1)
			// if len(artistsArr) > 0 {
			// 	artists := artistsArr[0][1].RemoveHtmlTags("").SplitWithRegexp(" - ", 2)
			// 	if artists.Length() == 2 {
			// 		video.Artists = artists[1].Split(", ").Filter(func(v dna.String, i dna.Int) dna.Bool {
			// 			if v != "" {
			// 				return true
			// 			} else {
			// 				return false
			// 			}
			// 		})
			// 	}
			// }

			// durationArr := data.FindAllStringSubmatch(`<meta itemprop="duration" content="(.+)" />`, 1)
			// if len(durationArr) > 0 {
			// 	durationF := durationArr[0][1].Replace("PT", "").Replace("H", "h").Replace("M", "m").Replace("S", "s")
			// 	duration, perr := time.ParseDuration(durationF.String())
			// 	if perr == nil {
			// 		video.Duration = dna.Float(duration.Seconds()).Round()
			// 	}
			// }

			// thumbArr := data.FindAllString(`<link rel="image_src".+`, 1)
			// if thumbArr.Length() > 0 {
			// 	video.Thumbnail = thumbArr[0].GetTagAttributes("href").Trim()
			// 	datecreatedArr := video.Thumbnail.FindAllStringSubmatch(`/([0-9]+)_[0-9]+\..+$`, -1)
			// 	if len(datecreatedArr) > 0 {
			// 		// Log(int64(datecreatedArr[0][1].ToInt()))
			// 		video.DateCreated = time.Unix(int64(datecreatedArr[0][1].ToInt()/1000), 0)
			// 	} else {
			// 		dateCreatedArr := video.Thumbnail.FindAllStringSubmatch(`/?(\d{4}/\d{2}/\d{2})`, -1)
			// 		if len(dateCreatedArr) > 0 {
			// 			year := dateCreatedArr[0][1].FindAllStringSubmatch(`(\d{4})/\d{2}/\d{2}`, -1)[0][1].ToInt()
			// 			month := dateCreatedArr[0][1].FindAllStringSubmatch(`\d{4}/(\d{2})/\d{2}`, -1)[0][1].ToInt()
			// 			day := dateCreatedArr[0][1].FindAllStringSubmatch(`\d{4}/\d{2}/(\d{2})`, -1)[0][1].ToInt()
			// 			video.DateCreated = time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
			// 		}
			// 	}
			// }

			// Find params for the number of video plays
			// itemIdArr := data.FindAllStringSubmatch(`NCTWidget.hitCounter\('(.+?)'.+`, 1)
			// timeArr := data.FindAllStringSubmatch(`NCTWidget.hitCounter\('.+?', '(.+?)'.+\);`, 1)
			// signArr := data.FindAllStringSubmatch(`NCTWidget.hitCounter\('.+?', '.+?', '(.+?)'.+;`, 1)
			// typeArr := data.FindAllStringSubmatch(`NCTWidget.hitCounter\('.+?', '.+?', '.+?', "(.+?)"\);`, 1)
			// if len(itemIdArr) > 0 && len(timeArr) > 0 && len(signArr) > 0 && len(typeArr) > 0 {
			// 	// boday has post form:
			// 	// item_id=2870710&time=1389009424631&sign=2499ab08f6662842a02b06aad603d8ab&type=video
			// 	body := dna.Sprintf(`item_id=%v&time=%v&sign=%v&type=%v`, itemIdArr[0][1], timeArr[0][1], signArr[0][1], typeArr[0][1])
			// 	getVideoPlays(video, body)
			// 	video.Type = typeArr[0][1].Trim()
			// 	if video.Type == "video" {
			// 		video.Type = "mv"
			// 	}
			// 	video.Id = itemIdArr[0][1].ToInt()
			// }

			// GetRelevantPortions(&result.Data)
		}
		channel <- true
	}()
	return channel
}
func getAPIVideoC(video *Video) <-chan bool {
	channel := make(chan bool, 1)
	go func() {
		apivideo, err := GetAPIVideo(video.Id)
		if err == nil {
			apivideo.FillVideo(video)
		} else {
			video.Id = 0 // So error will be returned
		}
		channel <- true
	}()
	return channel
}

// GetVideo returns a video or an error
// 	* key: A unique key of a video
// 	* Official : 0 or 1, if its value is unknown, set to 0
// 	* Returns a found video or an error
func GetVideo(id dna.Int) (*Video, error) {
	var video *Video = NewVideo()

	video.Id = id

	c := make(chan bool, 1)

	go func() {
		c <- <-getAPIVideoC(video)
	}()
	for i := 0; i < 1; i++ {
		<-c
	}
	<-getVideoFromMainPage(video)

	if video.Id == 0 {
		return nil, errors.New(dna.Sprintf("NCT - Video id:%v not found", id).String())
	} else {
		return video, nil
	}
}

// Fetch implements item.Item interface.
// Returns error if can not get item
func (video *Video) Fetch() error {
	_video, err := GetVideo(video.Id)
	if err != nil {
		return err
	} else {
		*video = *_video
		return nil
	}
}

// GetId implements GetId methods of item.Item interface
func (video *Video) GetId() dna.Int {
	return video.Id
}

// New implements item.Item interface
// Returns new item.Item interface
func (video *Video) New() item.Item {
	return item.Item(NewVideo())
}

// Init implements item.Item interface.
// It sets Id or key.
// dna.Interface v has type int or dna.Int, it calls Id field.
// Otherwise if v has type string or dna.String, it calls Key field.
func (video *Video) Init(v interface{}) {
	switch v.(type) {
	case int:
		video.Id = dna.Int(v.(int))
	case dna.Int:
		video.Id = v.(dna.Int)
	default:
		panic("Interface v has to be int")
	}
}

func (video *Video) Save(db *sqlpg.DB) error {
	return db.InsertIgnore(video)
}
