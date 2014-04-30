package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/PuerkitoBio/goquery"
	"net/url"
	"./dealcomsg"
	"./bestbuyworldsingapore"
	//"time"
	"strings"
	"strconv"
	"time"
)

//var SQLConnectionString string = "antpolis:Shohoku10_@tcp(antpolis.ctojb1zjwlsy.ap-southeast-1.rds.amazonaws.com:3306)/crawler"
var SQLConnectionString string = "antpolis:Shohoku10_@tcp(apple.antpolis.com:3306)/apdev_dealsCrawler"

func main() {
	var db *sql.DB
	var err error
	db, err = sql.Open("mysql", SQLConnectionString)

	if err != nil {
		fmt.Printf("Connection Error: %s\n", err.Error())
	}
	//db.SetMaxIdleConns(10)
	db.SetMaxIdleConns(150)
	defer db.Close()
	err = db.Ping() // This DOES open a connection if necessary. This makes sure the database is accessible
	fmt.Printf("Start\n")
	crawlAllLinks(dealcomsg.BaseURL, db)
	crawlAllLinks(bestbuyworldsingapore.BaseURL, db)

}
type URLStatuses map[string]bool

var URLQueue = URLStatuses {}
var URLCrawlQueue = URLStatuses {}
var baseURL *url.URL

func removeSessionID(u *url.URL) *url.URL{
	var queryString string
	var newQuery []string
	queryString = u.RawQuery
	queryArray := strings.Split(queryString, "&")
	var found bool
	found = false
	for _,value := range queryArray {

		if check := strings.Contains(value,"zenid="); check {
			found = true
		}
		if check := strings.Contains(value,"language="); check {
			found = true
		}
		if !found {
			newQuery = append(newQuery,value)
		}
	}
	u.RawQuery = strings.Join(newQuery,"&")
	return u
}
var threadCouunter int = 0


func crawlAllLinks (crawlURL string, db *sql.DB) {
	var doc *Document
	var e error
	baseURL,_ = url.Parse(crawlURL)
	var docChan = make (chan int)
	//var docChanTwo = make (chan int)
	var URLLinks = URLStatuses {}

	URLLinks[crawlURL] = true
	URLQueue[crawlURL] = true

	if doc, e = NewDocument(crawlURL); e != nil {
		panic(e.Error())
	}
	var f func(*Document, chan int)
	f = func(n *Document, c chan int){
		n.Find("a").Each(func(i int, s *Selection) {
			linkFound,exists := s.Attr("href")
			if exists {
				linkURLObj,error := url.Parse(linkFound)
				if error == nil {
					resolvedURL := baseURL.ResolveReference(linkURLObj)
					resolvedURL = removeSessionID(resolvedURL)
					if _, crawled := URLLinks[resolvedURL.String()]; !crawled && resolvedURL.Host == baseURL.Host && !ignoredPage(resolvedURL){
						URLLinks[resolvedURL.String()] = true;
						URLQueue[resolvedURL.String()] = true
						URLCrawlQueue[resolvedURL.String()] = true
						fmt.Println("Crawling: "+resolvedURL.String()+" ["+strconv.Itoa(len(URLQueue))+"]")
						newUrl := resolvedURL.String()//"http://www.deal.com.sg/shop/product/illuminating-finish-powder-compact-foundation-spf-12-5-honey";
						if doc, e = NewDocument(newUrl); e == nil {
							go f(doc,c)
						}
					}
				}
			}
		})
	}
	go f(doc,docChan)
	doneChan := make(chan bool)
	var timeout func()
	timeout = func() {
		if len(URLQueue)<=0 {
			doneChan <- true
		} else {
			time.Sleep(time.Minute * 2)
			timeout();
		}
	}
	time.Sleep(time.Second * 5)
	go timeout();



	var maxThread int
	maxThread = 50

	var ripPageTimer func(chan bool)
	var docR *Document
	var eR error
	var ripChan = make(chan bool)

	ripPageTimer = func(c chan bool) {
		if len(URLQueue)>0 {
			for ripUrl,_ := range URLQueue {
				if maxThread>=threadCouunter {
					if docR, eR = NewDocument(ripUrl); eR == nil {

						threadCouunter++;
						go ripPages(docR,c)
					} else {
						delete(URLQueue,docR.Url.String())
					}
				} else {
					break;
				}
			}
		}
		time.Sleep(time.Millisecond * 200)
		go ripPageTimer(c);
	}

	go ripPageTimer(ripChan);
	select {
		case msg := <-docChan: fmt.Printf("out of loop %s\n",msg)

	case done := <- doneChan:
		if done {
			return;
		}
	}

}

func ripPages(n *Document, c chan bool) {
	fmt.Println("Ripping: "+n.Url.String()+" ["+strconv.Itoa(len(URLQueue))+"]")
	var db *sql.DB
	var err error
	db, err = sql.Open("mysql", SQLConnectionString)

	if err != nil {
		fmt.Printf("Connection Error: %s\n", err.Error())
	}
	//db.SetMaxIdleConns(10)
	db.SetMaxIdleConns(150)
	defer db.Close()
	err = db.Ping() // This DOES open a connection if necessary. This makes sure the database is accessible
	delete(URLQueue,n.Url.String())
	switch n.Url.Host {
		case dealcomsg.GetBaseURLObject().Host: dealcomsg.StripPage(n,db)
			break;
		case bestbuyworldsingapore.GetBaseURLObject().Host: bestbuyworldsingapore.BestBuy.StripPage(n,db)
			break;
	}
	threadCouunter--
}

func ignoredPage(n *url.URL) bool{
	var returnBool bool
	switch n.Host {
		case dealcomsg.GetBaseURLObject().Host: returnBool = dealcomsg.IgnoreUrl(n)
			break;
		case bestbuyworldsingapore.GetBaseURLObject().Host: returnBool = bestbuyworldsingapore.BestBuy.IgnoreUrl(n)
			break;

	}
	return returnBool
}
