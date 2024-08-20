package main

import(
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api" //vk api
	"github.com/SevereCloud/vksdk/v3/object" // for type []object.WallWallpost
	"os" //file and dir creating
	"io" //io.Copy(file, response.Body)
	"net/http"
	"strconv" //Itoa() - int to string
	"regexp"
	"sync"
	"path/filepath"
	"time"
)

//config 
const PostsCount = 100
const Offset = 0
const WorkerCount = 10
const Token = "vk1.a.22ckzC3IFBQcFVImlbdrRd5y7l4H9W-oOkT-s_pRgm-cpTrkYvYsZJ1qnnN9o69pTqCffcK8mIhtu08Hr6hI_Gyo3CV6uVds3uz4pukrTT7u3UgTDqRx46to6oU8fIGZdrPnrg-jIELBG6dV6qJJQx7SKU9aUbnFYpfTp-aI8Pefla2i_CS9ZBDQ4y6sUaAvOOdQ6MtdBoECcJy2oe9YBQ" 
const WorkingDir = "conc_test"
var cwd string

type RWMap struct {
	table map[string]int
	m sync.RWMutex
}

func main() {
	start := time.Now()

	os.Mkdir(WorkingDir, 0700)
	cwd, _ := os.Getwd()
	cwd = filepath.Join(cwd, WorkingDir)
	err := os.Chdir(cwd)


	//Make request to get #mircocabbia posts
	vk := api.NewVK(Token)
	resp, err := vk.WallSearch(api.Params{
		"owner_id": -123754724,
		"query": "#mircocabbia",
		"owners_only": true,
		"count": PostsCount,
		"offset": Offset })
	if err != nil {
		fmt.Println("Error while making request")
	}
	
	//Not sync.Map because we write a lot 
	Names := RWMap{ table : make(map[string]int) }

	//make channel with tickets for Workers and initialize it
	tickets := make(chan int, PostsCount)
	for i := 0; i < PostsCount; i++ {
		tickets <- i
	}
	close(tickets)	//no more writing in this chanel 
	var wg sync.WaitGroup	//wait all goroutines to end before main goroutine end

	for i := 0; i < WorkerCount; i++ {
		wg.Add(1)
		go parsePost(resp.Items, &Names, tickets, &wg)	//important &wg - else deadlock 
	}

	wg.Wait()
	timeElapsed := time.Since(start)
	fmt.Println(timeElapsed)
}

func parsePost(Items []object.WallWallpost,  Names *RWMap, tickets chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	//Compile a regexp to find first #
	r, _ := regexp.Compile("#[\\w|@]+") // #([a-z,_,@,1-9]+)"

	for i := range tickets {	// Items[i] = post
		dirName := r.FindString(Items[i].Text)

		Names.m.RLock()
		_, ok := Names.table[dirName]
		Names.m.RUnlock()

		if !ok {
			Names.m.Lock()
			Names.table[dirName] = 0
			Names.m.Unlock()
			os.Mkdir(dirName, 0700)
		}
		
		for _, attach := range Items[i].Attachments {	//Iterating over Attachments
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
	fmt.Println("Worker end job!")
}

//map always passing by value, it won't copy the content
func Download(url string, dirName string, Names *RWMap) {
	response, e := http.Get(url)
	if e != nil {
		fmt.Println("Downloading err")
	}
	defer response.Body.Close()

	Names.m.Lock()
	fileNumber := Names.table[dirName]
	Names.table[dirName]++
	Names.m.Unlock()

	name := strconv.Itoa(fileNumber)+".png"
	file, err := os.Create(filepath.Join(cwd, dirName, name))

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