package zi

import (
	"dna"
	"dna/item"
	"dna/sqlpg"
	"errors"
	"time"
)

type Artist struct {
	Id          dna.Int
	Name        dna.String
	Alias       dna.String
	Birthname   dna.String
	Birthday    dna.String
	Sex         dna.Int
	Link        dna.String
	Topics      dna.StringArray
	Avatar      dna.String
	Coverart    dna.String
	Coverart2   dna.String
	ZmeAcc      dna.String
	Role        dna.String
	Website     dna.String
	Biography   dna.String
	Publisher   dna.String
	Country     dna.String
	IsOfficial  dna.Int
	YearActive  dna.String
	StatusId    dna.Int
	DateCreated time.Time
	Checktime   time.Time
}

func NewArtist() *Artist {
	artist := new(Artist)
	artist.Id = 0
	artist.Name = ""
	artist.Alias = ""
	artist.Birthname = ""
	artist.Birthday = ""
	artist.Sex = 0
	artist.Link = ""
	artist.Topics = dna.StringArray{}
	artist.Avatar = ""
	artist.Coverart = ""
	artist.Coverart2 = ""
	artist.ZmeAcc = ""
	artist.Role = ""
	artist.Website = ""
	artist.Biography = ""
	artist.Publisher = ""
	artist.Country = ""
	artist.IsOfficial = 0
	artist.YearActive = ""
	artist.StatusId = 1
	artist.DateCreated = time.Time{}
	artist.Checktime = time.Time{}
	return artist
}

//GetArtistFromAPI gets a artist from API.
func GetArtistFromAPI(id dna.Int) (*Artist, error) {
	var artist *Artist = NewArtist()
	artist.Id = id

	apiArtist, err := GetAPIArtist(id)
	if err != nil {
		return nil, err
	} else {
		if apiArtist.Response.MsgCode == 1 {
			artist.Name = apiArtist.Name
			artist.Alias = apiArtist.Alias
			artist.Birthname = apiArtist.Birthname
			artist.Birthday = apiArtist.Birthday
			artist.Sex = apiArtist.Sex
			artist.Link = apiArtist.Link
			artist.Topics = apiArtist.Topics.Split(", ").SplitWithRegexp(" / ").Unique()
			artist.Avatar = apiArtist.Avatar
			artist.Coverart = apiArtist.Coverart
			artist.Coverart2 = apiArtist.Coverart2
			artist.ZmeAcc = apiArtist.ZmeAcc
			artist.Role = apiArtist.Role
			artist.Website = apiArtist.Website
			artist.Biography = apiArtist.Biography
			artist.Publisher = apiArtist.Publisher
			artist.Country = apiArtist.Country
			artist.IsOfficial = apiArtist.IsOfficial
			artist.YearActive = apiArtist.YearActive
			artist.StatusId = apiArtist.StatusId
			if apiArtist.DateCreated > 0 {
				artist.DateCreated = time.Unix(int64(apiArtist.DateCreated), 0)
			}
			artist.Checktime = time.Now()
			return artist, nil
		} else {
			return nil, errors.New("Message code invalid " + apiArtist.Response.MsgCode.ToString().String())
		}

	}

}

// GetArtist returns a artist or an error
func GetArtist(id dna.Int) (*Artist, error) {
	artist, err := GetArtistFromAPI(id)
	if err == nil {
		if artist.Name == "" {
			return nil, errors.New(dna.Sprintf("Zing - Artist %v: Artist not found", artist.Id).String())
		} else {
			return artist, nil
		}
	} else {
		return nil, err
	}
}

// Fetch implements item.Item interface.
// Returns error if can not get item
func (artist *Artist) Fetch() error {
	_artist, err := GetArtist(artist.Id)
	if err != nil {
		return err
	} else {
		*artist = *_artist
		return nil
	}
}

// GetId implements GetId methods of item.Item interface
func (artist *Artist) GetId() dna.Int {
	return artist.Id
}

// New implements item.Item interface
// Returns new item.Item interface
func (artist *Artist) New() item.Item {
	return item.Item(NewArtist())
}

// Init implements item.Item interface.
// It sets Id or key.
// dna.Interface v has type int or dna.Int, it calls Id field.
// Otherwise if v has type string or dna.String, it calls Key field.
func (artist *Artist) Init(v interface{}) {
	switch v.(type) {
	case int:
		artist.Id = dna.Int(v.(int))
	case dna.Int:
		artist.Id = v.(dna.Int)
	default:
		panic("Interface v has to be int")
	}
}

func (artist *Artist) Save(db *sqlpg.DB) error {
	return db.InsertIgnore(artist)
}
