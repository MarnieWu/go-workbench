package main

import (
	"flag"
	"go-workbench/internal/httpapi"
	"go-workbench/internal/task"
	"log"
)

func main() {
	fixture := flag.String("fixture", "", "")
	flag.Parse()

	switch *fixture {
	case FixtureSuccess, FixtureEmpty, FixtureError:
		break
	case "":
		log.Fatal("missing local test scenario fixture argument")
	default:
		log.Fatal("unknown local test scenario fixture argument, allowed values: success, empty, error")
	}

	repository := LocalRepository{
		fixture: *fixture,
	}
	service := task.NewService(repository)
	router := httpapi.NewRouter(service, httpapi.LocalOwnerMiddleware(localOwnerId))

	router.Run(":8080")
}
