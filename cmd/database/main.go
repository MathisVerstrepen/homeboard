package database

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type Conn struct {
	Conn *sql.DB
}

func Connect() *Conn {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return &Conn{
		Conn: db,
	}
}

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

	req := c.Conn.QueryRow(
		"INSERT INTO background (filename) VALUES ($1) RETURNING *;", file.Filename,
	)
	res := Background{}
	err = req.Scan(&res.Id, &res.Filename, &res.Created_at, &res.Selected)
	if err != nil {
		return Background{}, err
	}
	fmt.Println(res)

	dst, err := os.Create(path.Join("./images/background", file.Filename))
	if err != nil {
		return Background{}, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
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
