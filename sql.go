// Package dblog for creating, reading, and writing to a sqlite database for my blog
package dblog

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3" // blank import so the sql engine can do its thing in init
)

var (
	pathToDB = "./posts.sqlite3"
)

// openDB just opens the connection to the db from the pathToDB package var
func openDB() (db *sql.DB) {
	db, err := sql.Open("sqlite3", pathToDB)
	if err != nil {
		log.Fatal(err.Error())
	}
	return db
}

// closeDB is the mirror of open. Generally just defer it immediately after
// openDB
func closeDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// MakeDB makes the basic schema for the database if it doesn't exist.
// originally, it had a trigger for update_date whenever something changed, but
// that was too clunky to deal with for me. I might eventually add it searching
// for a backup .sql file in the same directory as the db in case there's
// backup data.
func MakeDB() (err error) {
	db := openDB()
	defer closeDB(db)

	// post table
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS post (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL UNIQUE,
  file_name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL,
  pub_date TEXT NOT NULL CHECK(pub_date LIKE '____-__-__'),
  update_date TEXT NOT NULL CHECK(update_date LIKE '____-__-__'),
  thumbnail TEXT
  )`); err != nil {
		return err
	}

	// tag table
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS tag (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  category STRING NOT NULL DEFAULT 'content', -- for medium, content, and lang
  description TEXT
  )`); err != nil {
		return err
	}

	// associative identity
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS post_tag (
  post_id INTEGER,
  tag_id INTEGER,
  PRIMARY KEY (post_id, tag_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
  )`); err != nil {
		return err
	}

	return nil
}

// AggregatePosts gets all the posts and tags metadata from the db into a slice
// of posts sorted in reverse chron order. if you give it a tag, it'll only
// return posts with that tag. giving the func a negative number returns an
// empty slice and giving it zero will return all the entries that match. If
// you give it an integer less than the total amount of posts available with
// the filter, it'll return that many posts back in the slice, still in reverse
// chron order.
func AggregatePosts(postQty int, filterTag string) (posts []Post, err error) {
	if postQty < 0 {
		return []Post{}, nil
	}

	db := openDB()
	defer closeDB(db)

	var query string
	var filters []interface{}
	if filterTag == "" {
		query = `SELECT title, file_name, posts.description, pub_date, update_date
  FROM posts
  ORDER BY pub_date DESC`
	} else {
		query = `SELECT title, file_name, posts.description, pub_date, update_date
  FROM posts JOIN posts_tags
  ON posts.id = posts_tags.post_id JOIN tags
  ON posts_tags.tag_id = tags.id
  WHERE tags.name = ?
  ORDER BY pub_date DESC`
		filters = append(filters, filterTag)
	}

	if postQty > 0 {
		query = query + `
  LIMIT ?`
		filters = append(filters, postQty)
	}

	rows, err := db.Query(query, filters...)
	defer rows.Close()

	for rows.Next() {
		post := Post{}
		err := rows.Scan(&post.Title, &post.FileName, &post.Description, &post.PubDate, &post.UpdateDate)
		if err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// FetchPost brings back the struct data of a single post including a tag slice
// of all matching tags to post in associative identity.
func FetchPost(fileName string) (post Post, err error) {
	if !DoesPostExist(fileName) {
		return Post{}, errors.New(fileName + " doesn't exist")
	}

	db := openDB()
	defer closeDB(db)

	var id int
	var thumbnailJSON sql.NullString
	err = db.QueryRow(`SELECT id, title, file_name, description, pub_date, update_date, thumbnail
  FROM posts
  WHERE file_name = ?`, fileName).Scan(&id, &post.Title, &post.FileName, &post.Description, &post.PubDate, &post.UpdateDate, &thumbnailJSON)
	if err != nil {
		return Post{}, err
	}

	tagRows, err := db.Query(`SELECT tags.name
  FROM tags JOIN posts_tags
  ON tags.id = posts_tags.tag_id
  WHERE posts_tags.post_id = ?
  ORDER BY name`, id)
	if err != nil {
		return Post{}, err
	}
	defer tagRows.Close()

	var tags []Tag
	for tagRows.Next() {
		var name string
		err := tagRows.Scan(&name)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		tags = append(tags, Tag{Name: name})
	}
	post.Tags = tags

	// optional stuff
	if thumbnailJSON.Valid && len(thumbnailJSON.String) > 0 {
		// e.g. { "src" : "pic.jpeg", "alt" : "cool pic", "title" : "what you see if you hover"}
		var thumbnail Img
		err := json.Unmarshal([]byte(thumbnailJSON.String), &thumbnail)
		if err != nil {
			log.Println(err.Error())
			return Post{}, err
		}
		post.Thumbnail = thumbnail
	}

	return post, nil
}

// FetchThumbnail is for a very niche thing I needed for my home page that
// displays my latest photos post with a valid thumbnail. Sqlite doesn't have
// structs, and I don't wanna bother with pgsql or mariadb for something this
// small, so I just marshal and unmarshal json into the db as text. This func
// fetches just that post and gives back not only the img struct but the rest
// of the post as well for easily being able to link back to the post.
func FetchThumbnail() (post Post, err error) {
	db := openDB()
	defer closeDB(db)

	var thumbnailJSON sql.NullString
	err = db.QueryRow(`SELECT title, file_name, posts.description, pub_date, update_date, thumbnail
  FROM posts JOIN posts_tags
  ON posts.id = posts_tags.post_id JOIN tags
  ON posts_tags.tag_id = tags.id
  WHERE tags.name = 'photos'
  AND posts.thumbnail IS NOT NULL
  AND posts.thumbnail <> ''
  ORDER BY posts.pub_date DESC
  LIMIT 1`).Scan(&post.Title, &post.FileName, &post.Description, &post.PubDate, &post.UpdateDate, &thumbnailJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return Post{}, errors.New("no valid thumbnails exist")
		}
		return Post{}, err
	}

	err = json.Unmarshal([]byte(thumbnailJSON.String), &post.Thumbnail)
	if err != nil {
		return Post{}, err
	}
	return post, nil
}

// FetchTag is basically identical to FetchPost but way smaller and less
// complicated
func FetchTag(tagName string) (tag Tag, err error) {
	db := openDB()
	defer closeDB(db)

	err = db.QueryRow(`SELECT name, description, category
  FROM tags
  WHERE name = ?`, tagName).Scan(&tag.Name, &tag.Description, &tag.Category)
	if err != nil {
		return Tag{}, err
	}

	return tag, nil
}

// DoesPostExist is an evolution from my original filesystem func that I still
// use called doesFileExist that ensures a certain post in actually in the db.
func DoesPostExist(fileName string) bool {
	db := openDB()
	defer closeDB(db)

	var count int
	err := db.QueryRow(`SELECT COUNT(*)
  FROM posts
  WHERE file_name = ?`, fileName).Scan(&count)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return count > 0
}

// DoesTagExist is the exact same deal as DoesPostExist
func DoesTagExist(tag string) bool {
	db := openDB()
	defer closeDB(db)

	var count int
	err := db.QueryRow(`SELECT COUNT(*)
  FROM tags
  WHERE name = ?`, tag).Scan(&count)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return count > 0
}

// checkTag is idential but smaller
func checkTag(t Tag) error {
	if t.Name == "" ||
		t.Category == "" ||
		t.Description == "" {
		return errors.New("unfilled required attributes")
	}
	return nil
}

// AddTag is just a simpler version of AddPost. The only filtering done before
// is seeing if your tag struct has all the required attributes to add to the
// db.
func AddTag(tag Tag) (err error) {
	if err = checkTag(tag); err != nil {
		return err
	}

	db := openDB()
	defer closeDB(db)

	_, err = db.Exec(`INSERT INTO tags (name, category, description)
  VALUES
  (?,  ?,  ?)
`, tag.Name, tag.Category, tag.Description)
	if err != nil {
		return err
	}

	return nil
}

// deletePost deletes whatever filename for a post you feed it, if it exists.
func deletePost(fileName string) (err error) {
	if !DoesPostExist(fileName) {
		return errors.New(fileName + " doesn't exist")
	}

	db := openDB()
	defer closeDB(db)

	_, err = db.Exec(`DELETE FROM posts
  WHERE file_name = ?`, fileName)
	if err != nil {
		return err
	}

	return nil
}

// deleteTag deletes whatever tag name you give it, if it exists.
func deleteTag(tagName string) (err error) {
	if !DoesTagExist(tagName) {
		return errors.New(tagName + " doesn't exist")
	}

	db := openDB()
	defer closeDB(db)

	_, err = db.Exec(`DELETE FROM tags
  WHERE name = ?`, tagName)
	if err != nil {
		return err
	}

	return nil
}
