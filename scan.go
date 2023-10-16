package dblog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	pathToPostsDir = "./testDir" // very temporary; I'm planning on moving over to db method based approach.
)

// checkPost is an internal func to ensure that all the attributes in a post
// that are NOT NULL in the db are, in fact, not null before loading them into
// the db.
func checkPost(p Post) error { // don't need to check thumbnail
	if p.Title == "" ||
		p.FileName == "" ||
		p.Description == "" ||
		p.PubDate == "" ||
		p.UpdateDate == "" ||
		len(p.Tags) < 1 {
		return errors.New("unfilled required attributes")
	}
	return nil
}

// AddPost checks to see if a post is valid first will checkPost then also by
// ensuring every tag that is inside the post exists in the db. You don't even
// need to add anything more than the names of the tags in the tag slice of the
// post as that's how it's checked before the post is inserted into the db as
// well as being how the tag_id is filled into the posts_tags associative
// identity for linking the metadata together.
func AddPost(post Post) (err error) {
	if err = checkPost(post); err != nil {
		return err
	}

	// ensure tag existence
	for _, t := range post.Tags {
		if !DoesTagExist(t.Name) {
			return errors.New("missing tag: " + t.Name)
		}
	}

	db := openDB()
	defer closeDB(db)

	var jsonThumbnail []byte
	if post.Thumbnail.Src != "" {
		jsonThumbnail, err = json.Marshal(post.Thumbnail)
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(`INSERT INTO post (title, file_name, description, pub_date, update_date, thumbnail)
  VALUES
  (?, ?, ?, ?, ?, ?)
  `, post.Title, post.FileName, post.Description, post.PubDate, post.UpdateDate, string(jsonThumbnail))
	if err != nil {
		return err
	}

	for _, t := range post.Tags {
		_, err = db.Exec(`INSERT INTO post_tag (post_id, tag_id)
    VALUES
    (
    (SELECT id FROM post WHERE file_name = ?),
    (SELECT id FROM tag WHERE name = ?)
    )
    `, post.FileName, t.Name)
	}

	return nil
}

// ReadMeta is gonna be for unmarshalling meta.yml into a Post struct
func ReadMeta(path string) (p Post, err error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}
	err = yaml.Unmarshal(fileContent, &p)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}

// ScanPosts is incomplete
func ScanPosts() error {
	// need to actually make this
	err := filepath.Walk(pathToPostsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() && !DoesPostExist(info.Name()) && info.Name() != "testDir" {
			fmt.Println("adding " + info.Name() + "to db")
			post, err := ReadMeta(filepath.Join(pathToPostsDir, info.Name(), "meta.yml"))
			if err != nil {
				return err
			}

			err = AddPost(post)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
