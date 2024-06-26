package cmd

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func sendRequest(url string) ([]string, error) {
	var links []string
	var noHrefLinks = 0

	urlParts := strings.Split(url, "/")
	absUrl := urlParts[0] + "//" + urlParts[2]

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Cannot connect to this url")
		return []string{}, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch the page, status code: %d\n", response.StatusCode)
		return []string{}, nil // Return early if not OK
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Cannot read the content")
		return []string{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			if strings.Contains(link, "http") {
				links = append(links, link+",")
			} else {
				modifiedLink := absUrl + link
				links = append(links, modifiedLink+",")
			}
		} else {
			noHrefLinks++
		}
	})
	if noHrefLinks != 0 {
		fmt.Printf("Found %d no href links\n", noHrefLinks)
	}
	return links, nil

}
func sendConcurrentRequest(urls []string, c chan<- string) {
	var MSG = "END_OF_URLS"

	for _, url := range urls {
		links, err := sendRequest(url)
		if err != nil {
			fmt.Println(err)
		}
		for _, link := range links {
			c <- link
		}
	}
	c <- MSG
}

var rootCmd = &cobra.Command{
	Use:   "linker [url]",
	Short: "scrape all links within a web page",
	Long: `A longer description that spans multiple lines and likely contains
			examples and usage of using your application. For example:
			
			printing in terminal
			linker https://www.digikala.com/
			linker -s https://www.digikala.com/ https://emalls.ir/ https://torob.com/


			writing into a file
			linker url > path_to_file
`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No arguments provided ...")
		}
		url := args[0]
		links, err := sendRequest(url)
		if err != nil {
			panic(err)
		}
		// Print out the links
		for _, i := range links {
			fmt.Printf("%s\n", i)
		}
		multiPages, _ := cmd.Flags().GetBool("multiple")
		if multiPages == true {
			var numUrls = len(args)
			var numLinks int = 0
			c := make(chan string)
			go sendConcurrentRequest(args, c)

			//	Ranging over the channel for read the values
			for link := range c {
				if link == "END_OF_URLS" {
					break
				}
				numLinks++
				fmt.Println(link)
			}
			fmt.Printf("Found %d links within the provided urls", numLinks-numUrls)

		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("multiple", "m", false, "scrape multiple pages concurrently")
}
