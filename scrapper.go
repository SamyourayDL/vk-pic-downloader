package main

import(
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api" //vk api
	"os" //file and dir creating
	"io" //io.Copy(file, response.Body)
	"net/http"
	"strconv" //Itoa() - int to string
	"regexp"
	"time"
)

//config 
var fileName int = 0
var postsCount int = 100
var offset int = 0
var token string = "vk1.a.22ckzC3IFBQcFVImlbdrRd5y7l4H9W-oOkT-s_pRgm-cpTrkYvYsZJ1qnnN9o69pTqCffcK8mIhtu08Hr6hI_Gyo3CV6uVds3uz4pukrTT7u3UgTDqRx46to6oU8fIGZdrPnrg-jIELBG6dV6qJJQx7SKU9aUbnFYpfTp-aI8Pefla2i_CS9ZBDQ4y6sUaAvOOdQ6MtdBoECcJy2oe9YBQ" 

func main() {
	start := time.Now()
	//Change working dir to /Sciamano240
	cwd, _ := os.Getwd() // /home/mazino/go/vk_scrap
	err := os.Chdir(cwd+"/seq_test")

	//Make request to get #mircocabbia posts
	vk := api.NewVK(token)
	resp, err := vk.WallSearch(api.Params {
		"owner_id": -123754724,
		"query": "#mircocabbia",
		"owners_only": true,
		"count": postsCount,
		"offset": offset })
	if err != nil {
		fmt.Println("Error while making request")
	}
	
	//Compile a regexp to find first #
	r, _ := regexp.Compile("#[\\w|@]+") // #([a-z,_,@,1-9]+)"
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
	timeElapsed := time.Since(start)
	fmt.Println(timeElapsed)
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