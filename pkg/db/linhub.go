package database

import (
	"fmt"

	"diikstra.fr/homeboard/pkg/classification"
)

type LinkHubLink struct {
	Id         int
	Name       string
	Url        string
	Icon       string
	Is_nsfw    bool
	Is_starred bool
	Created_at string
}

type LinkHubImage struct {
	Linkhub_id int
	Ext        string
	Image_id   string
	Is_nsfw    bool
}

type LinkhubTag struct {
	Id      int
	Tag     string
	Is_nsfw bool
}

func (c *Conn) GetLinkhubLink(id int) (LinkHubLink, error) {
	res := LinkHubLink{}
	err := c.Conn.QueryRow(`
		SELECT * FROM linkhub_link ll
		WHERE ll.id = $1
	`, id).Scan(&res.Id, &res.Name, &res.Url, &res.Icon, &res.Is_nsfw, &res.Is_starred, &res.Created_at)

	return res, err
}

func (c *Conn) GetLinkhubLinks() (*[]LinkHubLink, error) {
	rows, err := c.Conn.Query(`SELECT * FROM linkhub_link`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Scan()

	linkHubLinks := make([]LinkHubLink, 0)
	for rows.Next() {
		res := LinkHubLink{}
		err := rows.Scan(&res.Id, &res.Name, &res.Url, &res.Icon, &res.Is_nsfw, &res.Is_starred, &res.Created_at)
		if err != nil {
			return nil, err
		}
		linkHubLinks = append(linkHubLinks, res)
	}

	return &linkHubLinks, nil
}

func (c *Conn) SetLinkhubLink(link LinkHubLink) (LinkHubLink, error) {
	req := c.Conn.QueryRow(`
        INSERT INTO linkhub_link (name, url, icon, is_nsfw)
        VALUES ($1, $2, $3, $4) RETURNING *;
    `, link.Name, link.Url, link.Icon, link.Is_nsfw)

	res := LinkHubLink{}
	err := req.Scan(&res.Id, &res.Name, &res.Url, &res.Icon, &res.Is_nsfw, &res.Is_starred, &res.Created_at)

	return res, err
}

func (c *Conn) GetLinkhubImages(linkhub_id int) (*[]LinkHubImage, error) {
	rows, err := c.Conn.Query(`SELECT * FROM linkhub_image li
								WHERE li.linkhub_id = $1`, linkhub_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Scan()

	linkHubImages := make([]LinkHubImage, 0)
	for rows.Next() {
		res := LinkHubImage{}
		err := rows.Scan(&res.Linkhub_id, &res.Ext, &res.Image_id, &res.Is_nsfw)
		if err != nil {
			return nil, err
		}
		linkHubImages = append(linkHubImages, res)
	}

	return &linkHubImages, nil
}

func (c *Conn) SetLinkhubImage(image LinkHubImage) error {
	_, err := c.Conn.Exec(`
        INSERT INTO linkhub_image (linkhub_id, ext, image_id, is_nsfw)
        VALUES ($1, $2, $3, $4)
    `, image.Linkhub_id, image.Ext, image.Image_id, image.Is_nsfw)

	return err
}

func (c *Conn) SetImageTags(image_id string, tags []classification.Classification) error {
	var tagIds []int

	for _, tag := range tags {
		var tagId int
		err := c.Conn.QueryRow(`
            INSERT INTO linkhub_tag (tag, is_nsfw)
            VALUES ($1, $2)
            ON CONFLICT (tag) DO UPDATE SET tag = EXCLUDED.tag
            RETURNING id
        `, tag.Label, false).Scan(&tagId)

		if err != nil {
			return fmt.Errorf("erreur lors de l'insertion du tag: %v", err)
		}
		tagIds = append(tagIds, tagId)
	}

	for _, tagId := range tagIds {
		_, err := c.Conn.Exec(`
            INSERT INTO linkhub_image_tag (image_id, tag_id)
            VALUES ($1, $2)
            ON CONFLICT DO NOTHING
        `, image_id, tagId)

		if err != nil {
			return fmt.Errorf("erreur lors de l'insertion du lien image-tag: %v", err)
		}
	}

	return nil
}

func (c *Conn) GetImageDetails(image_id string) (LinkHubImage, []LinkhubTag, error) {
	imageMeta := LinkHubImage{}
	err := c.Conn.QueryRow(`
		SELECT * FROM linkhub_image li
		WHERE li.image_id = $1
	`, image_id).Scan(&imageMeta.Linkhub_id, &imageMeta.Ext, &imageMeta.Image_id, &imageMeta.Is_nsfw)

	if err != nil {
		return LinkHubImage{}, nil, err
	}

	rows, err := c.Conn.Query(`SELECT lt.id, lt.tag, lt.is_nsfw FROM linkhub_image_tag lit
								LEFT JOIN linkhub_tag lt ON lit.tag_id = lt.id
								WHERE lit.image_id = $1`, image_id)
	if err != nil {
		return LinkHubImage{}, nil, err
	}
	defer rows.Close()

	rows.Scan()

	imageTags := make([]LinkhubTag, 0)
	for rows.Next() {
		res := LinkhubTag{}
		err := rows.Scan(&res.Id, &res.Tag, &res.Is_nsfw)
		if err != nil {
			return LinkHubImage{}, nil, err
		}
		imageTags = append(imageTags, res)
	}

	return imageMeta, imageTags, nil
}
