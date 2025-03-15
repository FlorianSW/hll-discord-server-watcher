package tcadmin

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
)

const (
	hllGameId = "1098726659"
	hllModId  = "0"
	hllFileId = "1"

	configUrlTemplate = "https://%s/Aspx/Interface/GameHosting/MvcConfigEditor.aspx?gameid=%s&modid=%s&fileid=%s&serviceid=%s"
	loginUrlTemplate  = "https://%s/Aspx/Interface/Base/Login.aspx"
)

type client struct {
	hc http.Client

	baseUrl string
	creds   Credentials
}

func NewClient(hc http.Client, baseUrl string, creds Credentials) *client {
	return &client{
		hc:      hc,
		baseUrl: baseUrl,
		creds:   creds,
	}
}

func (c *client) isLoggedIn() (bool, error) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf(loginUrlTemplate, c.baseUrl), nil)
	if err != nil {
		return false, err
	}
	res, err := c.hc.Do(r)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusFound, nil
}

func (c *client) login() error {
	if loggedIn, err := c.isLoggedIn(); err != nil {
		return err
	} else if loggedIn {
		return nil
	}

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf(loginUrlTemplate, c.baseUrl), nil)
	if err != nil {
		return err
	}
	r.SetBasicAuth(c.creds.Username, c.creds.Password)
	res, err := c.hc.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusFound {
		b, _ := io.ReadAll(res.Body)
		println(string(b))
		return errors.New("invalid username or password")
	}
	return nil
}

func (c *client) ServerInfo(serviceId string) (*ServerInfo, error) {
	if err := c.login(); err != nil {
		return nil, err
	}

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf(configUrlTemplate, c.baseUrl, hllGameId, hllModId, hllFileId, serviceId), nil)
	if err != nil {
		return nil, err
	}
	r.SetBasicAuth(c.creds.Username, c.creds.Password)
	res, err := c.hc.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code, expected 200, got %d", res.StatusCode)
	}
	h, err := html.Parse(res.Body)
	if err != nil {
		return nil, err
	}
	return &ServerInfo{
		Name:     valueOf(h, "Server Name"),
		Password: valueOf(h, "Server Password"),
	}, nil
}

func valueOf(h *html.Node, label string) string {
	n := findNode(h, func(n *html.Node) bool {
		for _, a := range n.Attr {
			if n.FirstChild == nil {
				continue
			}
			if a.Key == "class" && a.Val == "Label" && strings.Contains(n.FirstChild.Data, label) {
				return true
			}
		}
		return false
	})
	if n == nil {
		return ""
	}
	var pwNode *html.Node
	for _, a := range n.Parent.Attr {
		if a.Key == "for" {
			pwNode = findNode(n.Parent.Parent.Parent, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "input"
			})
		}
	}
	if pwNode == nil {
		return ""
	}
	for _, a := range pwNode.Attr {
		if a.Key == "value" {
			return a.Val
		}
	}
	return ""
}

func findNode(n *html.Node, selector func(n *html.Node) bool) *html.Node {
	if selector(n) {
		return n
	}
	for node := range n.ChildNodes() {
		if f := findNode(node, selector); f != nil {
			return f
		}
	}
	return nil
}
