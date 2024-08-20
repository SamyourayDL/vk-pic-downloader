package main

import(
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api" //vk api
	"os" //file and dir creating
	"io" //io.Copy(file, response.Body)
	"net/http"
	"strconv" //Itoa() - int to string
	"regexp"
)

//config 
var fileName int = 0
var postsCount int = 100
var offset int = 300
var token string = "vk1.a.LykzmnvTBc9MxXPyTTbblPYZ9RSnNLY6DtHlGrhvVdUK0pXB9H_S-rr6-z3IdMhy0UDbjjZut9Z8zbucA19nHj6UQYDWMYb7ucVY80uzeHVWKi0fc9bnFbvrTc8dWlFSkHnOXp9sAaXZyROCqxDh7peUT2kof2PTyueIfYYodvM7ruEb1MpI90kgCy70mhDYBwCeSMUSzBydcFfFUOLrig" 

func main() {
	//Change working dir to /Sciamano240
	cwd, _ := os.Getwd() // /home/mazino/go/vk_scrap
	err := os.Chdir(cwd+"/Sciamano240")

	//Make request to get #mircocabbia posts
	vk := api.NewVK(token)
	resp, err := vk.WallSearch(api.Params{
		"owner_id": -123754724,
		"query": "#mircocabbia",
		"owners_only": true,
		"count": postsCount,
		"offset": offset })
	if err != nil {
		fmt.Println("Error while making request")
	}
	
	//Compile a regexp to find first #
	r, _ := regexp.Compile("#([a-z,_,@,1-9]+)")
	//Create a map[dirName]Counter
	Names := make(map[string]int)

	for _, post := range resp.Items {	//Iterating over posts
		dirName := r.FindString(post.Text)
		_, ok := Names[dirName]
		if !ok {
			Names[dirName] = 0
			os.Mkdir(dirName, 0700)
		} 

		for _, attach := range post.Attachments {	//Iterating over Attachments
			var url string
			var biggestSize = 0
			for _, pic := range attach.Photo.Sizes {	//Iterating over different sizes
				size := int(pic.Height) + int(pic.Width)
				if size > biggestSize {
					url = pic.URL
					biggestSize = size
				}
			}

			Download(url, dirName, Names)
		}
	}
}

//map always passing by value, it won't copy the content
func Download(url string, dirName string, Names map[string]int) {
	response, e := http.Get(url)
	if e != nil {
		fmt.Println("Downloading err")
	}
	defer response.Body.Close()

	cwd, _ := os.Getwd()
	err := os.Chdir(cwd+"/"+dirName)

	name := strconv.Itoa(Names[dirName])+".png"
	file, err := os.Create(name)
	Names[dirName]++

	if err != nil {
		fmt.Println("Failed to create a file")
	}
	defer file.Close()
	
	_, err = io.Copy(file, response.Body)
	if err != nil {
		fmt.Println("Failed to Copy")
	}

	err = os.Chdir(cwd)
}