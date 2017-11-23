package function

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Handle a serverless request
func Handle(req []byte) []byte {

	pushEvent := PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	clonePath, err := clone(pushEvent)
	if err != nil {
		log.Println("Clone ", err.Error())
		os.Exit(-1)
	}

	stack, err := parseYAML(pushEvent, clonePath)
	if err != nil {
		log.Println("parseYAML ", err.Error())
		os.Exit(-1)
	}

	var shrinkWrapPath string
	shrinkWrapPath, err = shrinkwrap(pushEvent, clonePath)
	if err != nil {
		log.Println("Shrinkwrap ", err.Error())
		os.Exit(-1)
	}

	var tars []tarEntry
	tars, err = makeTar(pushEvent, shrinkWrapPath, stack)
	if err != nil {
		log.Println("Error creating tar(s): ", err.Error())
		os.Exit(-1)
	}

	err = deploy(tars, pushEvent.Repository.Owner.Login, pushEvent.Repository.Name)
	if err != nil {
		log.Println("Error deploying tar(s): ", err.Error())
		os.Exit(-1)
	}

	return []byte(fmt.Sprintf("Deployed tar from: %s", tars))
}
