/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package model

import (
	"encoding/xml"
)

type Context struct {
	XMLName      xml.Name     `xml:"Context"`
	ResourceLink ResourceLink `xml:"ResourceLink"`
	Path         string       `xml:"path,attr"`
	DocBase      string       `xml:"docBase,attr"`
}

type ResourceLink struct {
	XMLName xml.Name `xml:"ResourceLink"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Global  string   `xml:"global,attr"`
}

type DbConfig struct {
	DbResource DbResource `yaml:"db_resource"`
	DbContext  DbResource `yaml:"db_context"`
}
type DbResource struct {
	Local string `yaml:"local"`
	Dev   string `yaml:"dev"`
	Sit   string `yaml:"sit"`
}
