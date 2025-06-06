package entity

import "encoding/xml"

type Version struct {
	Hash    string
	PubDate string
}

type YmlCatalog struct {
	XMLName xml.Name `xml:"yml_catalog"`
	Date    string   `xml:"date,attr"`
	Shop    Shop     `xml:"shop"`
	Name    string   `xml:"name"`
	Company string   `xml:"company"`
}

type Shop struct {
	Categories Categories `xml:"categories"`
	Offers     Offers     `xml:"offers"`
}

type Categories struct {
	Category []Category `xml:"category"`
}

type Category struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:",chardata"`
}

type Offers struct {
	Offer []Offer `xml:"offer"`
}

type Offer struct {
	XMLName          xml.Name `xml:"offer"`
	ID               int      `xml:"id,attr"`
	Vendor           string   `xml:"vendor"`
	Price            int      `xml:"price"`
	CurrencyID       string   `xml:"currencyId"`
	CategoryID       int      `xml:"categoryId"`
	Picture          string   `xml:"picture"`
	URL              string   `xml:"url"`
	Name             string   `xml:"name"`
	Description      string   `xml:"description"`
	ShortDescription string   `xml:"shortDescription"`
}
