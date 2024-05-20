package database

import "log"

type ModulePosition struct {
	Position   string
	ModuleName string
}

func (c *Conn) GetHomeLayouts() *[]ModulePosition {
	rows, err := c.Conn.Query("SELECT * FROM home_layout hl")
	if err != nil {
		log.Fatal(err)
	}

	rows.Scan()

	modulePositions := make([]ModulePosition, 0)

	for rows.Next() {
		res := ModulePosition{}
		err := rows.Scan(&res.Position, &res.ModuleName)
		if err != nil {
			log.Fatal(err)
		}
		modulePositions = append(modulePositions, res)
	}

	return &modulePositions
}

func (c *Conn) SetHomeLayout(position string, moduleName string) error {
	err := c.Conn.QueryRow("INSERT INTO home_layout (position, module_name) VALUES ($1, $2);", position, moduleName).Err()
	return err
}
