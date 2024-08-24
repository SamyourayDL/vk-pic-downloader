package main

import(
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
	"log"
)

// Problem: rewriting dir and files in new iteration
// Solution: - More accurate check that dir or 0.png already exists
//   (ideas) - Add bd, make new dirs every launch, load everything in one launch

//config 
const PostsCount = 100
const Offset = 500
const WorkerCount = 10
const Token = "vk1.a.TSvxq5IdWoF4RAZhCi8LmpGKlPjg-CWw3QToomMji6qAXoc6KoxOQZx8I8TyhxmaKBkWpNxWzak2-hTJt_Qf1Ll7glmWsWhynfY-AuduOzjVZ-JLPjHDZrVFvzV38LzB0H_dtSbA0svntFz1R2J0mmU7ljQt-WfeV34PcljC7BXKZhB_cLOa7dMkmpnrjDUtzIEqvfhU8alnm2VFHUtROw" 
const WorkingDir = "Sciamano240 2"
var cwd string

type RWMap struct {
	table map[string]int
	m sync.RWMutex
}

func main() {
	start := time.Now()
	
	//err - local, will delete after if 
	if _, err := os.Stat(WorkingDir); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.Mkdir(WorkingDir, 0700)
	}
	
	cwd, _ := os.Getwd()
	cwd = filepath.Join(cwd, WorkingDir)
	err := os.Chdir(cwd)

	log.SetFlags(log.Ltime)
	log.SetPrefix("mainroutine: ")

	//Make request to get #mircocabbia posts
	vk := api.NewVK(Token)
	resp, err := vk.WallSearch(api.Params{
		"owner_id": -123754724,
		"query": "#mircocabbia",
		"owners_only": true,
		"count": PostsCount,
		"offset": Offset })
	if err != nil {
		log.Println("Error while making request")
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

	log.SetPrefix("goroutine: ")	//new logs will be written from goroutines 
	for i := 0; i < WorkerCount; i++ {
		wg.Add(1)
		go parsePost(resp.Items, &Names, tickets, &wg)	//important &wg - else deadlock 
	}

	wg.Wait()

	log.SetPrefix("")
	timeElapsed := time.Since(start)
	log.Println("Time spent: ", timeElapsed)
}

func parsePost(Items []object.WallWallpost,  Names *RWMap, tickets chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	//Compile a regexp to find first #
	r, _ := regexp.Compile("#[\\w|@]+") // #([a-z,_,@,1-9]+)"

	for i := range tickets {	// Items[i] = post
		dirName := r.FindString(Items[i].Text)

		// It is all one big ifWrite
		Names.m.Lock()
		_, ok := Names.table[dirName]
		if !ok {
			Names.table[dirName] = 0
			os.Mkdir(dirName, 0700)
		}
		Names.m.Unlock()
		
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
			//maybe Attachments was empty and url is empty too 
			if len(url) > 0 {
				Download(url, dirName, Names)
			}
		}
	}
	log.Println("Worker end job!")
}

//map always passing by value, it won't copy the content
func Download(url string, dirName string, Names *RWMap) {
	response, err := http.Get(url)
	if err != nil {
		log.Println("Failed to download picture")
	}
	defer response.Body.Close()

	Names.m.Lock()
	fileNumber := Names.table[dirName]
	Names.table[dirName]++
	Names.m.Unlock()

	name := strconv.Itoa(fileNumber)+".png"
	file, err := os.Create(filepath.Join(cwd, dirName, name))

	if err != nil {
		log.Println("Failed to create a file")
	}
	defer file.Close()
	
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Println("Failed to Copy")
	}

	err = os.Chdir(cwd)
}