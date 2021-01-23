// XML DOM processing for Golang, supports xpath query
package xmldom

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"strings"
)

func Must(doc *Document, err error) *Document {
	if err != nil {
		panic(err)
	}
	return doc
}

func ParseXML(s string) (*Document, error) {
	return Parse(strings.NewReader(s))
}

func ParseFile(filename string) (*Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(file)
}

func Parse(r io.Reader) (*Document, error) {
	p := xml.NewDecoder(r)
	t, err := p.Token()
	if err != nil {
		return nil, err
	}

	doc := new(Document)
	var e *Node
	for t != nil {
		switch token := t.(type) {
		case xml.StartElement:
			// a new node
			el := new(Node)
			el.Document = doc
			el.Parent = e
			if doc.Root != nil {
				el.NS = doc.getNamespaceByURI(token.Name.Space)
			}
			el.Name = token.Name.Local

			for _, attr := range token.Attr {
				if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
					doc.NamespaceList = append(doc.NamespaceList, &Namespace{
						Name: "",
						URI:  attr.Value,
					})
				} else if attr.Name.Space == "xmlns" {
					doc.NamespaceList = append(doc.NamespaceList, &Namespace{
						Name: attr.Name.Local,
						URI:  attr.Value,
					})
				} else {
					el.Attributes = append(el.Attributes, &Attribute{
						NS:    doc.getNamespaceByURI(attr.Name.Space),
						Name:  attr.Name.Local,
						Value: attr.Value,
					})
				}
			}
			if e != nil {
				e.Children = append(e.Children, el)
			}
			e = el

			if doc.Root == nil {
				doc.Root = e
			}
		case xml.EndElement:
			e = e.Parent
		case xml.CharData:
			// text node
			if e != nil {
				e.Text = string(token)
			}
		case xml.ProcInst:
			doc.ProcInst = stringifyProcInst(&token)
		case xml.Directive:
			doc.Directives = append(doc.Directives, stringifyDirective(&token))
		}

		// get the next token
		t, err = p.Token()
	}

	// Make sure that reading stopped on EOF
	if err != io.EOF {
		return nil, err
	}

	// All is good, return the document
	return doc, nil
}
