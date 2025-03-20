package tcadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const (
	configUrlTemplate  = "https://%s/Aspx/Interface/GameHosting/MvcConfigEditor.aspx?gameid=%s&modid=%s&fileid=%s&serviceid=%s"
	loginUrlTemplate   = "https://%s/Aspx/Interface/Base/Login.aspx"
	restartUrlTemplate = "https://%s/Aspx/Interface/Base/CallBacks/ServiceManager.aspx/Restart"

	homeLocation = "/Aspx/Interface/Base/Home.aspx"
)

type client struct {
	hc http.Client

	baseUrl      string
	gameId       string
	modId        string
	configFileId string
	creds        Credentials
}

func NewClient(hc http.Client, baseUrl, gameId, modId, configFileId string, creds Credentials) *client {
	return &client{
		hc:           hc,
		baseUrl:      baseUrl,
		creds:        creds,
		gameId:       gameId,
		modId:        modId,
		configFileId: configFileId,
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
	return res.StatusCode == http.StatusFound && res.Header.Get("Location") == homeLocation, nil
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
	c.hc.Jar, _ = cookiejar.New(nil)

	res, err := c.hc.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusFound {
		return errors.New("invalid username or password")
	}
	return nil
}

func (c *client) ServerInfo(serviceId string) (*ServerInfo, error) {
	if err := c.login(); err != nil {
		return nil, err
	}

	h, err := c.configEditorPage(serviceId)
	if err != nil {
		return nil, err
	}
	return &ServerInfo{
		Name:     valueFor(h, "Server Name"),
		Password: valueFor(h, "Server Password"),
	}, nil
}

func (c *client) configEditorPage(serviceId string) (*html.Node, error) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf(configUrlTemplate, c.baseUrl, c.gameId, c.modId, c.configFileId, serviceId), nil)
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
		return nil, fmt.Errorf("invalid response code, expected 200, got %d with Location %s", res.StatusCode, res.Header.Get("Location"))
	}
	h, err := html.Parse(res.Body)
	return h, err
}

func (c *client) SetServerInfo(serviceId string, name, pw string) error {
	if err := c.login(); err != nil {
		return err
	}

	h, err := c.configEditorPage(serviceId)

	vstate := valueOf(h, "__VSTATE")
	evvalidation := valueOf(h, "__EVENTVALIDATION")
	encoding := valueOf(h, "ctl00$ContentPlaceHolderMain$MvcConfigEditor1$HiddenFieldEncoding")
	rconPw := valueOf(h, "ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$TextBox__DEFAULT_VARIABLE_False73$TextBox1")
	if vstate == nil || evvalidation == nil || encoding == nil || rconPw == nil {
		return errors.New("encountered invalid state")
	}
	formData := url.Values{
		"__EVENTTARGET":            {"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$ButtonSave"},
		"__EVENTARGUMENT":          {""},
		"__VSTATE":                 {*vstate},
		"__VIEWSTATE":              {""},
		"__SCROLLPOSITIONX":        {"0"},
		"__SCROLLPOSITIONY":        {"0"},
		"__VIEWSTATEENCRYPTED":     {""},
		"__EVENTVALIDATION":        {*evvalidation},
		"ctl00$NumericTextBoxItem": {"0.00"},
		"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$HiddenFieldEncoding": {*encoding},
		// Server Name
		"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$TextBox__DEFAULT_VARIABLE_False61$TextBox1": {name},
		// Server Password
		"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$TextBox__DEFAULT_VARIABLE_False82$TextBox1": {pw},
		// RCON Password
		"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$TextBox__DEFAULT_VARIABLE_False73$TextBox1": {*rconPw},
		//Enable GDK
		"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$CheckBox109872665934$CheckBox1": {checked(h, "ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$CheckBox109872665934$CheckBox1")},
		// Enable Steam
		"ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$CheckBox109872665925$CheckBox1": {checked(h, "ctl00$ContentPlaceHolderMain$MvcConfigEditor1$FormViewer1$CheckBox109872665925$CheckBox1")},
	}

	r, err := http.NewRequest(http.MethodPost, fmt.Sprintf(configUrlTemplate, c.baseUrl, c.gameId, c.modId, c.configFileId, serviceId), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	r.SetBasicAuth(c.creds.Username, c.creds.Password)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.hc.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response code, expected 200, got %d with Location %s", res.StatusCode, res.Header.Get("Location"))
	}
	return nil
}

func (c *client) Restart(serviceId string) (string, error) {
	if err := c.login(); err != nil {
		return "", err
	}

	p := map[string]string{
		"serviceId": serviceId,
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	r, err := http.NewRequest(http.MethodPost, fmt.Sprintf(restartUrlTemplate, c.baseUrl), bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", "application/json")
	res, err := c.hc.Do(r)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response code, expected 200, got %d", res.StatusCode)
	}
	var d struct {
		D []string `json:"d"`
	}
	err = json.NewDecoder(res.Body).Decode(&d)
	if err != nil {
		return "", err
	}
	if len(d.D) < 4 {
		return "", nil
	}
	return d.D[3], nil
}

func valueFor(h *html.Node, label string) string {
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

func valueOf(h *html.Node, name string) *string {
	n := findNode(h, func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == "name" && a.Val == name {
				return true
			}
		}
		return false
	})
	if n == nil {
		return nil
	}
	for _, a := range n.Attr {
		if a.Key == "value" {
			return &a.Val
		}
	}
	return nil
}

func checked(h *html.Node, name string) string {
	n := findNode(h, func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == "name" && a.Val == name {
				return true
			}
		}
		return false
	})
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == "checked" {
			return "on"
		}
	}
	return "off"
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
