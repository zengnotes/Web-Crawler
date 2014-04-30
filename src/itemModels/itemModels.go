package itemModels

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
	"strconv"
	//"os"
)

type DatabaseStore interface {
	save()
	prepare()
	findByUrl(sql.DB,string) (*items)
}

var pInsert string = "insert into item (`name`,`brand`,`costPrice`,`discountPercent`,`finalPrice`,`discountValue`,`status`,`domain`,`url`,`featureImage`,`createDate`,`createBy`,`lastModifiedBy`,`siteRef`,`totalQty`,`lastCrawlDate`,`capacity`,`postID`) values "+
"(?,?,?,?,?,?,?,?,?,?,now(),0,0,?,?,now(),?,?)"

var pUpdate string = "update item set `name`=?,`brand`=?,`costPrice`=?,`discountPercent`=?,`finalPrice`=?,`discountValue`=?,`status`=?,`domain`=?,`url`=?,`featureImage`=?,`siteRef`=?,`totalQty`=?,`lastCrawlDate`=now(), `capacity`=?, `postID`=? where id=?"

var pUpdatePostID string = "update item set `postID`=? where id=?"

var qtyInsert string = "insert into ItemQtyBought (`qty`,`crawlDate`,`lastModifiedBy`,`createBy`,`createDate`,`itemID`) values "+
		"(?,now(),0,0,now(),?)"

var categoryInsert string = "insert into label (`label`) values (?)"

var itemTagLabel string = "insert into itemLabel (`itemID`,`labelID`,`createDate`) values (?,?,now())"

var itemRemoveLabel string = "update itemLabel set active = 0 where `itemID` = ?"

type items struct {
	Name,Brand,Domain,Url,FeatureImage,SiteRef string
	CostPrice,DiscountPercent,FinalPrice,DiscountValue float64
	CreateDate, CreateBy, LastModifiedBy, LastCrawlDate,Capacity string
	Status,TotalQty,Qty int
	Id,PostID int64
	db *sql.DB
	pInsert, pUpdate, qtyInsert, categoryInsert,itemTagLabel,itemRemoveLabel *sql.Stmt
	isNewRecord bool
	Category []int64
	OriginalCategory []string
}

func NewItems(db *sql.DB) *items {
	m := new(items)
	m.isNewRecord = true
	m.Id = 0
	m.DiscountPercent=0
	m.Qty=0
	m.CostPrice=0
	m.DiscountValue=0
	m.FinalPrice = 0
	m.db = db
	m.PostID = 0
	//m.prepare()
	return m;
}

func CheckDatabase(db *sql.DB, url string) *items {
	var m *items;
	m = new(items)
	exists := m.findByUrl(db,url)
	if !exists {
		m = NewItems(db)
	}
	return m;
}

func (q *items) GetCategory(data []string) []int64 {
	var returnVal []int64
	var err bool
	var _data string
	err = false
	for _,_data = range data {
		_data = strings.TrimSpace(_data)
		if _data != "" && len(_data) > 0 {
			var foundID int64
			foundID = 0
			row := q.db.QueryRow("select `id` from `label` where `label` like ?",_data).Scan(&foundID)
			switch {
			case row == sql.ErrNoRows:
				err = true
			case row != nil:
				err = true
			}
			if foundID==0 && err {
				result,err := q.db.Exec(categoryInsert,_data)
				if err == nil {
					foundID, err = result.LastInsertId()
					if err != nil {
						fmt.Printf("%s",err.Error())
					}
				} else {
					fmt.Printf("%s",err.Error())
				}
			}
			returnVal = append(returnVal,foundID)
			q.OriginalCategory = append(q.OriginalCategory,_data)
		}

	}
	return returnVal
}

func (q *items) findByUrl(db *sql.DB, url string) (bool) {
	var err bool
	err= false
	q.db = db
	//q.prepare()
	q.isNewRecord = false
	q.Id = 0;
	row := q.db.QueryRow("select `id`,`name`,`brand`,`costPrice`,`discountPercent`,`finalPrice`,`discountValue`,`status`,`domain`,`url`,`featureImage`,`createDate`,`createBy`,`lastModifiedBy`,`siteRef`,`totalQty`,`lastCrawlDate`,`capacity`,`postID` from item where url=?",url).Scan(&q.Id,&q.Name,&q.Brand,&q.CostPrice,&q.DiscountPercent,&q.FinalPrice,&q.DiscountValue,&q.Status,&q.Domain,&q.Url,&q.FeatureImage,&q.CreateDate,&q.CreateBy,&q.LastModifiedBy,&q.SiteRef,&q.TotalQty,&q.LastCrawlDate,&q.Capacity,&q.PostID)
	switch {
	case row == sql.ErrNoRows:
		err = true
	case row != nil:
		err = true
	}
	if err && q.Id==0 {
		return false;
	}else {
		return true;
	}
}

