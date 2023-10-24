package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"reflect"
	"strings"
)

type DisplayInfo struct {
	ID_display   int
	Diagonal     float32
	Resolution   string
	Type_display string
	GSync        bool
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "user=postgres password=Lax212212 dbname=Monitors sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createTable(db *sql.DB, model interface{}) error {
	table := reflect.TypeOf(model)

	if table.Kind() != reflect.Struct {
		return fmt.Errorf("Model должен быть структурой")
	}

	tableName := table.Name()

	_, err := db.Exec(fmt.Sprintf("Create table if not exists %s (ID_display serial not null primary key, Diagonal real not null, Resolution varchar not null, Type_display varchar not null, Gsync boolean not null)", tableName))

	return err
}

func insertInto(db *sql.DB, model interface{}) error {
	table := reflect.TypeOf(model)

	if table.Kind() != reflect.Struct {
		return fmt.Errorf("Model должен быть структурой")
	}

	tableName := table.Name()

	fields := []interface{}{}
	columnNames := []string{}
	placeholders := []string{}

	for i := 0; i < table.NumField(); i++ {
		field := table.Field(i)
		fieldName := field.Name
		columnNames = append(columnNames, fieldName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		fieldValue := reflect.ValueOf(model).Field(i).Interface()

		if fieldName == "GSync" {
			fieldValue = fmt.Sprintf("%t", fieldValue)
		}

		fields = append(fields, fieldValue)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columnNames, ", "), strings.Join(placeholders, ", "))
	_, err := db.Exec(query, fields...)

	return err
}

func ReadTable(db *sql.DB, model interface{}, id int) error {
	table := reflect.TypeOf(model)

	if table.Kind() != reflect.Ptr || table.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("Model должен быть указателем на структуру")
	}

	tableName := table.Elem().Name()

	query := fmt.Sprintf("SELECT * FROM %s WHERE ID_display = $1", tableName)
	row := db.QueryRow(query, id)

	values := make([]interface{}, table.Elem().NumField())

	for i := 0; i < table.Elem().NumField(); i++ {
		values[i] = reflect.New(table.Elem().Field(i).Type).Interface()
	}

	if err := row.Scan(values...); err != nil {
		return err
	}

	result := reflect.ValueOf(model).Elem()

	for i := 0; i < table.Elem().NumField(); i++ {
		result.Field(i).Set(reflect.ValueOf(values[i]).Elem())
	}

	fmt.Printf("Читаю запись из базы данных: %+v\n", model)
	return nil
}

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	display := DisplayInfo{
		Diagonal:     27.0,
		Resolution:   "2560x1440",
		Type_display: "IPS",
		GSync:        true,
	}

	if err := createTable(db, display); err != nil {
		log.Fatal(err)
	}

	if err := insertInto(db, display); err != nil {
		log.Fatal(err)
	}

	if err := ReadTable(db, &display, 1); err != nil {
		log.Fatal(err)
	}

}
