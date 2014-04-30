package bestbuyworldsingapore
import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/PuerkitoBio/goquery"
	"strings"
	"strconv"
	"../itemModels"
	"net/url"
)
type striperStruct struct {

}
type Stripper interface {
	StripPage( *Document,  *sql.DB, chan string)
	IgnoreUrl(*Document)
}

var BaseURL string = "http://sg.bestbuy-world.com"

var BestBuy striperStruct

func GetBaseURLObject() *url.URL{
	u,_ := url.Parse(BaseURL)
	return u
}

func (stripper *striperStruct)StripPage(n *Document, db *sql.DB) {
	if n.Find("body#productinfoBody").Size() > 0 {
		stripProduct(n,db)
	}
}

func (stripper *striperStruct)IgnoreUrl(n *url.URL) (bool) {
	if n.Scheme == "https" {
		return true;
	}
	if strings.Contains(n.String(),"http://sg.bestbuy-world.com/index.php?main_page=time_out") {
		return true;
	}
	return false;
}

func stripProduct(n *Document, db *sql.DB) {
	msg := n.Url.String()
	item := itemModels.CheckDatabase(db,msg)

	var searchCategory []string
	if n.Find("head meta").Size() >0 {
		n.Find("head meta").Each(func (i int, s *Selection){
			name,exists := s.Attr("name")
			if exists {
				if name == "keywords" {
					value,exists := s.Attr("content")
					if exists {
						searchCategory = strings.Split(value,",")
					}
				}
			}
		});
	}
	item.Category = item.GetCategory(searchCategory)


	fmt.Printf("Strip Product: %s\n",msg)
	contentBlock := n.Find("#productInfoBox")
	var cleanValue string
	if contentBlock.Find("h1#productName").Size() > 0 {
		text,can := contentBlock.Find("h1#productName").Html()
		if can == nil {
			myStringArray := strings.Split(text, "<br/>")
			//if _, crawled := myStringArray[0]; !crawled {
			item.Name = myStringArray[0]
			//}
			if len(myStringArray) > 1 {
				item.Capacity = myStringArray[1]
			}
		}
	}
	if contentBlock.Find("#productPrices .productSpecialPrice").Size() >0 {
		cleanValue = contentBlock.Find("#productPrices .productSpecialPrice").First().Text()
		item.FinalPrice,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(strings.Replace(cleanValue,"S","",-1),",","",-1),"$","",-1)),64)
	}
	if(contentBlock.Find("#productPrices .normalprice").Size() >0) {
		cleanValue = contentBlock.Find("#productPrices .normalprice").First().Text()
		item.CostPrice,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(strings.Replace(cleanValue,"S","",-1),",","",-1),"$","",-1)),64)
	}
	cleanValue = ""
	if(contentBlock.Find("#product-title-desc h1").Size() >0) {
		cleanValue = contentBlock.Find("#product-title-desc h1").First().Text()
		item.Brand = cleanValue
	}
	if item.CostPrice > item.FinalPrice {
		discountPercent := (item.CostPrice-item.FinalPrice)/item.CostPrice
		item.DiscountPercent = discountPercent
		item.DiscountValue = item.CostPrice-item.FinalPrice
	}
	item.Qty = 0
	newU := removeSessionID(n.Url)
	item.Url = newU.String()
	item.Domain = n.Url.Host
	item.Save()
}
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