func (q *items) sendToSmokage() {
	value,e := json.Marshal(q)

	if e != nil {
		fmt.Println(q)
		fmt.Println(e.Error())
		//os.Exit(1)
	}
	resp, err := http.PostForm("http://smokage.sg/newPost.php?check=12738612hk12kbqkjnb12o83y192kjbigew12381293",
		url.Values{"data": {string(value)}, "id": {strconv.Itoa(int(q.PostID))}})
	if err == nil {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
		}
		postIDString := string(contents);


		postID,err := strconv.Atoi(postIDString)
		if err == nil {
			q.PostID = int64(postID)
		}
	}
}

func (q *items) Save() (bool) {
	q.sendToSmokage()
	var err error
	var result sql.Result
	if q.isNewRecord {
		result,err = q.db.Exec(pInsert,q.Name,q.Brand,q.CostPrice,q.DiscountPercent,q.FinalPrice,q.DiscountValue,q.Status,q.Domain,q.Url,q.FeatureImage,q.SiteRef,q.TotalQty,q.Capacity,q.PostID) //q.pInsert.Exec(q.Name,q.Brand,q.CostPrice,q.DiscountPercent,q.FinalPrice,q.DiscountValue,q.Status,q.Domain,q.Url,q.FeatureImage,q.SiteRef,q.TotalQty)
		if err == nil {
			q.Id,err = result.LastInsertId()
			if err != nil {
				fmt.Printf("%s",err.Error())
			}
		} else {
			fmt.Printf("%s",err.Error())
		}

	} else if q.Id != 0 {
		q.db.Exec(pUpdate,q.Name,q.Brand,q.CostPrice,q.DiscountPercent,q.FinalPrice,q.DiscountValue,q.Status,q.Domain,q.Url,q.FeatureImage,q.SiteRef,q.TotalQty,q.Capacity,q.PostID,q.Id)
		//q.pUpdate.Exec(q.Name,q.Brand,q.CostPrice,q.DiscountPercent,q.FinalPrice,q.DiscountValue,q.Status,q.Domain,q.Url,q.FeatureImage,q.SiteRef,q.TotalQty,q.Id)
	}
	_,err = q.db.Exec(qtyInsert,q.Qty,q.Id) //q.qtyInsert.Exec(q.Qty,q.Id)



	if err != nil {
		fmt.Printf("%s",err.Error())
	}
	_,err = q.db.Exec(itemRemoveLabel,q.Id)// q.itemRemoveLabel.Exec(q.Id)

	if err != nil {
		fmt.Printf("%s",err.Error())
	}

	for _,cid := range q.Category {
		_,err = q.db.Exec(itemTagLabel,q.Id,cid)// q.itemTagLabel.Exec(q.Id,cid)
		if err != nil {
			fmt.Printf("%s",err.Error())
		}
	}

	return true;
}

func (q *items) prepare() {
	q.pInsert,_ = q.db.Prepare("insert into item (`name`,`brand`,`costPrice`,`discountPercent`,`finalPrice`,`discountValue`,`status`,`domain`,`url`,`featureImage`,`createDate`,`createBy`,`lastModifiedBy`,`siteRef`,`totalQty`,`lastCrawlDate`) values "+
			"(?,?,?,?,?,?,?,?,?,?,now(),0,0,?,?,now())")
	q.pUpdate,_ = q.db.Prepare("update item set `name`=?,`brand`=?,`costPrice`=?,`discountPercent`=?,`finalPrice`=?,`discountValue`=?,`status`=?,`domain`=?,`url`=?,`featureImage`=?,`siteRef`=?,`totalQty`=?,`lastCrawlDate`=now() where id=?")
	q.qtyInsert,_ = q.db.Prepare("insert into ItemQtyBought (`qty`,`crawlDate`,`lastModifiedBy`,`createBy`,`createDate`,`itemID`) values "+
			"(?,now(),0,0,now(),?)")
	q.categoryInsert,_ = q.db.Prepare("insert into label (`label`) values "+
			"(?)")
	q.itemTagLabel,_ = q.db.Prepare("insert into itemLabel (`itemID`,`labelID`,`createDate`) values "+
			"(?,?,now())")
	q.itemRemoveLabel,_ = q.db.Prepare("update itemLabel set active = 0 where `itemID` = ?")
}
