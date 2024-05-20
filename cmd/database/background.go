package database

import (
	"io"
	"log"
	"os"
	"path"
	"strings"

	// apt install libvips-dev
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
)

type Background struct {
	Id         int
	Filename   string
	Created_at string
	Selected   bool
}

func (c *Conn) GetBackgrounds() *[]Background {
	rows, err := c.Conn.Query("SELECT * FROM background bg ORDER BY bg.selected DESC")
	if err != nil {
		log.Fatal(err)
	}

	rows.Scan()

	backgrounds := make([]Background, 0)

	for rows.Next() {
		res := Background{}
		err := rows.Scan(&res.Id, &res.Filename, &res.Created_at, &res.Selected)
		if err != nil {
			log.Fatal(err)
		}
		backgrounds = append(backgrounds, res)
	}

	return &backgrounds
}

func (c *Conn) GetSelectedBackground() (Background, error) {
	res := Background{}
	err := c.Conn.QueryRow("SELECT * FROM background bg WHERE selected").Scan(
		&res.Id, &res.Filename, &res.Created_at, &res.Selected)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (c *Conn) UploadBackground(ctx echo.Context) (Background, error) {
	file, err := ctx.FormFile("file")
	if err != nil {
		return Background{}, err
	}
	src, err := file.Open()
	if err != nil {
		return Background{}, err
	}
	defer src.Close()

	imgPath := path.Join("./images/background", file.Filename)
	dst, err := os.Create(imgPath)
	if err != nil {
		return Background{}, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return Background{}, err
	}

	// Convert image to webp
	buffer, err := bimg.Read(imgPath)
	if err != nil {
		return Background{}, err
	}

	newImage, err := bimg.NewImage(buffer).Convert(bimg.WEBP)
	if err != nil {
		return Background{}, err
	}

	if bimg.NewImage(newImage).Type() != "webp" {
		return Background{}, err
	}
	newPath := strings.Replace(imgPath, path.Ext(imgPath), ".webp", 1)
	bimg.Write(newPath, newImage)

	err = os.Remove(imgPath)
	if err != nil {
		return Background{}, err
	}

	req := c.Conn.QueryRow(
		"INSERT INTO background (filename) VALUES ($1) RETURNING *;", path.Base(newPath),
	)
	res := Background{}
	err = req.Scan(&res.Id, &res.Filename, &res.Created_at, &res.Selected)
	if err != nil {
		return Background{}, err
	}

	return res, nil
}

func (c *Conn) DeleteBackground(id int) error {
	var filename string
	err := c.Conn.QueryRow("DELETE FROM background WHERE id = $1 RETURNING filename;", id).Scan(&filename)
	if err != nil {
		return err
	}

	err = os.Remove(path.Join("./images/background", filename))
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) SetSelectedBackground(id int) (Background, error) {
	res := Background{}
	err := c.Conn.QueryRow("UPDATE background SET selected='f' WHERE id != $1;", id).Err()
	if err != nil {
		return res, err
	}
	err = c.Conn.QueryRow("UPDATE background SET selected='t' WHERE id = $1 RETURNING *;", id).Scan(
		&res.Id, &res.Filename, &res.Created_at, &res.Selected)
	if err != nil {
		return res, err
	}

	return res, nil
}
