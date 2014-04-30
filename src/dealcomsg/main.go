package dealcomsg

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
var BaseURL string = "http://www.deal.com.sg"

func GetBaseURLObject() *url.URL{
	u,_ := url.Parse(BaseURL)
	return u
}

func IgnoreUrl(n *url.URL) (bool) {
	if n.Scheme == "https" {
		return true;
	}
	if strings.Contains(n.String(),"http://www.deal.com.sg/deal_join_popup") {
		return true;
	}
	return false;
}

func stripProduct(n *Document, db *sql.DB) {
	msg := n.Url.String()
	bodyClass,exists := n.Find("#content-left-area .meta").Attr("class")
	item := itemModels.CheckDatabase(db,msg)
	var searchCategory []string
	if exists {
		classArray := strings.Split(bodyClass," ")
		var s string
		for _,s = range classArray {
			s = strings.ToLower(strings.TrimSpace(s))
			if strings.Contains(s,"category-") {
				searchCategory = append(searchCategory, strings.Replace(s,"category-","",-1))
			}
		}

	}
	if n.Find("#content-left-area .meta .links.inline li").Size() >0 {
		n.Find("#content-left-area .meta .links.inline li").Each(func (i int, s *Selection){
			searchCategory = append(searchCategory,s.Find("a").Text())
		});
	}
	item.Category = item.GetCategory(searchCategory)


	fmt.Printf("Strip Product: %s\n",msg)
	contentBlock := n.Find("#content-left-area")
	var cleanValue string
	item.Name = strings.TrimSpace(contentBlock.Find("h1.title").First().Text())
	if contentBlock.Find("#product-details .product-info.product.sell .deal-theme-list-price").Size() >0 {
		cleanValue = contentBlock.Find("#product-details .product-info.product.sell .deal-theme-list-price").First().Text()
		item.FinalPrice,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(cleanValue,",","",-1),"$","",-1)),64)
	}
	if(contentBlock.Find("#product-top-info .main-product-image a").Size() >0) {
		imageLink := contentBlock.Find("#product-top-info .main-product-image a").First();
		_imageLink,exist := imageLink.Attr("href")
		if exist {
			item.FeatureImage = _imageLink
		}
	}
	if(contentBlock.Find("#product-details .product-info.product.list .deal-theme-list-price").Size() >0) {
		cleanValue = contentBlock.Find("#product-details .product-info.product.list .deal-theme-list-price").First().Text()
		item.CostPrice,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(cleanValue,",","",-1),"$","",-1)),64)
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

	item.Url = n.Url.String()
	item.Domain = n.Url.Host
	item.Save()
}

func StripPage(n *Document, db *sql.DB) {
	if n.Find("#content-left-area #today-deal-block").Size() > 0 {
		stripDeals(n,db)
	} else if n.Find("#content-left-area .meta").Size() > 0 && n.Find("#content-left-area #product-group").Size() > 0 {
		stripProduct(n,db)
	}
}

func stripDeals(n *Document, db *sql.DB) {
	msg := n.Url.String()
	bodyClass,exists := n.Find("body").Attr("class")

	item := itemModels.CheckDatabase(db,msg)
	if exists {
		classArray := strings.Split(bodyClass," ")
		var searchCategory []string
		var s string
		for _,s = range classArray {
			s = strings.ToLower(strings.TrimSpace(s))
			if strings.Contains(s,"category-") {
				searchCategory = append(searchCategory, strings.Replace(s,"category-","",-1))
			}
		}
		item.Category = item.GetCategory(searchCategory)
	}
	fmt.Printf("Strip Deals: %s\n",msg)
	contentBlock := n.Find("#content-left-area")
	var cleanValue string
	item.Name = strings.TrimSpace(contentBlock.Find("h1.title").First().Text())

	if(contentBlock.Find("#today-deal-content .img-wrapper img").Size() >0) {
		imageLink := contentBlock.Find("#today-deal-content .img-wrapper img").First();
		_imageLink,exist := imageLink.Attr("src")
		if exist {
			item.FeatureImage = _imageLink
		}
	}

	if contentBlock.Find("#deal-price-sell .currency-number").Size() >0 {
		cleanValue = contentBlock.Find("#deal-price-sell .currency-number").First().Text()
		item.FinalPrice,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(cleanValue,",","",-1),"$","",-1)),64)
	}
	if(contentBlock.Find("#deal-price-orig .currency-number").Size() >0) {
		cleanValue = contentBlock.Find("#deal-price-orig .currency-number").First().Text()
		item.CostPrice,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(cleanValue,",","",-1),"$","",-1)),64)
	}
	if(contentBlock.Find("#deal-price-disc .currency-number").Size() >0) {
		cleanValue = contentBlock.Find("#deal-price-disc .currency-number").First().Text()
		item.DiscountValue,_ = strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(cleanValue,",","",-1),"$","",-1)),64)
	}
	if item.CostPrice > item.FinalPrice {
		discountPercent := (item.CostPrice-item.FinalPrice)/item.CostPrice
		item.DiscountPercent = discountPercent
	}
	if contentBlock.Find("#deal-sold #slider-current-display .message").Size() >0 {
		cleanValue = strings.TrimSpace(strings.Replace(contentBlock.Find("#deal-sold #slider-current-display .message").First().Text(),"Already ","",-1))
		cleanValue = strings.TrimSpace(strings.Replace(cleanValue," bought!","",-1))
		item.Qty,_ = strconv.Atoi(cleanValue)
	}

	item.Url = n.Url.String()
	item.Domain = n.Url.Host
	item.Save()
}
