package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (a Arguments) ItemList() ([]byte, error) {
	f, err := os.OpenFile(a["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func fUnmarshal(f *os.File) (users []User, err error) {
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(content) != 0 {
		err := json.Unmarshal(content, &users)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func fFlush(f *os.File, users []User) (data []byte, err error) {
	content, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}
	err = f.Truncate(0)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	_, err = f.WriteString(string(content))
	return content, err
}

func (a Arguments) AddItem() ([]byte, error) {
	var user User
	err := json.Unmarshal([]byte(a["item"]), &user)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(a["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	users, err := fUnmarshal(f)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.Id == user.Id {
			return []byte(fmt.Sprintf("Item with id %v already exists", user.Id)), nil
		}
	}
	users = append(users, user)
	return fFlush(f, users)
}

func (a Arguments) RemoveUser() ([]byte, error) {
	f, err := os.OpenFile(a["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	oldusers, err := fUnmarshal(f)
	if err != nil {
		return nil, err
	}
	newusers := []User{}
	for _, u := range oldusers {
		if u.Id != a["id"] {
			newusers = append(newusers, u)
		}
	}
	if len(newusers) == len(oldusers) {
		return []byte(fmt.Errorf("Item with id %v not found", a["id"]).Error()), nil
	}
	return fFlush(f, newusers)
}

func (a Arguments) FindById() ([]byte, error) {
	f, err := os.OpenFile(a["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	users, err := fUnmarshal(f)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.Id == a["id"] {
			return json.Marshal(u)
		}
	}
	return []byte(""), nil
}

func Perform(args Arguments, writer io.Writer) error {

	var content []byte
	var err error

	if args["fileName"] == "" {
		return errors.New("-fileName flag has to be specified")
	}

	switch args["operation"] {
	case "":
		return errors.New("-operation flag has to be specified")
	case "list":
		content, err = args.ItemList()
		if err != nil {
			return err
		}
		writer.Write(content)
		return nil
	case "add":
		if args["item"] == "" {
			return errors.New("-item flag has to be specified")
		}
		content, err := args.AddItem()
		if err != nil {
			return err
		}
		writer.Write(content)
		return nil
	case "remove":
		if args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		content, err = args.RemoveUser()
		if err != nil {
			return err
		}
		writer.Write(content)
		return nil
	case "findById":
		if args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		content, err = args.FindById()
		if err != nil {
			return err
		}
		writer.Write(content)
		return nil
	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}
}

func parseArgs() Arguments {
	id := flag.String("id", "", "user id")
	item := flag.String("item", "", "user data")
	operation := flag.String("operation", "", "requested action")
	fileName := flag.String("fileName", "", "output file name")
	flag.Parse()
	return Arguments{"id": *id, "item": *item, "operation": *operation, "fileName": *fileName}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
