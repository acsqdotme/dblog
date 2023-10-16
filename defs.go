package dblog

// DB custom type is gonna be my attempt at making paths to the database
// defined in the original package. Still a todo for now.
type DB struct {
	PathToDB string
}

// Post type has pretty much all of the content and metadata that I'd need
// to manipulate for a complex blogging system that I can still manage.
type Post struct {
	Title       string `yaml:"Title"`
	FileName    string `yaml:"FileName"`
	Description string `yaml:"Description"`
	PubDate     string `yaml:"PubDate"`
	UpdateDate  string `yaml:"UpdateDate"`
	Tags        []Tag  `yaml:"Tags"`
	Thumbnail   Img    `yaml:"Thumbnail"`
}

// Img struct exists for my thumbnail image in the front of my site, but
// I will definitely use this package for other things that I'll need for
// other friend's sites I'm designing
type Img struct {
	Src   string `json:"src" yaml:"Src"`
	Alt   string `json:"alt" yaml:"Alt"`
	Title string `json:"title" yaml:"Title"`
}

// Tag struct is there to aggregate together posts with common themes.
// I might add an html related attribute like content in post to let
// tag pages be more than just a paragraph.
type Tag struct {
	Name        string `yaml:"Name"`
	Category    string `yaml:"Category"`
	Description string `yaml:"Description"`
}
