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

var rootCmd = &cobra.Command{
	Use:   "linker [url]",
	Short: "scrape all links within a web page",
	Long: `A longer description that spans multiple lines and likely contains
			examples and usage of using your application. For example:.`,

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

}
