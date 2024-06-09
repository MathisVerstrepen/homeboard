package database

import (
	"fmt"
	"log"
	"strings"
)

func (c *Conn) GetModuleVariables(position string, variables *map[string]string) {
	rows, err := c.Conn.Query(`
		SELECT variable_name, value FROM module_variable mv 
		WHERE position = ($1)
		ORDER BY variable_name
		`, position)
	if err != nil {
		log.Fatal(err)
	}

	rows.Scan()

	for rows.Next() {
		var (
			variable_name  string
			variable_value string
		)
		err := rows.Scan(&variable_name, &variable_value)
		if err != nil {
			log.Fatal(err)
		}
		(*variables)[variable_name] = variable_value
	}
}

func (c *Conn) SetModuleVariables(position string, variables *map[string]string) error {
	var insert []string

	for variable_name, variable_value := range *variables {
		insert = append(insert, fmt.Sprintf(`
			INSERT INTO module_variable (position, variable_name, value) VALUES ('%s', '%s', '%s')
			ON CONFLICT (position, variable_name)
			DO UPDATE SET value = EXCLUDED.value;
			`, position, variable_name, variable_value,
		))
	}

	err := c.Conn.QueryRow(strings.Join(insert, "; ")).Err()
	return err
}
